package main

import (
	"fmt"

	applog "github.com/rizkirmdhnnn/lamboserver/internal/log"
	"github.com/rizkirmdhnnn/lamboserver/internal/node"
)

// --- Node Methods ---

func (a *App) GetNodeVersions() ([]node.NodeVersion, error) {
	a.Debug.Info("GetNodeVersions called")
	versions, err := a.Node.ListInstalled()
	a.Debug.Action("GetNodeVersions", err)
	return versions, err
}

func (a *App) GetAvailableNodeVersions() ([]string, error) {
	a.Debug.Info("GetAvailableNodeVersions called")
	versions, err := a.Node.ListAvailable()
	a.Debug.Action("GetAvailableNodeVersions", err)
	return versions, err
}

func (a *App) InstallNode(version string) error {
	a.Debug.Info("InstallNode called: version=%s", version)
	err := a.Node.Install(version)
	a.Debug.Action(fmt.Sprintf("InstallNode(%s)", version), err)
	return err
}

func (a *App) SetActiveNode(version string) error {
	a.Debug.Info("SetActiveNode called: version=%s", version)

	if err := a.Node.SetActive(version); err != nil {
		a.Debug.Action(fmt.Sprintf("SetActiveNode(%s)", version), err)
		return err
	}

	nodeDir := a.Paths.NodeVersionDir(version)
	a.Shell.LinkNodeBinaries(nodeDir)
	a.Debug.Info("Updated Node symlinks: %s -> %s", version, nodeDir)
	a.Debug.Action(fmt.Sprintf("SetActiveNode(%s)", version), nil)
	return nil
}

func (a *App) UninstallNode(version string) error {
	a.Debug.Info("UninstallNode called: version=%s", version)
	err := a.Node.Uninstall(version)
	a.Debug.Action(fmt.Sprintf("UninstallNode(%s)", version), err)
	return err
}

// --- Log Methods ---

func (a *App) GetLogFiles() []applog.LogFile {
	logFiles := a.Logs.ListLogFiles()
	a.Debug.Info("GetLogFiles: %d files", len(logFiles))
	return logFiles
}

func (a *App) ReadLog(filePath string, lines int) ([]string, error) {
	a.Debug.Info("ReadLog called: path=%s, lines=%d", filePath, lines)
	data, err := a.Logs.ReadLastN(filePath, lines)
	a.Debug.Action(fmt.Sprintf("ReadLog(%s)", filePath), err)
	return data, err
}
