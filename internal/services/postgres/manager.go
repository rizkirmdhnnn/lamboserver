package postgres

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	// Import the pgx/v5 driver for its side-effect of registering the "pgx" driver
	// with database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// osFileSystem is the production implementation of FileSystem using the os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (os.FileInfo, error)              { return os.Stat(name) }
func (osFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osFileSystem) RemoveAll(path string) error                   { return os.RemoveAll(path) }
func (osFileSystem) ReadDir(name string) ([]os.DirEntry, error)    { return os.ReadDir(name) }
func (osFileSystem) ReadFile(name string) ([]byte, error)          { return os.ReadFile(name) }

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
	return system.RunCommand(name, args...)
}

// realDBOpener is the production implementation of DBOpener using database/sql.
type realDBOpener struct{}

func (realDBOpener) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return sql.Open(driverName, dataSourceName)
}

// Manager handles PostgreSQL binary discovery, configuration generation, pg_ctl
// lifecycle management, and database connections via Unix socket.
type Manager struct {
	paths  *system.Paths
	binary *system.BinaryLocator
	fs     FileSystem
	cmd    CommandRunner
	helper HelperRunner
	admin  AdminRunner
	dbOpen DBOpener
}

// NewManager creates a Manager with injected dependencies. paths provides directory
// locations; fs is the filesystem abstraction; cmd runs non-privileged commands.
func NewManager(paths *system.Paths, fs FileSystem, cmd CommandRunner, helper HelperRunner, admin AdminRunner) *Manager {
	return &Manager{
		paths:  paths,
		fs:     fs,
		cmd:    cmd,
		helper: helper,
		admin:  admin,
		binary: &system.BinaryLocator{
			Name:      "pg_ctl",
			LocalPath: filepath.Join(paths.PostgreSQLBinDir(), "pg_ctl"),
		},
		dbOpen: realDBOpener{},
	}
}

// NewManagerWithDBOpener creates a Manager with a custom DBOpener — used in tests to
// mock database/sql.Open without requiring a live PostgreSQL socket.
func NewManagerWithDBOpener(paths *system.Paths, fs FileSystem, cmd CommandRunner, helper HelperRunner, admin AdminRunner, dbOpen DBOpener) *Manager {
	m := NewManager(paths, fs, cmd, helper, admin)
	m.dbOpen = dbOpen
	return m
}

// --- Service Lifecycle Methods ---

// IsInstalled reports whether the pg_ctl binary is present at the expected path or on
// system PATH.
func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// Install downloads and extracts PostgreSQL 17 from theseus-rs/postgresql-binaries into
// ~/.lamboserver/postgresql/. Architecture (aarch64 or x86_64) is detected automatically.
// NOTE: PostgreSQL tarball uses "aarch64" (not "arm64" like MySQL) for Apple Silicon.
func (m *Manager) Install() error {
	arch := system.GetArchitecture()
	archStr := "aarch64" // PostgreSQL: aarch64 (NOT arm64 like MySQL)
	if arch == "amd64" {
		archStr = "x86_64"
	}

	url := fmt.Sprintf(
		"https://github.com/theseus-rs/postgresql-binaries/releases/download/%s/postgresql-%s-%s-apple-darwin.tar.gz",
		postgresqlVersion, postgresqlVersion, archStr,
	)

	// Ensure PostgreSQL root dir exists before extraction.
	if err := m.fs.MkdirAll(m.paths.PostgreSQLDir(), 0755); err != nil {
		return fmt.Errorf("failed to create postgresql dir: %w", err)
	}

	// Extract directly into PostgreSQL root dir; --strip-components=1 removes
	// the top-level "postgresql-17.9.0-aarch64-apple-darwin/" prefix.
	postgresqlDir := filepath.Clean(m.paths.PostgreSQLDir())
	_, err := m.cmd.Run("sh", "-c",
		fmt.Sprintf("curl -sL '%s' | tar xz --strip-components=1 -C '%s'",
			url, postgresqlDir))
	if err != nil {
		return fmt.Errorf("failed to download and extract PostgreSQL: %w", err)
	}

	// Verify pg_ctl binary was extracted successfully.
	if _, err := m.fs.Stat(filepath.Join(m.paths.PostgreSQLBinDir(), "pg_ctl")); err != nil {
		return fmt.Errorf("pg_ctl binary not found after extraction: %w", err)
	}

	return nil
}

// InitDataDir runs initdb to create the PostgreSQL data directory (0700, enforced by
// initdb), then overwrites postgresql.conf and pg_hba.conf with LamboServer settings.
// If the data directory is already initialized (PG_VERSION present), returns early.
func (m *Manager) InitDataDir() error {
	// Guard: if already initialized, skip (re-running initdb on non-empty dir fails).
	if _, err := m.fs.Stat(filepath.Join(m.paths.PostgreSQLDataDir(), "PG_VERSION")); err == nil {
		return nil
	}

	// NOTE: Do NOT pre-create data dir with MkdirAll — initdb requires an empty or
	// absent directory and creates it at 0700 itself.
	initdb := filepath.Join(m.paths.PostgreSQLBinDir(), "initdb")
	_, err := m.cmd.Run(initdb, "-D", m.paths.PostgreSQLDataDir(), "--username=postgres")
	if err != nil {
		// Auto-clean partial data dir so next retry starts fresh (D-09).
		// initdb creates the data dir itself, so cleanup removes what it partially created.
		_ = m.fs.RemoveAll(m.paths.PostgreSQLDataDir())
		return fmt.Errorf("failed to initialize PostgreSQL data directory: %w", err)
	}

	// Overwrite configs after initdb (initdb would overwrite if done before).
	if err := m.writeConf(); err != nil {
		return err
	}
	if err := m.writeHba(); err != nil {
		return err
	}
	return nil
}

// writeConf generates and writes the minimal postgresql.conf (D-11).
func (m *Manager) writeConf() error {
	conf := fmt.Sprintf(`# LamboServer PostgreSQL configuration
listen_addresses = 'localhost'
port = 5432
unix_socket_directories = '%s'
logging_collector = on
log_directory = 'log'
`,
		m.paths.PostgreSQLSocketDir(),
	)
	if err := m.fs.WriteFile(m.paths.PostgreSQLConfFile(), []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write postgresql.conf: %w", err)
	}
	return nil
}

// writeHba generates and writes pg_hba.conf with trust auth for local dev (D-12).
func (m *Manager) writeHba() error {
	hba := `# LamboServer pg_hba.conf — trust all local connections for dev
# TYPE  DATABASE  USER  ADDRESS       METHOD
local   all       all                 trust
host    all       all   127.0.0.1/32  trust
host    all       all   ::1/128       trust
`
	if err := m.fs.WriteFile(m.paths.PostgreSQLHbaFile(), []byte(hba), 0644); err != nil {
		return fmt.Errorf("failed to write pg_hba.conf: %w", err)
	}
	return nil
}

// Start performs preflight checks, recovers stale PIDs, and starts PostgreSQL via
// pg_ctl. Returns an error if PostgreSQL is not installed, port 5432 is occupied,
// or pg_ctl fails.
func (m *Manager) Start() error {
	if !m.IsInstalled() {
		return fmt.Errorf("PostgreSQL is not installed")
	}

	// D-09: Port preflight — reject immediately if 5432 is in use.
	if err := m.checkPort(); err != nil {
		return err
	}

	// D-08: Remove stale postmaster.pid if PID is dead.
	if err := m.removeStalePostmasterPID(); err != nil {
		return err
	}

	// Ensure socket dir exists (PostgreSQL writes .s.PGSQL.5432 here).
	if err := m.fs.MkdirAll(m.paths.PostgreSQLSocketDir(), 0755); err != nil {
		return fmt.Errorf("failed to create socket dir: %w", err)
	}

	// Ensure log dir exists (pg_ctl -l requires parent dir; initdb creates data/ not data/log/).
	logDir := filepath.Join(m.paths.PostgreSQLDataDir(), "log")
	if err := m.fs.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log dir: %w", err)
	}

	pgCtl := filepath.Join(m.paths.PostgreSQLBinDir(), "pg_ctl")
	_, err := m.cmd.Run(pgCtl,
		"-D", m.paths.PostgreSQLDataDir(),
		"-l", m.paths.PostgreSQLLogFile(),
		"start",
	)
	if err != nil {
		return fmt.Errorf("failed to start PostgreSQL: %w", err)
	}
	return nil
}

// Stop stops PostgreSQL via pg_ctl stop -m fast and waits up to 30 seconds for
// postmaster.pid to disappear, ensuring clean shutdown (D-14).
func (m *Manager) Stop() error {
	pgCtl := filepath.Join(m.paths.PostgreSQLBinDir(), "pg_ctl")
	_, err := m.cmd.Run(pgCtl,
		"-D", m.paths.PostgreSQLDataDir(),
		"stop", "-m", "fast",
	)
	if err != nil {
		return fmt.Errorf("failed to stop PostgreSQL: %w", err)
	}

	// Poll for postmaster.pid removal to confirm clean shutdown (mirrors MySQL socket poll).
	pidFile := filepath.Join(m.paths.PostgreSQLDataDir(), "postmaster.pid")
	for i := 0; i < 30; i++ {
		if _, err := m.fs.Stat(pidFile); os.IsNotExist(err) {
			break
		}
		time.Sleep(time.Second)
	}
	return nil
}

// Status returns a snapshot of the current PostgreSQL service state.
func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.IsRunning(),
		Port:      5432,
	}
}

// IsRunning returns true if postmaster.pid exists and the PID it contains is alive.
func (m *Manager) IsRunning() bool {
	pidFile := filepath.Join(m.paths.PostgreSQLDataDir(), "postmaster.pid")
	data, err := m.fs.ReadFile(pidFile)
	if err != nil {
		return false
	}
	lines := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return false
	}
	err = syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

// checkPort probes port 5432. Returns an error with a descriptive message if occupied.
func (m *Manager) checkPort() error {
	ln, err := net.Listen("tcp", "127.0.0.1:5432")
	if err != nil {
		return fmt.Errorf("port 5432 is already in use by another process")
	}
	ln.Close()
	return nil
}

// --- Database CRUD Methods ---

// openDB opens a connection to the local PostgreSQL server via Unix socket.
// The pgx/v5 stdlib driver uses libpq keyword-value DSN format where host= is
// the DIRECTORY containing the socket file (not the socket file itself).
func (m *Manager) openDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=postgres dbname=postgres sslmode=disable",
		m.paths.PostgreSQLSocketDir(),
	)
	db, err := m.dbOpen.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

// ListDatabases returns all non-system databases with their human-readable sizes.
func (m *Manager) ListDatabases() ([]PgDatabase, error) {
	db, err := m.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	const query = `
        SELECT datname, pg_size_pretty(pg_database_size(datname))
        FROM pg_database
        WHERE datistemplate = false
          AND datname NOT IN ('postgres', 'template0', 'template1')
        ORDER BY datname
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []PgDatabase
	for rows.Next() {
		var d PgDatabase
		if err := rows.Scan(&d.Name, &d.Size); err != nil {
			return nil, fmt.Errorf("failed to scan database row: %w", err)
		}
		databases = append(databases, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating databases: %w", err)
	}
	return databases, nil
}

// CreateDatabase creates a new PostgreSQL database with the given name.
// The name is validated against dbNameRe before execution.
func (m *Manager) CreateDatabase(name string) error {
	if !dbNameRe.MatchString(name) {
		return fmt.Errorf("invalid database name: %q", name)
	}

	db, err := m.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE DATABASE "` + name + `"`); err != nil {
		return fmt.Errorf("failed to create database %q: %w", name, err)
	}
	return nil
}

// DropDatabase drops the named PostgreSQL database. System databases are rejected.
func (m *Manager) DropDatabase(name string) error {
	if !dbNameRe.MatchString(name) {
		return fmt.Errorf("invalid database name: %q", name)
	}

	if systemDatabases[name] {
		return fmt.Errorf("cannot drop system database: %q", name)
	}

	db, err := m.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(`DROP DATABASE "` + name + `"`); err != nil {
		return fmt.Errorf("failed to drop database %q: %w", name, err)
	}
	return nil
}

// removeStalePostmasterPID reads postmaster.pid and silently removes it if the PID is
// dead. Uses syscall.Kill(pid, 0) per POSIX convention — ESRCH means no such process.
func (m *Manager) removeStalePostmasterPID() error {
	pidFile := filepath.Join(m.paths.PostgreSQLDataDir(), "postmaster.pid")
	data, err := m.fs.ReadFile(pidFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read postmaster.pid: %w", err)
	}
	lines := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return fmt.Errorf("failed to parse PID from postmaster.pid: %w", err)
	}
	err = syscall.Kill(pid, 0)
	if err == syscall.ESRCH {
		return m.fs.RemoveAll(pidFile)
	}
	if err != nil {
		return fmt.Errorf("failed to check PID %d liveness: %w", pid, err)
	}
	return nil
}
