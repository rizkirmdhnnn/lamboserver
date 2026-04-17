package pgweb

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// osFileSystem is the production implementation of FileSystem using the os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (os.FileInfo, error)              { return os.Stat(name) }
func (osFileSystem) MkdirAll(path string, perm fs.FileMode) error       { return os.MkdirAll(path, perm) }

// OsFileSystem is the exported production FileSystem implementation.
type OsFileSystem = osFileSystem

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
	return system.RunCommand(name, args...)
}

// SystemCommandRunner is the exported production CommandRunner implementation.
type SystemCommandRunner = systemCommandRunner

// Manager handles pgweb binary download, process lifecycle, and status reporting.
type Manager struct {
	paths *system.Paths
	fs    FileSystem
	cmd   CommandRunner
	proc  *exec.Cmd // nil when stopped
	mu    sync.Mutex
}

// Compile-time check: Manager must implement DaemonWebAdminService.
var _ DaemonWebAdminService = (*Manager)(nil)

// NewManager creates a Manager with injected dependencies.
func NewManager(paths *system.Paths, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{
		paths: paths,
		fs:    fs,
		cmd:   cmd,
	}
}

// binaryPath returns the expected path of the pgweb binary.
func (m *Manager) binaryPath() string {
	return filepath.Join(m.paths.PgwebDir(), "pgweb")
}

// IsInstalled reports whether the pgweb binary exists at the expected path.
func (m *Manager) IsInstalled() bool {
	_, err := m.fs.Stat(m.binaryPath())
	return err == nil
}

// Install downloads and extracts the pgweb binary from GitHub releases.
func (m *Manager) Install() error {
	arch := "arm64"
	if system.GetArchitecture() == "amd64" {
		arch = "amd64"
	}

	url := fmt.Sprintf(downloadURLTemplate, pgwebVersion, arch)
	dir := m.paths.PgwebDir()
	zipPath := filepath.Join(dir, "pgweb.zip")

	if err := m.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create pgweb dir: %w", err)
	}

	if _, err := m.cmd.Run("sh", "-c", fmt.Sprintf("curl -sL '%s' -o '%s'", url, zipPath)); err != nil {
		return fmt.Errorf("failed to download pgweb: %w", err)
	}

	if _, err := m.cmd.Run("sh", "-c", fmt.Sprintf("unzip -o '%s' -d '%s'", zipPath, dir)); err != nil {
		return fmt.Errorf("failed to extract pgweb: %w", err)
	}

	// The zip contains pgweb_darwin_{arch}, rename to pgweb.
	extractedName := fmt.Sprintf("pgweb_darwin_%s", arch)
	extractedPath := filepath.Join(dir, extractedName)
	if _, err := m.cmd.Run("mv", extractedPath, m.binaryPath()); err != nil {
		return fmt.Errorf("failed to rename pgweb binary: %w", err)
	}

	if _, err := m.cmd.Run("chmod", "+x", m.binaryPath()); err != nil {
		return fmt.Errorf("failed to make pgweb executable: %w", err)
	}

	// Best-effort cleanup; ignore error.
	m.cmd.Run("rm", "-f", zipPath) //nolint:errcheck

	// Post-install verification.
	if _, err := m.fs.Stat(m.binaryPath()); err != nil {
		return fmt.Errorf("pgweb binary not found at %s after extraction", m.binaryPath())
	}

	return nil
}

// checkPort probes port 8081. Returns an error with a descriptive message if occupied.
func (m *Manager) checkPort() error {
	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		return fmt.Errorf("port 8081 is already in use by another process")
	}
	ln.Close()
	return nil
}

// Start launches the pgweb HTTP daemon on port 8081.
// Uses exec.Command + cmd.Start() so the process runs in the background.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsRunning() {
		return nil // idempotent
	}

	if err := m.checkPort(); err != nil {
		return err
	}

	cmd := exec.Command(m.binaryPath(),
		"--host=127.0.0.1",
		"--user=postgres",
		"--bind=127.0.0.1",
		"--listen=8081",
		"--skip-open",
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pgweb: %w", err)
	}

	m.proc = cmd
	return nil
}

// Stop terminates the running pgweb process. Idempotent — safe to call when not running.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.proc == nil {
		return nil // idempotent
	}

	m.proc.Process.Kill() //nolint:errcheck
	m.proc.Wait()         //nolint:errcheck — required to reap zombie process
	m.proc = nil
	return nil
}

// IsRunning returns true if the pgweb process is alive (signal 0 probe).
func (m *Manager) IsRunning() bool {
	if m.proc == nil || m.proc.Process == nil {
		return false
	}
	err := m.proc.Process.Signal(syscall.Signal(0))
	return err == nil
}

// URL returns the HTTP address where pgweb is served.
func (m *Manager) URL() string {
	return "http://127.0.0.1:8081"
}

// Version returns the pinned pgweb version.
func (m *Manager) Version() string {
	return pgwebVersion
}

// Status returns a snapshot of the current pgweb state.
func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.IsRunning(),
		Port:      defaultPort,
	}
}
