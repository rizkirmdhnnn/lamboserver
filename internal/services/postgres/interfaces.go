// Package postgresql provides PostgreSQL service lifecycle management for LamboServer.
// It handles binary installation, data directory initialization, and pg_ctl-based
// start/stop lifecycle — without using launchd (PostgreSQL runs as the current user).
package postgres

import (
	"database/sql"
	"io/fs"
	"regexp"
)

// ServiceLabel is the identifier for the PostgreSQL service within LamboServer.
// (No launchd label needed — pg_ctl manages the process directly.)
const ServiceLabel = "com.lamboserver.postgresql"

// postgresqlVersion is the PostgreSQL 17 LTS version downloaded from
// github.com/theseus-rs/postgresql-binaries.
const postgresqlVersion = "17.9.0"

// ServiceStatus reports whether the PostgreSQL binary is present, the server is
// running, and the port it is configured on.
type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Port      int  `json:"port"`
}

// FileSystem is the filesystem subset that postgresql.Manager needs.
// ReadFile is added beyond MySQL's interface to support testable postmaster.pid reads.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

// CommandRunner is the exec subset that postgresql.Manager needs for non-privileged commands.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}

// HelperRunner executes privileged helper actions via the lambo-helper script.
type HelperRunner interface {
	Run(args ...string) (string, error)
}

// AdminRunner abstracts privilege elevation for operations that need root access.
type AdminRunner interface {
	RunWithPrivileges(command string) error
}

// DBOpener abstracts database/sql.Open to allow mocking in tests.
type DBOpener interface {
	Open(driverName, dataSourceName string) (*sql.DB, error)
}

// PgDatabase represents a user database with its human-readable size.
type PgDatabase struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// dbNameRe validates database names; only alphanumeric characters and underscores
// are allowed, with a maximum length of 63 characters (PostgreSQL NAMEDATALEN-1).
var dbNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{1,63}$`)

// systemDatabases lists databases that must not be dropped.
var systemDatabases = map[string]bool{
	"postgres":  true,
	"template0": true,
	"template1": true,
}
