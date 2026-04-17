package nginx

import (
	"io/fs"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FileSystem is the filesystem subset that nginx.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}

// LaunchdService is the launchd subset that nginx.Manager needs.
type LaunchdService interface {
	Install(cfg system.ServiceConfig) error
	Uninstall(cfg system.ServiceConfig) error
	IsRunning(label string) bool
}

// HelperRunner executes privileged helper actions.
// Abstracted separately from LaunchdService because the concrete *system.Helper
// is returned by value from LaunchdManager.Helper() — injecting it directly avoids
// the covariance mismatch that would arise if Helper() were part of LaunchdService.
type HelperRunner interface {
	Run(args ...string) (string, error)
}
