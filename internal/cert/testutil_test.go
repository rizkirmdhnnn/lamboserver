// Package cert_test provides test helpers for the cert package.
package cert_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// RealTestFS implements cert.FileSystem by delegating to real os.* calls
// against a temp directory. This is required because pem.Encode writes to
// *os.File instances that must be real writable files.
type RealTestFS struct {
	root string
}

func newRealTestFS(root string) *RealTestFS {
	return &RealTestFS{root: root}
}

func (r *RealTestFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r *RealTestFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r *RealTestFS) Create(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.Create(name)
}

func (r *RealTestFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(name, flag, perm)
}

func (r *RealTestFS) Remove(name string) error {
	return os.Remove(name)
}

// MockAdminRunner is a testify mock implementing cert.AdminRunner.
type MockAdminRunner struct {
	mock.Mock
}

func (m *MockAdminRunner) RunWithPrivileges(command string) error {
	args := m.Called(command)
	return args.Error(0)
}

// newTestPaths returns a Paths struct rooted at a temp directory.
func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

// newTestCertManager creates a cert.Manager backed by a real filesystem
// (writes to t.TempDir()) and a MockAdminRunner.
// It also calls EnsureDirectories so cert directory structure exists.
func newTestCertManager(t *testing.T) (*cert.Manager, *system.Paths, *MockAdminRunner) {
	t.Helper()
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())
	adminMock := &MockAdminRunner{}
	rfs := newRealTestFS(paths.Home)
	mgr := cert.NewManager(paths, rfs, adminMock)
	return mgr, paths, adminMock
}
