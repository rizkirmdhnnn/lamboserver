// Package site manages Nginx site configurations and SSL certificate provisioning
// for local development domains.
package sites

import (
	"io/fs"
	"os"
)

// FileSystem is the filesystem subset that site.Manager needs.
// Defined at the consumer site per Go interface convention (D-02).
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	Create(name string) (*os.File, error)
	Remove(name string) error
}
