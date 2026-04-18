package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	// Import the MySQL driver for its side-effect of registering the "mysql" driver
	// with database/sql.
	_ "github.com/go-sql-driver/mysql"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// mysqlVersion is the MySQL 8.4 LTS version downloaded from cdn.mysql.com.
// Update this constant when a new patch release should be used.
const mysqlVersion = "8.4.3"

// systemDatabases is the list of MySQL system databases that must never be dropped.
var systemDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
}

// realDBOpener is the production implementation of DBOpener using database/sql.
type realDBOpener struct{}

func (realDBOpener) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return sql.Open(driverName, dataSourceName)
}

// osFileSystem is the production implementation of FileSystem using the os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (os.FileInfo, error)              { return os.Stat(name) }
func (osFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osFileSystem) RemoveAll(path string) error                   { return os.RemoveAll(path) }
func (osFileSystem) ReadDir(name string) ([]os.DirEntry, error)    { return os.ReadDir(name) }

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
	return system.RunCommand(name, args...)
}

// Manager handles MySQL binary discovery, configuration generation, LaunchDaemon
// lifecycle, and database CRUD operations via Unix socket.
type Manager struct {
	paths   *system.Paths
	launchd LaunchdService
	binary  *system.BinaryLocator
	fs      FileSystem
	cmd     CommandRunner
	helper  HelperRunner
	admin   AdminRunner
	dbOpen  DBOpener
}

// NewManager creates a Manager with injected dependencies. paths provides directory
// locations; launchd manages the system service; fs is the filesystem abstraction;
// cmd runs non-privileged commands; helper runs privileged operations.
func NewManager(paths *system.Paths, launchd LaunchdService, fs FileSystem, cmd CommandRunner, helper HelperRunner, admin AdminRunner) *Manager {
	return &Manager{
		paths:   paths,
		launchd: launchd,
		fs:      fs,
		cmd:     cmd,
		helper:  helper,
		admin:   admin,
		binary: &system.BinaryLocator{
			Name:      "mysqld",
			LocalPath: filepath.Join(paths.MySQLBinDir(), "mysqld"),
		},
		dbOpen: realDBOpener{},
	}
}

// NewManagerWithDBOpener creates a Manager with a custom DBOpener — used in tests to
// mock database/sql.Open without requiring a live MySQL socket.
func NewManagerWithDBOpener(paths *system.Paths, launchd LaunchdService, fs FileSystem, cmd CommandRunner, helper HelperRunner, admin AdminRunner, dbOpen DBOpener) *Manager {
	m := NewManager(paths, launchd, fs, cmd, helper, admin)
	m.dbOpen = dbOpen
	return m
}

// --- Service Lifecycle Methods ---

// IsInstalled reports whether the mysqld binary is present at the expected path or on
// system PATH.
func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// Install downloads and extracts MySQL 8.4 LTS from dev.mysql.com into ~/.lamboserver/mysql/.
// The architecture (arm64 or x86_64) is detected automatically.
func (m *Manager) Install() error {
	arch := system.GetArchitecture()
	archStr := "arm64"
	if arch == "amd64" {
		archStr = "x86_64"
	}

	url := fmt.Sprintf(
		"https://cdn.mysql.com/archives/mysql-8.4/mysql-%s-macos14-%s.tar.gz",
		mysqlVersion, archStr,
	)

	if err := m.fs.MkdirAll(m.paths.MySQLBinDir(), 0755); err != nil {
		return fmt.Errorf("failed to create mysql bin dir: %w", err)
	}

	// Download and extract the tarball directly into the MySQL root directory.
	// --strip-components=1 removes the top-level mysql-X.X.X-macos14-{arch}/ prefix.
	mysqlDir := filepath.Clean(m.paths.MySQLDir())
	_, err := m.cmd.Run("sh", "-c",
		fmt.Sprintf("curl -sL '%s' | tar xz --strip-components=1 -C '%s'",
			url, mysqlDir))
	if err != nil {
		return fmt.Errorf("failed to download and extract MySQL: %w", err)
	}

	// The MySQL tarball may contain files with restrictive permissions. Ensure
	// everything is user-writable so we can create my.cnf, data dir, etc. later.
	// This follows the Herd approach: the entire MySQL tree is owned and writable
	// by the current user.
	if _, err := m.cmd.Run("chmod", "-R", "u+rwX", mysqlDir); err != nil {
		return fmt.Errorf("failed to fix MySQL directory permissions: %w", err)
	}

	// Verify the mysqld binary was extracted successfully.
	if _, err := m.fs.Stat(filepath.Join(m.paths.MySQLBinDir(), "mysqld")); err != nil {
		return fmt.Errorf("mysqld binary not found after extraction: %w", err)
	}

	return nil
}

// InitDataDir creates the MySQL data directory, generates my.cnf, and runs
// mysqld --initialize-insecure to create the system tables. Following the Laravel
// Herd approach, MySQL runs as the current user (not _mysql) — this avoids all
// permission issues with data directory ownership on macOS.
func (m *Manager) InitDataDir() error {
	// Create data directory as current user — no privilege elevation needed.
	if err := m.fs.MkdirAll(m.paths.MySQLDataDir(), 0755); err != nil {
		return fmt.Errorf("failed to create MySQL data dir: %w", err)
	}

	// Generate my.cnf configuration file.
	if err := m.writeConf(); err != nil {
		return err
	}

	// Initialize the data directory as current user (no --user flag).
	mysqldBin := filepath.Join(m.paths.MySQLBinDir(), "mysqld")
	_, err := m.cmd.Run(mysqldBin,
		"--initialize-insecure",
		"--datadir="+m.paths.MySQLDataDir(),
		"--basedir="+m.paths.MySQLDir(),
	)
	if err != nil {
		// Auto-clean partial data dir so next retry starts fresh (D-06).
		_ = m.fs.RemoveAll(m.paths.MySQLDataDir())
		return fmt.Errorf("failed to initialize MySQL data directory: %w", err)
	}

	return nil
}

// writeConf generates and writes the MySQL configuration file (my.cnf).
func (m *Manager) writeConf() error {
	conf := fmt.Sprintf(`[mysqld]
datadir     = %s
socket      = %s
pid-file    = %s/mysql.pid
bind-address = 127.0.0.1
port        = 3306
log-error   = %s/mysql-error.log

[client]
socket = %s
`,
		m.paths.MySQLDataDir(),
		m.paths.MySQLSocket(),
		m.paths.MySQLDir(),
		m.paths.LogsDir(),
		m.paths.MySQLSocket(),
	)

	if err := m.fs.WriteFile(m.paths.MySQLConf(), []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write my.cnf: %w", err)
	}
	return nil
}

// Start installs MySQL as a user-level LaunchAgent and starts it. Following the
// Laravel Herd approach, MySQL runs as the current user — no privilege elevation
// needed for start/stop. If the data directory has not been initialized, InitDataDir
// is called first.
func (m *Manager) Start() error {
	if !m.IsInstalled() {
		return fmt.Errorf("MySQL is not installed")
	}

	// Initialise data directory if it has not been set up yet.
	if _, err := m.fs.Stat(filepath.Join(m.paths.MySQLDataDir(), "mysql")); os.IsNotExist(err) {
		if err := m.InitDataDir(); err != nil {
			return fmt.Errorf("failed to initialize MySQL data directory: %w", err)
		}
	}

	// Ensure log files exist before daemon starts.
	ensureLogFile(m.paths.LogsDir() + "/mysql-stdout.log")
	ensureLogFile(m.paths.LogsDir() + "/mysql-stderr.log")

	return m.launchd.Install(system.ServiceConfig{
		Label:      ServiceLabel,
		Program:    m.binary.Find(),
		Args:       []string{"--defaults-file=" + m.paths.MySQLConf()},
		Type:       system.ServiceAgent, // user-level, no admin privileges needed
		RunAtLoad:  true,
		KeepAlive:  true,
		StdoutPath: m.paths.LogsDir() + "/mysql-stdout.log",
		StderrPath: m.paths.LogsDir() + "/mysql-stderr.log",
	})
}

// Stop uninstalls the MySQL LaunchDaemon and waits up to 30 seconds for the Unix socket
// to be removed, ensuring InnoDB completes a clean shutdown before returning (T-07-05).
func (m *Manager) Stop() error {
	if err := m.launchd.Uninstall(system.ServiceConfig{
		Label: ServiceLabel,
		Type:  system.ServiceAgent,
	}); err != nil {
		return err
	}

	// Poll for socket removal to ensure clean InnoDB shutdown (T-07-05).
	for i := 0; i < 30; i++ {
		if _, err := m.fs.Stat(m.paths.MySQLSocket()); os.IsNotExist(err) {
			break
		}
		time.Sleep(time.Second)
	}

	return nil
}

// Status returns a snapshot of the current MySQL service state.
func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.launchd.IsRunning(ServiceLabel),
		Port:      3306,
	}
}

// --- Database CRUD Methods ---

// openDB opens a connection to the local MySQL daemon via Unix socket using the
// root user with no password (standard for local dev; T-07-06 accept disposition).
func (m *Manager) openDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("root@unix(%s)/", m.paths.MySQLSocket())
	db, err := m.dbOpen.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

// ListDatabases returns all non-system databases on the local MySQL instance.
// System databases (information_schema, mysql, performance_schema, sys) are filtered out.
func (m *Manager) ListDatabases() ([]string, error) {
	db, err := m.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		if !systemDatabases[name] {
			databases = append(databases, name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating databases: %w", err)
	}

	return databases, nil
}

// CreateDatabase creates a new database with the given name. The name is validated
// against ^[a-zA-Z0-9_]{1,64}$ before execution to prevent SQL injection (T-07-01).
func (m *Manager) CreateDatabase(name string) error {
	if !dbNameRe.MatchString(name) {
		return fmt.Errorf("invalid database name: %q", name)
	}

	db, err := m.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec("CREATE DATABASE `" + name + "`"); err != nil {
		return fmt.Errorf("failed to create database %q: %w", name, err)
	}
	return nil
}

// DropDatabase drops the database with the given name. The name is validated
// against ^[a-zA-Z0-9_]{1,64}$ and system databases are explicitly rejected (T-07-01, T-07-02).
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

	if _, err := db.Exec("DROP DATABASE `" + name + "`"); err != nil {
		return fmt.Errorf("failed to drop database %q: %w", name, err)
	}
	return nil
}
