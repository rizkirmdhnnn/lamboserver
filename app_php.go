package main

import (
	"fmt"

	"github.com/rizkirmdhnnn/lamboserver/internal/php"
)

func (a *App) GetPhpVersions() ([]php.PhpVersion, error) {
	a.Debug.Info("GetPhpVersions called")
	versions, err := a.Php.ListInstalled()
	a.Debug.Action("GetPhpVersions", err)
	if err == nil {
		a.Debug.Info("Found %d PHP versions", len(versions))
	}
	return versions, err
}

func (a *App) GetAvailablePhpVersions() ([]php.AvailablePhpVersion, error) {
	a.Debug.Info("GetAvailablePhpVersions called")
	versions, err := a.Php.ListAvailable()
	a.Debug.Action("GetAvailablePhpVersions", err)
	return versions, err
}

func (a *App) InstallPhp(version string) error {
	a.Debug.Info("InstallPhp called: version=%s", version)
	err := a.Php.Install(version)
	a.Debug.Action(fmt.Sprintf("InstallPhp(%s)", version), err)
	return err
}

func (a *App) UninstallPhp(version string) error {
	a.Debug.Info("UninstallPhp called: version=%s", version)
	err := a.Php.Uninstall(version)
	a.Debug.Action(fmt.Sprintf("UninstallPhp(%s)", version), err)
	return err
}

func (a *App) SetActivePhp(version string) error {
	a.Debug.Info("SetActivePhp called: version=%s", version)

	versions, err := a.Php.ListInstalled()
	if err != nil {
		return err
	}

	var target *php.PhpVersion
	for _, v := range versions {
		if v.Version == version {
			target = &v
			break
		}
	}
	if target == nil {
		return fmt.Errorf("PHP version %s not found", version)
	}

	if err := a.Php.SetActive(version); err != nil {
		a.Debug.Action(fmt.Sprintf("SetActivePhp(%s)", version), err)
		return err
	}

	a.Shell.LinkPhpVersion(target.Binary, target.FpmBin, target.Path)
	a.Debug.Info("Updated PHP symlinks: %s -> %s", version, target.Binary)

	// Restart PHP-FPM to use the new version
	if a.PhpFpm.Status().Running {
		a.Debug.Info("Restarting PHP-FPM for version %s", version)
		if err := a.PhpFpm.Restart(); err != nil {
			a.Debug.Error("Failed to restart PHP-FPM: %v", err)
		}
	}

	a.Debug.Action(fmt.Sprintf("SetActivePhp(%s)", version), nil)
	return nil
}

// --- PHP-FPM Methods ---

func (a *App) GetPhpFpmStatus() php.FpmStatus {
	status := a.PhpFpm.Status()
	a.Debug.Info("GetPhpFpmStatus: running=%v, version=%s", status.Running, status.Version)
	return status
}

func (a *App) StartPhpFpm() error {
	a.Debug.Info("StartPhpFpm called")
	err := a.PhpFpm.Start()
	a.Debug.Action("StartPhpFpm", err)
	return err
}

func (a *App) StopPhpFpm() error {
	a.Debug.Info("StopPhpFpm called")
	err := a.PhpFpm.Stop()
	a.Debug.Action("StopPhpFpm", err)
	return err
}
