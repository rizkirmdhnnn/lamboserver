package dnsmasq

import (
	"io/fs"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FileSystem is the filesystem subset that dns.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	ReadFile(name string) ([]byte, error)
	MkdirAll(path string, perm fs.FileMode) error
}

// LaunchdService is the launchd subset that dns.Manager needs.
type LaunchdService interface {
	Install(cfg system.ServiceConfig) error
	Uninstall(cfg system.ServiceConfig) error
	IsRunning(label string) bool
}

// AdminRunner abstracts privilege elevation for DNS resolver setup.
type AdminRunner interface {
	RunWithPrivileges(command string) error
}
