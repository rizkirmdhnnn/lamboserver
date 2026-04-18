package phpmyadmin

import (
	"io/fs"
	"os"
)

// ServiceStatus reports whether the phpMyAdmin web panel is installed.
type ServiceStatus struct {
	Installed bool `json:"installed"`
}

// FileSystem is the filesystem subset that phpmyadmin.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
	Remove(name string) error
	Create(name string) (*os.File, error)
}

// CertManager is the certificate manager subset that phpmyadmin.Manager needs.
type CertManager interface {
	GenerateCert(domain string) (certPath, keyPath string, err error)
	RemoveCert(domain string) error
}

// NginxReloader is the Nginx manager subset that phpmyadmin.Manager needs.
type NginxReloader interface {
	Reload() error
}

// CommandRunner is the exec subset that phpmyadmin.Manager needs.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}
