// Package node provides Node.js version management.
package nodejs

import "io/fs"

// FileSystem is the filesystem subset that node.Manager needs.
type FileSystem interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
}

// CommandRunner is the exec subset that node.Manager needs.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}
