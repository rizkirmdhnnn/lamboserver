// Package dns manages dnsmasq installation and LaunchDaemon lifecycle to provide
// wildcard DNS resolution for the .test TLD on 127.0.0.1.
package dnsmasq

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// ServiceLabel is the launchd daemon label used for the dnsmasq service.
const ServiceLabel = "com.lamboserver.dnsmasq"

// ServiceStatus reports whether dnsmasq is installed, the daemon is running, and the
// /etc/resolver/{tld} file is present for OS-level DNS delegation.
type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Resolver  bool `json:"resolver"`
}

// osFileSystem is the production implementation of FileSystem using os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (fs.FileInfo, error)                    { return os.Stat(name) }
func (osFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error { return os.WriteFile(name, data, perm) }
func (osFileSystem) ReadFile(name string) ([]byte, error)                      { return os.ReadFile(name) }
func (osFileSystem) MkdirAll(name string, perm fs.FileMode) error              { return os.MkdirAll(name, perm) }

// systemAdminRunner is the production implementation of AdminRunner.
type systemAdminRunner struct{}

func (systemAdminRunner) RunWithPrivileges(command string) error {
	return system.RunWithAdminPrivileges(command)
}

// launchdAdapter wraps *system.LaunchdManager to satisfy LaunchdService interface.
type launchdAdapter struct {
	m *system.LaunchdManager
}

func (a *launchdAdapter) Install(cfg system.ServiceConfig) error   { return a.m.Install(cfg) }
func (a *launchdAdapter) Uninstall(cfg system.ServiceConfig) error { return a.m.Uninstall(cfg) }
func (a *launchdAdapter) IsRunning(label string) bool              { return a.m.IsRunning(label) }

// Manager handles dnsmasq binary discovery, configuration generation, resolver setup,
// and LaunchDaemon lifecycle. dnsmasq runs as a root daemon so it can bind to port 53.
type Manager struct {
	paths   *system.Paths
	launchd LaunchdService
	fs      FileSystem
	admin   AdminRunner
	binary  *system.BinaryLocator
}

// NewManager creates a Manager with injected dependencies.
func NewManager(paths *system.Paths, launchd LaunchdService, fs FileSystem, admin AdminRunner) *Manager {
	return &Manager{
		paths:   paths,
		launchd: launchd,
		fs:      fs,
		admin:   admin,
		binary: &system.BinaryLocator{
			Name:      "dnsmasq",
			LocalPath: paths.DnsmasqBin(),
		},
	}
}

// newManagerWithDeps creates a Manager with injected dependencies (used in tests).
func newManagerWithDeps(paths *system.Paths, launchd LaunchdService, fs FileSystem, admin AdminRunner) *Manager {
	return &Manager{
		paths:   paths,
		launchd: launchd,
		fs:      fs,
		admin:   admin,
		binary: &system.BinaryLocator{
			Name:      "dnsmasq",
			LocalPath: paths.DnsmasqBin(),
		},
	}
}

// IsInstalled reports whether the dnsmasq binary is present at the expected path or on system PATH.
func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// IsDaemonInstalled checks if the LaunchDaemon plist is already installed.
func (m *Manager) IsDaemonInstalled() bool {
	_, err := m.fs.Stat("/Library/LaunchDaemons/" + ServiceLabel + ".plist")
	return err == nil
}

// EnsureConfig writes the dnsmasq configuration file, routing all .test TLD queries
// to 127.0.0.1 and binding to localhost port 53.
func (m *Manager) EnsureConfig() error {
	conf := fmt.Sprintf("address=/.%s/127.0.0.1\nlisten-address=127.0.0.1\nport=53\n", system.TestTLD)
	return m.fs.WriteFile(m.paths.DnsmasqConf(), []byte(conf), 0644)
}

// EnsureResolver creates /etc/resolver/{tld} so macOS delegates .test DNS queries to
// 127.0.0.1. The file is written via AdminRunner since /etc/resolver/ requires root.
// If the resolver file already contains the correct content, no action is taken.
func (m *Manager) EnsureResolver() error {
	content := "nameserver 127.0.0.1\n"
	resolverFile := m.paths.ResolverFile()

	if existing, err := m.fs.ReadFile(resolverFile); err == nil && string(existing) == content {
		return nil
	}

	tmpFile := os.TempDir() + "/lambo-resolver"
	if err := m.fs.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp resolver file: %w", err)
	}

	cmd := fmt.Sprintf("mkdir -p '%s' && cp '%s' '%s'",
		filepath.Clean(m.paths.ResolverDir()),
		filepath.Clean(tmpFile),
		filepath.Clean(resolverFile),
	)
	return m.admin.RunWithPrivileges(cmd)
}

// Start installs dnsmasq as a root LaunchDaemon and starts it. It generates the
// dnsmasq config, ensures log files exist, then calls launchd.Install. If dnsmasq
// is not yet installed, it attempts to install the binary first.
func (m *Manager) Start() error {
	if !m.IsInstalled() {
		if err := m.binary.Install(); err != nil {
			return fmt.Errorf("dnsmasq is not installed: %w", err)
		}
	}

	if err := m.EnsureConfig(); err != nil {
		return fmt.Errorf("failed to generate dnsmasq config: %w", err)
	}

	// Ensure log files exist with user ownership
	ensureLogFile(m.paths.LogsDir() + "/dnsmasq-stdout.log")
	ensureLogFile(m.paths.LogsDir() + "/dnsmasq-stderr.log")

	// Install as LaunchDaemon (asks admin password once, then auto-starts on boot)
	return m.launchd.Install(system.ServiceConfig{
		Label:      ServiceLabel,
		Program:    m.binary.Find(),
		Args:       []string{"--keep-in-foreground", "-C", m.paths.DnsmasqConf()},
		RunAtLoad:  true,
		KeepAlive:  true,
		StdoutPath: m.paths.LogsDir() + "/dnsmasq-stdout.log",
		StderrPath: m.paths.LogsDir() + "/dnsmasq-stderr.log",
		Type:       system.ServiceDaemon,
	})
}

// Stop uninstalls the dnsmasq LaunchDaemon, which terminates the dnsmasq process.
func (m *Manager) Stop() error {
	return m.launchd.Uninstall(system.ServiceConfig{
		Label: ServiceLabel,
		Type:  system.ServiceDaemon,
	})
}

// Status returns a snapshot of the current dnsmasq service state, including whether
// the resolver file at /etc/resolver/{tld} is present.
func (m *Manager) Status() ServiceStatus {
	_, resolverErr := m.fs.Stat(m.paths.ResolverFile())
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.launchd.IsRunning(ServiceLabel),
		Resolver:  resolverErr == nil,
	}
}

func ensureLogFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte{}, 0644)
	}
}
