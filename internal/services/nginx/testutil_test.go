// Package nginx_test provides test helpers and mock definitions for the nginx package.
package nginx_test

import (
	"io/fs"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/mock"
)

// --- Mock definitions ---

// MockFileSystem is a testify mock implementing nginx.FileSystem.
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

func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	args := m.Called(name, data, perm)
	return args.Error(0)
}

func (m *MockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

// MockLaunchdService is a testify mock implementing nginx.LaunchdService.
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

// MockHelperRunner is a testify mock implementing nginx.HelperRunner.
type MockHelperRunner struct {
	mock.Mock
}

func (m *MockHelperRunner) Run(args ...string) (string, error) {
	callArgs := m.Called(args)
	return callArgs.String(0), callArgs.Error(1)
}

// --- captureWriteFile helper ---

// capturedWrites holds path->data mappings captured from WriteFile calls.
type capturedWrites map[string][]byte

// captureWriteFile sets up the MockFileSystem to capture all WriteFile calls.
// It returns a capturedWrites map that will be populated as WriteFile is called.
func captureWriteFile(mockFS *MockFileSystem) capturedWrites {
	writes := make(capturedWrites)
	mockFS.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			path := args.String(0)
			data := args.Get(1).([]byte)
			writes[path] = data
		}).
		Return(nil)
	return writes
}

// --- Factory helpers ---

// newTestPaths returns a Paths struct rooted at a temporary directory.
func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

// newTestManager creates a Manager wired with mock dependencies for testing.
// Returns the manager, the three mocks, and the paths for assertion helpers.
func newTestManager(t *testing.T) (*nginx.Manager, *MockFileSystem, *MockLaunchdService, *MockHelperRunner, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{}
	mockLaunchd := &MockLaunchdService{}
	mockHelper := &MockHelperRunner{}
	mgr := nginx.NewManager(paths, mockLaunchd, mockFS, mockHelper)
	return mgr, mockFS, mockLaunchd, mockHelper, paths
}
