// Package integration_test provides integration tests for critical LamboServer workflows.
// Tests use real filesystem operations in t.TempDir() and mock launchd/admin to avoid
// privileged operations during CI.
package integration_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/pkg/logger"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- RealTestFS ---

// RealTestFS implements both cert.FileSystem and php.FileSystem using real os.* calls.
// All operations go against the real filesystem (t.TempDir() as root). Required because
// pem.Encode and other operations write to *os.File instances that must be real files.
type RealTestFS struct {
	root string
}

func newRealTestFS(root string) *RealTestFS {
	return &RealTestFS{root: root}
}

// cert.FileSystem methods

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

// php.FileSystem methods (in addition to the overlapping cert.FileSystem methods)

func (r *RealTestFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (r *RealTestFS) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *RealTestFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

func (r *RealTestFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (r *RealTestFS) Chmod(name string, mode fs.FileMode) error {
	return os.Chmod(name, mode)
}

// Compile-time interface checks.
var _ cert.FileSystem = (*RealTestFS)(nil)
var _ php.FileSystem = (*RealTestFS)(nil)

// --- MockAdminRunner ---

// MockAdminRunner is a testify mock implementing cert.AdminRunner.
// Default: RunWithPrivileges returns nil (success) for TrustCA calls.
type MockAdminRunner struct {
	mock.Mock
}

func (m *MockAdminRunner) RunWithPrivileges(command string) error {
	args := m.Called(command)
	return args.Error(0)
}

// Compile-time interface check.
var _ cert.AdminRunner = (*MockAdminRunner)(nil)

// --- MockLaunchdService ---

// MockLaunchdService is a testify mock implementing php.LaunchdService.
// Used for FPM lifecycle in version switch tests — prevents real launchctl calls.
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

// --- MockCommandRunner ---

// MockCommandRunner is a testify mock implementing php.CommandRunner.
// Used for PHP detection — prevents real exec calls.
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

// --- Helper Factories ---

// newTestPaths returns a Paths struct with Home set to a dedicated temp subdirectory.
// The .lamboserver home is created within a t.TempDir() so it is auto-cleaned.
func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	base := t.TempDir()
	return &system.Paths{Home: filepath.Join(base, ".lamboserver")}
}

// newTestStore creates a config.Store backed by a temp config file derived from paths.
func newTestStore(t *testing.T, paths *system.Paths) *config.Store {
	t.Helper()
	require.NoError(t, os.MkdirAll(paths.Home, 0755))
	return config.NewStore(paths.ConfigFile())
}

// newTestCertManager creates a cert.Manager backed by RealTestFS and MockAdminRunner.
// EnsureDirectories is called so the certs directory exists before tests run.
func newTestCertManager(t *testing.T, paths *system.Paths) (*cert.Manager, *MockAdminRunner) {
	t.Helper()
	require.NoError(t, paths.EnsureDirectories())
	admin := &MockAdminRunner{}
	rfs := newRealTestFS(paths.Home)
	mgr := cert.NewManager(paths, rfs, admin)
	return mgr, admin
}

// newTestShell creates a system.Integration with WithHome pointing to a separate temp dir
// so shell rc file operations are isolated from the paths home.
func newTestShell(t *testing.T, paths *system.Paths) *system.Integration {
	t.Helper()
	shellHome := t.TempDir()
	return system.NewIntegration(paths).WithHome(shellHome)
}

// newTestDebugLogger creates a logger.Logger that writes to paths.LogsDir().
func newTestDebugLogger(t *testing.T, paths *system.Paths) *logger.Logger {
	t.Helper()
	require.NoError(t, os.MkdirAll(paths.LogsDir(), 0755))
	return logger.NewLogger(paths.LogsDir())
}
