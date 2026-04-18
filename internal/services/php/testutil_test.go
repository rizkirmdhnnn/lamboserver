// Package php_test provides test helpers and mock definitions for the php package.
package php_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/mock"
)

// --- Mock Definitions ---

// MockFileSystem implements php.FileSystem using testify/mock.
type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(fs.FileInfo), args.Error(1)
}

func (m *MockFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]fs.DirEntry), args.Error(1)
}

func (m *MockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	args := m.Called(name, data, perm)
	return args.Error(0)
}

func (m *MockFileSystem) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockFileSystem) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystem) Chmod(name string, mode fs.FileMode) error {
	args := m.Called(name, mode)
	return args.Error(0)
}

// Compile-time interface check.
var _ php.FileSystem = (*MockFileSystem)(nil)

// MockCommandRunner implements php.CommandRunner using testify/mock.
type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) Run(name string, args ...string) (string, error) {
	callArgs := make([]interface{}, 1+len(args))
	callArgs[0] = name
	for i, a := range args {
		callArgs[i+1] = a
	}
	ret := m.Called(callArgs...)
	return ret.String(0), ret.Error(1)
}

func (m *MockCommandRunner) LookPath(file string) (string, error) {
	args := m.Called(file)
	return args.String(0), args.Error(1)
}

// Compile-time interface check.
var _ php.CommandRunner = (*MockCommandRunner)(nil)

// MockLaunchdService implements php.LaunchdService using testify/mock.
type MockLaunchdService struct {
	mock.Mock
}

func (m *MockLaunchdService) Install(cfg system.ServiceConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockLaunchdService) Uninstall(cfg system.ServiceConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockLaunchdService) IsRunning(label string) bool {
	args := m.Called(label)
	return args.Bool(0)
}

// Compile-time interface check.
var _ php.LaunchdService = (*MockLaunchdService)(nil)

// MockDirEntry implements fs.DirEntry for use in ReadDir return values.
type MockDirEntry struct {
	name  string
	isDir bool
}

func (d *MockDirEntry) Name() string               { return d.name }
func (d *MockDirEntry) IsDir() bool                { return d.isDir }
func (d *MockDirEntry) Type() fs.FileMode          { return 0 }
func (d *MockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// Compile-time interface check.
var _ fs.DirEntry = (*MockDirEntry)(nil)

// --- Helper Factories ---

// newTestPaths returns a Paths struct rooted at a temporary directory.
func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

// newTestStore creates a Store backed by a temporary directory.
func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	dir := t.TempDir()
	return config.NewStore(filepath.Join(dir, "config.json"))
}

// newTestManager creates a Manager wired with mocks, returning the manager plus mocks for assertion setup.
func newTestManager(t *testing.T) (*php.Manager, *MockFileSystem, *MockCommandRunner, *system.Paths, *config.Store) {
	t.Helper()
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := new(MockFileSystem)
	mockCmd := new(MockCommandRunner)
	mgr := php.NewManager(paths, store, mockFS, mockCmd)
	return mgr, mockFS, mockCmd, paths, store
}

// newTestFpmManager creates a FpmManager wired with mocks, returning the manager plus mocks for assertion setup.
func newTestFpmManager(t *testing.T) (*php.FpmManager, *MockFileSystem, *MockLaunchdService, *system.Paths, *config.Store) {
	t.Helper()
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := new(MockFileSystem)
	mockLaunchd := new(MockLaunchdService)
	mgr := php.NewFpmManager(paths, store, mockLaunchd, mockFS)
	return mgr, mockFS, mockLaunchd, paths, store
}
