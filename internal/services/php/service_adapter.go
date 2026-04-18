package php

import (
	"bufio"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/services"
)

// ServiceAdapter is a facade that composes Manager (version management) and
// FpmManager (process lifecycle) to implement the unified services.VersionedService
// interface.
type ServiceAdapter struct {
	mgr *Manager
	fpm *FpmManager
}

// NewServiceAdapter creates a ServiceAdapter wrapping the given Manager and FpmManager.
func NewServiceAdapter(mgr *Manager, fpm *FpmManager) *ServiceAdapter {
	return &ServiceAdapter{mgr: mgr, fpm: fpm}
}

// Compile-time interface check.
var _ services.VersionedService = (*ServiceAdapter)(nil)

// Start starts PHP-FPM for the active version.
func (a *ServiceAdapter) Start() error {
	return a.fpm.Start()
}

// Stop stops PHP-FPM.
func (a *ServiceAdapter) Stop() error {
	return a.fpm.Stop()
}

// Restart restarts PHP-FPM.
func (a *ServiceAdapter) Restart() error {
	return a.fpm.Restart()
}

// Install downloads and installs the given PHP version.
func (a *ServiceAdapter) Install(version string) error {
	return a.mgr.Install(version)
}

// Status returns the unified service status for PHP-FPM.
func (a *ServiceAdapter) Status() services.ServiceStatus {
	s := a.fpm.Status()
	if s.Running {
		return services.StatusRunning
	}
	return services.StatusStopped
}

// Logs reads the last 100 lines from the PHP-FPM log.
func (a *ServiceAdapter) Logs() ([]string, error) {
	logPath := a.mgr.paths.LogsDir() + "/php-fpm.log"
	return readLastNLines(logPath, 100)
}

// Version returns the currently active PHP version string.
func (a *ServiceAdapter) Version() string {
	return a.mgr.store.Get().ActivePhpVersion
}

// SwitchVersion changes the active PHP version and restarts FPM.
func (a *ServiceAdapter) SwitchVersion(version string) error {
	if err := a.mgr.SetActive(version); err != nil {
		return err
	}
	return a.fpm.Restart()
}

// InstalledVersions returns the version strings of all installed PHP versions.
func (a *ServiceAdapter) InstalledVersions() []string {
	versions, err := a.mgr.ListInstalled()
	if err != nil {
		return nil
	}
	var result []string
	for _, v := range versions {
		result = append(result, v.Version)
	}
	return result
}

// ActiveVersion returns the currently active PHP version.
func (a *ServiceAdapter) ActiveVersion() string {
	return a.mgr.store.Get().ActivePhpVersion
}

func readLastNLines(filePath string, n int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, scanner.Err()
}
