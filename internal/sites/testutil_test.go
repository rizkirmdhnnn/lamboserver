// Package site_test provides test helpers for the site package.
package sites_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/sites"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSiteFS implements site.FileSystem.
// Stat and Remove delegate to mock for test control.
// Create returns a real temp file because template.Execute needs a real writer.
type MockSiteFS struct {
	mock.Mock
	tmpDir      string
	createdFiles map[string]string // maps requested path -> actual temp file path
}

func newMockSiteFS(t *testing.T) *MockSiteFS {
	t.Helper()
	return &MockSiteFS{
		tmpDir:      t.TempDir(),
		createdFiles: make(map[string]string),
	}
}

func (m *MockSiteFS) Stat(name string) (fs.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(fs.FileInfo), args.Error(1)
}

// Create returns a real writable file (required for template.Execute).
// The actual file is created in tmpDir; the mapping is stored for later assertions.
func (m *MockSiteFS) Create(name string) (*os.File, error) {
	// Create a real file in our temp directory using the basename of the requested path
	realPath := filepath.Join(m.tmpDir, filepath.Base(name))
	if err := os.MkdirAll(filepath.Dir(realPath), 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(realPath)
	if err != nil {
		return nil, err
	}
	m.createdFiles[name] = realPath
	return f, nil
}

// CreatedFilePath returns the real temp file path for a given requested path.
// Used by tests to read back written content.
func (m *MockSiteFS) CreatedFilePath(name string) string {
	return m.createdFiles[name]
}

func (m *MockSiteFS) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// MockAdminRunner is a testify mock implementing cert.AdminRunner.
type MockAdminRunner struct {
	mock.Mock
}

func (m *MockAdminRunner) RunWithPrivileges(command string) error {
	args := m.Called(command)
	return args.Error(0)
}

// realCertFS implements cert.FileSystem using real os.* calls
// against a temp directory (same approach as cert testutil_test.go).
type realCertFS struct{}

func (r *realCertFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r *realCertFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r *realCertFS) Create(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.Create(name)
}

func (r *realCertFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(name, flag, perm)
}

func (r *realCertFS) Remove(name string) error {
	return os.Remove(name)
}

// newTestPaths returns a Paths struct rooted at a temp directory.
func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

// newTestStore creates a Store backed by a temporary directory.
func newTestStore(t *testing.T, paths *system.Paths) *config.Store {
	t.Helper()
	return config.NewStore(paths.ConfigFile())
}

// newTestCertManager creates a real cert.Manager backed by a real filesystem
// writing to paths' temp directories. Sets up the CA so GenerateCert works.
// The MockAdminRunner is set to accept TrustCA calls with nil error.
func newTestCertManager(t *testing.T, paths *system.Paths) *cert.Manager {
	t.Helper()
	require.NoError(t, paths.EnsureDirectories())
	adminMock := &MockAdminRunner{}
	adminMock.On("RunWithPrivileges", mock.Anything).Return(nil)
	mgr := cert.NewManager(paths, &realCertFS{}, adminMock)
	require.NoError(t, mgr.SetupCA())
	return mgr
}

// newTestSiteManager creates a site.Manager wired to test doubles.
// Returns the manager, MockSiteFS, paths, and config store.
func newTestSiteManager(t *testing.T) (*sites.Manager, *MockSiteFS, *system.Paths, *config.Store) {
	t.Helper()
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())
	store := newTestStore(t, paths)
	certMgr := newTestCertManager(t, paths)
	siteFS := newMockSiteFS(t)
	mgr := sites.NewManager(paths, store, certMgr, siteFS)
	return mgr, siteFS, paths, store
}
