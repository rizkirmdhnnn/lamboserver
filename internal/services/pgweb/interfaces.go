// Package pgweb provides lifecycle management for the pgweb PostgreSQL web admin tool.
// pgweb is a standalone Go binary that serves a browser-based PostgreSQL client on localhost.
package pgweb

import "io/fs"

// pgwebVersion is the pinned pgweb release downloaded from sosedoff/pgweb.
const pgwebVersion = "0.17.0"

// downloadURLTemplate is the URL template for the pgweb zip archive.
// Parameters: version, arch (arm64 or amd64).
const downloadURLTemplate = "https://github.com/sosedoff/pgweb/releases/download/v%s/pgweb_darwin_%s.zip"

// defaultPort is the HTTP port pgweb binds to.
const defaultPort = 8081

// DaemonWebAdminService is the interface implemented by pgweb.Manager.
// It extends web admin capabilities with daemon lifecycle (Start/Stop/IsRunning).
type DaemonWebAdminService interface {
	Install() error
	IsInstalled() bool
	URL() string
	Version() string
	Start() error
	Stop() error
	IsRunning() bool
}

// ServiceStatus reports the current state of the pgweb tool.
type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Port      int  `json:"port"`
}

// FileSystem is the filesystem subset that pgweb.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
}

// CommandRunner is the exec subset that pgweb.Manager needs for non-daemon commands.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}
