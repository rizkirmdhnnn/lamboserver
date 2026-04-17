// Package cert manages local CA generation and self-signed certificate creation
// for development sites.
package cert

import (
	"io/fs"
	"os"
)

// FileSystem is the filesystem subset that cert.Manager needs.
// Defined at the consumer site per Go interface convention (D-02).
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	Create(name string) (*os.File, error)
	OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error)
	Remove(name string) error
}

// AdminRunner abstracts privilege elevation for CA trust operations.
type AdminRunner interface {
	RunWithPrivileges(command string) error
}
