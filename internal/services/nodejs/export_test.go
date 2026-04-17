package nodejs

import (
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// NewManagerWithDeps exposes the internal constructor for use in external test packages.
func NewManagerWithDeps(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager {
	return newManagerWithDeps(paths, store, fs, cmd)
}
