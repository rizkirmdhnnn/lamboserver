package nginx

import (
	"fmt"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

const ServiceLabel = "com.lamboserver.nginx"

type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
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
			Name:      "nginx",
			LocalPath: paths.NginxBin(),
		},
	}
}

func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// IsDaemonInstalled checks if the LaunchDaemon plist is already installed
func (m *Manager) IsDaemonInstalled() bool {
	_, err := os.Stat("/Library/LaunchDaemons/" + ServiceLabel + ".plist")
	return err == nil
}

func (m *Manager) Start() error {
	if !m.IsInstalled() {
		if err := m.binary.Install(); err != nil {
			return fmt.Errorf("nginx is not installed: %w", err)
		}
	}

	if err := m.EnsureConfig(); err != nil {
		return fmt.Errorf("failed to generate nginx config: %w", err)
	}

	// Ensure log files exist with user ownership before daemon starts as root
	ensureLogFile(m.paths.LogsDir() + "/nginx-stdout.log")
	ensureLogFile(m.paths.LogsDir() + "/nginx-stderr.log")
	ensureLogFile(m.paths.NginxErrorLog())
	ensureLogFile(m.paths.NginxAccessLog())

	// Install as LaunchDaemon (asks admin password once, then auto-starts on boot)
	return m.launchd.Install(system.ServiceConfig{
		Label:      ServiceLabel,
		Program:    m.binary.Find(),
		Args:       []string{"-c", m.paths.NginxConf(), "-g", "daemon off;"},
		RunAtLoad:  true,
		KeepAlive:  true,
		StdoutPath: m.paths.LogsDir() + "/nginx-stdout.log",
		StderrPath: m.paths.LogsDir() + "/nginx-stderr.log",
		Type:       system.ServiceDaemon,
	})
}

func (m *Manager) Stop() error {
	return m.launchd.Uninstall(system.ServiceConfig{
		Label: ServiceLabel,
		Type:  system.ServiceDaemon,
	})
}

func (m *Manager) Reload() error {
	// Reload via helper (root) since nginx master runs as root LaunchDaemon
	_, err := m.launchd.Helper().Run("reload-nginx")
	return err
}

func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.launchd.IsRunning(ServiceLabel),
	}
}

func (m *Manager) EnsureConfig() error {
	return m.generateMasterConfig()
}

func ensureLogFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte{}, 0644)
	}
}
