package dnsmasq

import (
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// NewManagerWithDeps exposes the internal constructor for use in external test packages.
func NewManagerWithDeps(paths *system.Paths, launchd LaunchdService, fs FileSystem, admin AdminRunner) *Manager {
	return newManagerWithDeps(paths, launchd, fs, admin)
}
