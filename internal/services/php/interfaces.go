// Package php provides PHP version management and FPM lifecycle control.
package php

import (
	"io/fs"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FileSystem is the filesystem subset that php.Manager and detector need.
// Defined at consumer site per D-02.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Remove(name string) error
	RemoveAll(path string) error
	Chmod(name string, mode fs.FileMode) error
}

// CommandRunner is the exec subset that php.Manager and detector need.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
	LookPath(file string) (string, error)
}

// LaunchdService is the launchd subset that php.FpmManager needs.
type LaunchdService interface {
	Install(cfg system.ServiceConfig) error
	Uninstall(cfg system.ServiceConfig) error
	IsRunning(label string) bool
}
