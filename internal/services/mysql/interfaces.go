package mysql

import (
	"database/sql"
	"io/fs"
	"regexp"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// ServiceLabel is the launchd daemon label used for the MySQL service.
const ServiceLabel = "com.lamboserver.mysql"

// dbNameRe validates database names; only alphanumeric characters and underscores
// are allowed, with a maximum length of 64 characters (per MySQL limit).
var dbNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)

// ServiceStatus reports whether the MySQL binary is present, the daemon is running,
// and the port it is configured on.
type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Port      int  `json:"port"`
}

// FileSystem is the filesystem subset that mysql.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
	ReadDir(name string) ([]fs.DirEntry, error)
}

// LaunchdService is the launchd subset that mysql.Manager needs.
type LaunchdService interface {
	Install(cfg system.ServiceConfig) error
	Uninstall(cfg system.ServiceConfig) error
	IsRunning(label string) bool
}

// CommandRunner is the exec subset that mysql.Manager needs for non-privileged commands.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}

// HelperRunner executes privileged helper actions via the lambo-helper script.
type HelperRunner interface {
	Run(args ...string) (string, error)
}

// AdminRunner abstracts privilege elevation for operations that need root access
// (e.g. mkdir/chown for MySQL data dir). Uses osascript admin dialog.
type AdminRunner interface {
	RunWithPrivileges(command string) error
}

// DBOpener abstracts database/sql.Open to allow mocking in tests.
type DBOpener interface {
	Open(driverName, dataSourceName string) (*sql.DB, error)
}
