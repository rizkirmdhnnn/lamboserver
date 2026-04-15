package php

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

const FpmServiceLabel = "com.lamboserver.php-fpm"

type FpmStatus struct {
	Running bool   `json:"running"`
	Version string `json:"version"`
	Socket  string `json:"socket"`
}

type FpmManager struct {
	paths   *system.Paths
	store   *config.Store
	launchd *system.LaunchdManager
}

func NewFpmManager(paths *system.Paths, store *config.Store, launchd *system.LaunchdManager) *FpmManager {
	return &FpmManager{paths: paths, store: store, launchd: launchd}
}

// SocketPath returns the php-fpm socket path for the active version
func (f *FpmManager) SocketPath() string {
	return filepath.Join(f.paths.Home, "php-fpm.sock")
}

// Start starts php-fpm for the active PHP version
func (f *FpmManager) Start() error {
	version := f.store.Get().ActivePhpVersion
	if version == "" {
		return fmt.Errorf("no active PHP version set")
	}

	fpmBin := filepath.Join(f.paths.PhpVersionDir(version), "sbin", "php-fpm")
	if _, err := os.Stat(fpmBin); err != nil {
		return fmt.Errorf("php-fpm binary not found for version %s", version)
	}

	// Generate fpm config
	if err := f.generateConfig(version); err != nil {
		return fmt.Errorf("failed to generate fpm config: %w", err)
	}

	confPath := f.configPath()

	return f.launchd.Install(system.ServiceConfig{
		Label:      FpmServiceLabel,
		Program:    fpmBin,
		Args:       []string{"--nodaemonize", "--fpm-config", confPath},
		RunAtLoad:  true,
		KeepAlive:  true,
		StdoutPath: f.paths.LogsDir() + "/php-fpm-stdout.log",
		StderrPath: f.paths.LogsDir() + "/php-fpm-stderr.log",
		Type:       system.ServiceAgent,
	})
}

// Stop stops php-fpm
func (f *FpmManager) Stop() error {
	return f.launchd.Uninstall(system.ServiceConfig{
		Label: FpmServiceLabel,
		Type:  system.ServiceAgent,
	})
}

// Restart stops then starts php-fpm (e.g. after PHP version switch)
func (f *FpmManager) Restart() error {
	f.Stop()
	return f.Start()
}

// Status returns the current fpm status
func (f *FpmManager) Status() FpmStatus {
	return FpmStatus{
		Running: f.launchd.IsRunning(FpmServiceLabel),
		Version: f.store.Get().ActivePhpVersion,
		Socket:  f.SocketPath(),
	}
}

func (f *FpmManager) configPath() string {
	return filepath.Join(f.paths.Home, "php-fpm.conf")
}

func (f *FpmManager) generateConfig(version string) error {
	username := os.Getenv("USER")
	if username == "" {
		username = "nobody"
	}

	conf := fmt.Sprintf(`[global]
error_log = %s/php-fpm.log
daemonize = no

[lamboserver]
user = %s
group = staff
listen = %s
listen.owner = %s
listen.group = staff
listen.mode = 0777

pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 5

php_admin_value[error_log] = %s/php-fpm.log
php_admin_flag[log_errors] = on
`,
		f.paths.LogsDir(),
		username,
		f.SocketPath(),
		username,
		f.paths.LogsDir(),
	)

	return os.WriteFile(f.configPath(), []byte(conf), 0644)
}
