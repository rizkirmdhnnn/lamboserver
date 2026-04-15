package dns

import (
	"fmt"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

const ServiceLabel = "com.lamboserver.dnsmasq"

type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Resolver  bool `json:"resolver"`
}

type Manager struct {
	paths   *system.Paths
	launchd *system.LaunchdManager
	binary  *system.BinaryLocator
}

func NewManager(paths *system.Paths, launchd *system.LaunchdManager) *Manager {
	return &Manager{
		paths:   paths,
		launchd: launchd,
		binary: &system.BinaryLocator{
			Name:      "dnsmasq",
			LocalPath: paths.DnsmasqBin(),
		},
	}
}

func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// IsDaemonInstalled checks if the LaunchDaemon plist is already installed
func (m *Manager) IsDaemonInstalled() bool {
	_, err := os.Stat("/Library/LaunchDaemons/" + ServiceLabel + ".plist")
	return err == nil
}

func (m *Manager) EnsureConfig() error {
	conf := fmt.Sprintf("address=/.%s/127.0.0.1\nlisten-address=127.0.0.1\nport=53\n", system.TestTLD)
	return os.WriteFile(m.paths.DnsmasqConf(), []byte(conf), 0644)
}

func (m *Manager) EnsureResolver() error {
	content := "nameserver 127.0.0.1\n"
	resolverFile := m.paths.ResolverFile()

	if existing, err := os.ReadFile(resolverFile); err == nil && string(existing) == content {
		return nil
	}

	tmpFile := os.TempDir() + "/lambo-resolver"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp resolver file: %w", err)
	}

	cmd := fmt.Sprintf("mkdir -p %s && cp %s %s", m.paths.ResolverDir(), tmpFile, resolverFile)
	return system.RunWithAdminPrivileges(cmd)
}

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

func (m *Manager) Stop() error {
	return m.launchd.Uninstall(system.ServiceConfig{
		Label: ServiceLabel,
		Type:  system.ServiceDaemon,
	})
}

func (m *Manager) Status() ServiceStatus {
	_, resolverErr := os.Stat(m.paths.ResolverFile())
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
