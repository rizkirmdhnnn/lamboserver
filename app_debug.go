package main

// --- Debug Methods ---

func (a *App) EnableDebug() error {
	if err := a.Debug.Enable(); err != nil {
		return err
	}
	return a.Config.SetDebugMode(true)
}

func (a *App) DisableDebug() {
	a.Debug.Disable()
	a.Config.SetDebugMode(false)
}

func (a *App) IsDebugEnabled() bool {
	return a.Debug.IsEnabled()
}

func (a *App) GetDebugLogPath() string {
	return a.Debug.LogPath()
}

func (a *App) ClearDebugLog() error {
	return a.Debug.ClearLog()
}

// --- Shell Integration Methods ---

func (a *App) IsShellIntegrated() bool {
	return a.Shell.IsInstalled()
}

func (a *App) InstallShellIntegration() error {
	a.Debug.Info("InstallShellIntegration called")
	err := a.Shell.Install()
	a.Debug.Action("InstallShellIntegration", err)
	return err
}

func (a *App) UninstallShellIntegration() error {
	a.Debug.Info("UninstallShellIntegration called")
	err := a.Shell.Uninstall()
	a.Debug.Action("UninstallShellIntegration", err)
	return err
}
