package nodejs_test

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// --- MockDirEntry ---

// MockDirEntry implements fs.DirEntry for testing.
type MockDirEntry struct {
	name  string
	isDir bool
}

func (e *MockDirEntry) Name() string               { return e.name }
func (e *MockDirEntry) IsDir() bool                { return e.isDir }
func (e *MockDirEntry) Type() fs.FileMode          { return 0 }
func (e *MockDirEntry) Info() (fs.FileInfo, error) { return &mockFileInfo{name: e.name, isDir: e.isDir}, nil }

type mockFileInfo struct {
	name  string
	isDir bool
}

func (fi *mockFileInfo) Name() string      { return fi.name }
func (fi *mockFileInfo) Size() int64       { return 0 }
func (fi *mockFileInfo) Mode() fs.FileMode { return 0755 }
func (fi *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *mockFileInfo) IsDir() bool       { return fi.isDir }
func (fi *mockFileInfo) Sys() any          { return nil }

// --- MockFileSystem ---

// MockFileSystem implements nodejs.FileSystem for testing.
type MockFileSystem struct {
	ReadDirFunc  func(name string) ([]fs.DirEntry, error)
	MkdirAllFunc func(path string, perm fs.FileMode) error
	RemoveAllFunc func(path string) error

	// Call tracking
	ReadDirCalls  []string
	MkdirAllCalls []string
	RemoveAllCalls []string
}

func (m *MockFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	m.ReadDirCalls = append(m.ReadDirCalls, name)
	if m.ReadDirFunc != nil {
		return m.ReadDirFunc(name)
	}
	return nil, nil
}

func (m *MockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	m.MkdirAllCalls = append(m.MkdirAllCalls, path)
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(path, perm)
	}
	return nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	m.RemoveAllCalls = append(m.RemoveAllCalls, path)
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(path)
	}
	return nil
}

// --- MockCommandRunner ---

// MockCommandRunner implements nodejs.CommandRunner for testing.
type MockCommandRunner struct {
	RunFunc func(name string, args ...string) (string, error)

	// Call tracking
	RunCalls []RunCall
}

type RunCall struct {
	Name string
	Args []string
}

func (m *MockCommandRunner) Run(name string, args ...string) (string, error) {
	m.RunCalls = append(m.RunCalls, RunCall{Name: name, Args: args})
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return "", nil
}

// --- Test helpers ---

func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	tmp := t.TempDir()
	return config.NewStore(filepath.Join(tmp, "config.json"))
}
