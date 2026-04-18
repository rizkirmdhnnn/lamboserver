package process

import (
	"path/filepath"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// LaunchdProcessManager implements ProcessManager using macOS launchd.
// It wraps the existing system.LaunchdManager to adapt it to the
// platform-neutral ProcessManager interface.
type LaunchdProcessManager struct {
	launchd *system.LaunchdManager
	logDir  string
}

// NewLaunchdProcessManager creates a LaunchdProcessManager wrapping the given LaunchdManager.
func NewLaunchdProcessManager(launchd *system.LaunchdManager, logDir string) *LaunchdProcessManager {
	return &LaunchdProcessManager{
		launchd: launchd,
		logDir:  logDir,
	}
}

// toServiceConfig converts a ProcessConfig to a system.ServiceConfig for launchd.
func (l *LaunchdProcessManager) toServiceConfig(cfg ProcessConfig) system.ServiceConfig {
	serviceType := system.ServiceAgent
	if cfg.Privileged {
		serviceType = system.ServiceDaemon
	}

	logDir := cfg.LogDir
	if logDir == "" {
		logDir = l.logDir
	}

	return system.ServiceConfig{
		Label:                cfg.Label,
		Program:              cfg.Program,
		Args:                 cfg.Args,
		RunAtLoad:            cfg.RunAtLoad,
		KeepAlive:            cfg.KeepAlive,
		WorkingDir:           cfg.WorkingDir,
		StdoutPath:           filepath.Join(logDir, cfg.Label+".log"),
		StderrPath:           filepath.Join(logDir, cfg.Label+"-error.log"),
		Type:                 serviceType,
		EnvironmentVariables: cfg.Env,
	}
}

// Start installs and loads a launchd plist for the service.
func (l *LaunchdProcessManager) Start(name string, cfg ProcessConfig) error {
	return l.launchd.Install(l.toServiceConfig(cfg))
}

// Stop unloads and removes the launchd plist for the service.
func (l *LaunchdProcessManager) Stop(name string) error {
	// We need the original config to determine daemon vs agent type.
	// For now, try stopping as both - launchd.Stop just sends a signal by label.
	return l.launchd.Stop(name)
}

// Restart stops then starts the service.
func (l *LaunchdProcessManager) Restart(name string, cfg ProcessConfig) error {
	// Uninstall the old plist, then install a fresh one.
	_ = l.launchd.Uninstall(l.toServiceConfig(cfg))
	return l.launchd.Install(l.toServiceConfig(cfg))
}

// IsRunning checks if the launchd service is currently loaded and running.
func (l *LaunchdProcessManager) IsRunning(name string) bool {
	return l.launchd.IsRunning(name)
}
