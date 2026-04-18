// Package phpmyadmin_test contains unit tests for the phpmyadmin package.
package phpmyadmin_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/phpmyadmin"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock definitions ---

// mockFS is a simple fake implementation of phpmyadmin.FileSystem.
// WriteFile stores content in the written map. Create uses a real temp file
// and stores the created path in the created set for verification.
type mockFS struct {
	statErr     error
	written     map[string][]byte
	created     map[string]bool
	removeAllFn func(path string) error
	removeFn    func(name string) error
}

func newMockFS() *mockFS {
	return &mockFS{
		written: make(map[string][]byte),
		created: make(map[string]bool),
	}
}

func (m *mockFS) Stat(name string) (fs.FileInfo, error) {
	return nil, m.statErr
}

func (m *mockFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.written[name] = append([]byte(nil), data...)
	return nil
}

func (m *mockFS) MkdirAll(path string, perm fs.FileMode) error {
	return nil
}

func (m *mockFS) RemoveAll(path string) error {
	if m.removeAllFn != nil {
		return m.removeAllFn(path)
	}
	return nil
}

func (m *mockFS) Remove(name string) error {
	if m.removeFn != nil {
		return m.removeFn(name)
	}
	return nil
}

func (m *mockFS) Create(name string) (*os.File, error) {
	// Create a real temp file so template.Execute can write to it.
	// We record the path was created; content can be read back from disk.
	f, err := os.CreateTemp("", "phpmyadmin-vhost-*")
	if err != nil {
		return nil, err
	}
	m.created[name] = true
	return f, nil
}

// --- realFS is used by TestVhostContents to capture vhost content from disk ---

// realFS writes to real paths under t.TempDir() and records written content.
type realFS struct {
	written  map[string][]byte
	created  map[string]string // logical path -> actual tmp file path
}

func newRealFS() *realFS {
	return &realFS{
		written: make(map[string][]byte),
		created: make(map[string]string),
	}
}

func (r *realFS) Stat(name string) (fs.FileInfo, error)              { return os.Stat(name) }
func (r *realFS) MkdirAll(path string, perm fs.FileMode) error       { return os.MkdirAll(path, perm) }
func (r *realFS) RemoveAll(path string) error                         { return os.RemoveAll(path) }
func (r *realFS) Remove(name string) error                            { return os.Remove(name) }
func (r *realFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	r.written[name] = append([]byte(nil), data...)
	return nil
}
func (r *realFS) Create(name string) (*os.File, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	r.created[name] = name
	return f, nil
}

// createdContent reads the content of a file created via Create().
func (r *realFS) createdContent(name string) ([]byte, error) {
	path, ok := r.created[name]
	if !ok {
		return nil, errors.New("path not in created map")
	}
	return os.ReadFile(path)
}

// --- Mock cert manager ---

type mockCertManager struct {
	generateCertFn func(domain string) (string, string, error)
	removeCertFn   func(domain string) error
	generateCalled string
	removeCalled   string
}

func (m *mockCertManager) GenerateCert(domain string) (string, string, error) {
	m.generateCalled = domain
	if m.generateCertFn != nil {
		return m.generateCertFn(domain)
	}
	return "/tmp/cert.pem", "/tmp/key.pem", nil
}

func (m *mockCertManager) RemoveCert(domain string) error {
	m.removeCalled = domain
	if m.removeCertFn != nil {
		return m.removeCertFn(domain)
	}
	return nil
}

// --- Mock nginx reloader ---

type mockNginxReloader struct {
	reloadCalled bool
	reloadErr    error
}

func (m *mockNginxReloader) Reload() error {
	m.reloadCalled = true
	return m.reloadErr
}

// --- Mock command runner ---

type mockCommandRunner struct {
	calls [][]string
	runFn func(name string, args ...string) (string, error)
}

func (m *mockCommandRunner) Run(name string, args ...string) (string, error) {
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)
	if m.runFn != nil {
		return m.runFn(name, args...)
	}
	return "", nil
}

// --- Test helpers ---

func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

func newTestManager(t *testing.T) (*phpmyadmin.Manager, *mockFS, *mockCertManager, *mockNginxReloader, *mockCommandRunner, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mfs := newMockFS()
	mcert := &mockCertManager{}
	mnginx := &mockNginxReloader{}
	mcmd := &mockCommandRunner{}
	mgr := phpmyadmin.NewManager(paths, mcert, mnginx, mfs, mcmd)
	return mgr, mfs, mcert, mnginx, mcmd, paths
}

// --- Tests ---

func TestInstall_Success(t *testing.T) {
	mgr, mfs, mcert, mnginx, mcmd, paths := newTestManager(t)

	err := mgr.Install()
	require.NoError(t, err)

	// cmd.Run should be called 4 times: download, extract, move, cleanup
	assert.Len(t, mcmd.calls, 4)

	// First call is download (sh -c curl ...)
	assert.Equal(t, "sh", mcmd.calls[0][0])
	assert.Contains(t, mcmd.calls[0][2], "curl")
	assert.Contains(t, mcmd.calls[0][2], "files.phpmyadmin.net")

	// Second call is extract (sh -c unzip ...)
	assert.Equal(t, "sh", mcmd.calls[1][0])
	assert.Contains(t, mcmd.calls[1][2], "unzip")

	// Third call is move (sh -c mv ... && rm -rf ...)
	assert.Equal(t, "sh", mcmd.calls[2][0])
	assert.Contains(t, mcmd.calls[2][2], "mv")

	// Fourth call is cleanup (rm -f <zipPath>)
	assert.Equal(t, "rm", mcmd.calls[3][0])

	// WriteFile should be called for config.inc.php
	configPath := paths.PhpMyAdminConf()
	_, ok := mfs.written[configPath]
	assert.True(t, ok, "config.inc.php should have been written")

	// certMgr.GenerateCert should be called for phpmyadmin.test
	assert.Equal(t, "phpmyadmin.test", mcert.generateCalled)

	// A vhost file should have been created
	vhostPath := filepath.Join(paths.NginxSitesDir(), "phpmyadmin.conf")
	assert.True(t, mfs.created[vhostPath], "vhost file should have been created")

	// nginxRld.Reload should have been called
	assert.True(t, mnginx.reloadCalled)
}

func TestInstall_DownloadFails(t *testing.T) {
	mgr, _, _, _, mcmd, _ := newTestManager(t)

	mcmd.runFn = func(name string, args ...string) (string, error) {
		return "", errors.New("network error")
	}

	err := mgr.Install()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download")
}

func TestUninstall_Success(t *testing.T) {
	mgr, mfs, mcert, mnginx, _, paths := newTestManager(t)

	var removedAllPaths []string
	var removedPaths []string

	mfs.removeAllFn = func(path string) error {
		removedAllPaths = append(removedAllPaths, path)
		return nil
	}
	mfs.removeFn = func(name string) error {
		removedPaths = append(removedPaths, name)
		return nil
	}

	err := mgr.Uninstall()
	require.NoError(t, err)

	// RemoveAll called with PhpMyAdminDir
	assert.Contains(t, removedAllPaths, paths.PhpMyAdminDir())

	// Remove called with phpmyadmin.conf
	vhostPath := filepath.Join(paths.NginxSitesDir(), "phpmyadmin.conf")
	assert.Contains(t, removedPaths, vhostPath)

	// certMgr.RemoveCert called with phpmyadmin.test
	assert.Equal(t, "phpmyadmin.test", mcert.removeCalled)

	// Nginx reloaded
	assert.True(t, mnginx.reloadCalled)
}

func TestIsInstalled_True(t *testing.T) {
	mgr, mfs, _, _, _, _ := newTestManager(t)

	// Stat returns nil error — file exists
	mfs.statErr = nil

	assert.True(t, mgr.IsInstalled())
}

func TestIsInstalled_False(t *testing.T) {
	mgr, mfs, _, _, _, _ := newTestManager(t)

	// Stat returns error — file not found
	mfs.statErr = os.ErrNotExist

	assert.False(t, mgr.IsInstalled())
}

func TestConfigContents(t *testing.T) {
	mgr, mfs, _, _, _, paths := newTestManager(t)

	err := mgr.Install()
	require.NoError(t, err)

	configPath := paths.PhpMyAdminConf()
	data, ok := mfs.written[configPath]
	require.True(t, ok, "config.inc.php was not written")

	content := string(data)
	assert.Contains(t, content, "auth_type")
	assert.Contains(t, content, "'config'")
	assert.Contains(t, content, "'root'")
	assert.Contains(t, content, paths.MySQLSocket())
	assert.Contains(t, content, "AllowNoPassword")

	// Verify blowfish_secret is exactly 32 chars
	idx := strings.Index(content, "blowfish_secret'] = '")
	require.True(t, idx >= 0, "blowfish_secret not found in config")
	start := idx + len("blowfish_secret'] = '")
	secret := content[start : start+32]
	assert.Len(t, secret, 32)
	// The next char should be the closing quote
	assert.Equal(t, "'", string(content[start+32]))
}

func TestVhostContents(t *testing.T) {
	// Use a realFS to capture actual vhost content written to disk.
	paths := &system.Paths{Home: t.TempDir()}
	require.NoError(t, os.MkdirAll(paths.NginxSitesDir(), 0755))
	require.NoError(t, os.MkdirAll(paths.PhpMyAdminDir(), 0755))

	rfs := newRealFS()
	mcert := &mockCertManager{}
	mnginx := &mockNginxReloader{}
	mcmd := &mockCommandRunner{}

	mgr := phpmyadmin.NewManager(paths, mcert, mnginx, rfs, mcmd)
	err := mgr.Install()
	require.NoError(t, err)

	vhostPath := filepath.Join(paths.NginxSitesDir(), "phpmyadmin.conf")
	content, err := rfs.createdContent(vhostPath)
	require.NoError(t, err)

	vhost := string(content)
	assert.Contains(t, vhost, "listen 127.0.0.1:443 ssl")
	assert.Contains(t, vhost, "server_name phpmyadmin.test")
	assert.Contains(t, vhost, "php-fpm.sock")
	assert.Contains(t, vhost, "libraries|setup|doc|locale")
}

func TestStatus(t *testing.T) {
	mgr, mfs, _, _, _, _ := newTestManager(t)

	// When index.php exists
	mfs.statErr = nil
	status := mgr.Status()
	assert.True(t, status.Installed)

	// When index.php does not exist
	mfs.statErr = os.ErrNotExist
	status = mgr.Status()
	assert.False(t, status.Installed)
}

