package system

import (
	"io/fs"
	"os"
)

// RealFS delegates all filesystem operations to the os package.
// It satisfies the FileSystem interfaces defined in consumer packages
// (php, nginx, dns, node, site, cert). Each consumer defines its own
// interface with only the methods it needs (per D-02); RealFS implements
// the superset of all those methods.
type RealFS struct{}

// Stat returns file info for the named file.
func (RealFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

// ReadDir reads the directory named by dirname and returns a list of entries.
func (RealFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }

// MkdirAll creates the directory path along with any necessary parents.
func (RealFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

// WriteFile writes data to the named file, creating it if necessary.
func (RealFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Remove removes the named file or empty directory.
func (RealFS) Remove(name string) error { return os.Remove(name) }

// RemoveAll removes path and any children it contains.
func (RealFS) RemoveAll(path string) error { return os.RemoveAll(path) }

// ReadFile reads the named file and returns its contents.
func (RealFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// Create creates or truncates the named file.
func (RealFS) Create(name string) (*os.File, error) { return os.Create(name) }

// OpenFile opens the named file with the given flag and permission bits.
func (RealFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// Rename renames (moves) oldpath to newpath.
func (RealFS) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }

// Chmod changes the mode of the named file.
func (RealFS) Chmod(name string, mode fs.FileMode) error { return os.Chmod(name, mode) }
