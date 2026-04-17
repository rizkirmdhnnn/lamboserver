package php

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FpmServiceLabel is the launchd agent label for the PHP-FPM service.
const FpmServiceLabel = "com.lamboserver.php-fpm"

// FpmStatus reports the current state of the PHP-FPM process.
type FpmStatus struct {
	Running bool   `json:"running"`
	Version string `json:"version"`
	Socket  string `json:"socket"`
}

// FpmManager manages the PHP-FPM LaunchAgent lifecycle for the active PHP version.
// It generates an FPM config file and registers the service with launchd so that
// PHP-FPM starts automatically on login.
type FpmManager struct {
	paths   *system.Paths
	store   *config.Store
	launchd LaunchdService
	fs      FileSystem
}

// NewFpmManager creates a FpmManager with injected dependencies.
// paths provides directory locations; store supplies the active PHP version at runtime.
func NewFpmManager(paths *system.Paths, store *config.Store, launchd LaunchdService, fs FileSystem) *FpmManager {
	return &FpmManager{paths: paths, store: store, launchd: launchd, fs: fs}
}

// SocketPath returns the unix socket path used by PHP-FPM (~/.lamboserver/php-fpm.sock).
// This path is shared across all PHP versions and referenced by Nginx site configs.
func (f *FpmManager) SocketPath() string {
	return filepath.Join(f.paths.Home, "php-fpm.sock")
}

// Start starts PHP-FPM for the currently active PHP version as a LaunchAgent.
// It generates the FPM configuration file and installs the launchd service with
// RunAtLoad and KeepAlive so PHP-FPM restarts automatically. Returns an error if
// no active PHP version is set or the FPM binary is not found.
func (f *FpmManager) Start() error {
	version := f.store.Get().ActivePhpVersion
	if version == "" {
		return fmt.Errorf("no active PHP version set")
	}

	fpmBin := filepath.Join(f.paths.PhpVersionDir(version), "sbin", "php-fpm")
	if _, err := f.fs.Stat(fpmBin); err != nil {
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

// Stop stops PHP-FPM by uninstalling the launchd agent, which terminates the process.
func (f *FpmManager) Stop() error {
	return f.launchd.Uninstall(system.ServiceConfig{
		Label: FpmServiceLabel,
		Type:  system.ServiceAgent,
	})
}

// Restart stops then starts PHP-FPM, picking up any active version change.
// Stop errors are intentionally ignored so a non-running FPM does not block restart.
func (f *FpmManager) Restart() error {
	f.Stop()
	return f.Start()
}

// Status returns a snapshot of the current PHP-FPM state including whether the
// launchd agent is running, the active PHP version, and the socket path.
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

	return f.fs.WriteFile(f.configPath(), []byte(conf), 0644)
}
