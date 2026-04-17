package sites_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/sites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Link tests ---

func TestManager_Link_Success(t *testing.T) {
	mgr, siteFS, paths, store := newTestSiteManager(t)

	projectPath := t.TempDir()

	// Project path exists (stat returns nil error with a real dir info)
	info, err := os.Stat(projectPath)
	require.NoError(t, err)

	// No public/ subdir
	publicDir := filepath.Join(projectPath, "public")

	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", publicDir).Return(nil, errors.New("not found"))

	err = mgr.Link(projectPath, "myapp")
	require.NoError(t, err)

	// Check store has the site with auto-appended .test suffix
	sites := store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "myapp.test", sites[0].Domain)
	assert.Equal(t, projectPath, sites[0].Path)

	// Check cert files were generated
	certPath := paths.SiteCert("myapp.test")
	keyPath := paths.SiteKey("myapp.test")
	_, err = os.Stat(certPath)
	assert.NoError(t, err, "site cert file should exist")
	_, err = os.Stat(keyPath)
	assert.NoError(t, err, "site key file should exist")

	// Check the nginx conf content
	expectedConfPath := filepath.Join(paths.NginxSitesDir(), "myapp.test.conf")
	realPath := siteFS.CreatedFilePath(expectedConfPath)
	require.NotEmpty(t, realPath, "nginx conf should have been created")

	content, err := os.ReadFile(realPath)
	require.NoError(t, err)

	confStr := string(content)
	assert.Contains(t, confStr, "server_name myapp.test;")
	assert.Contains(t, confStr, "root "+projectPath+";")
	assert.Contains(t, confStr, "ssl_certificate ")
	assert.Contains(t, confStr, "fastcgi_pass unix:")
	assert.Contains(t, confStr, "php-fpm.sock")
}

func TestManager_Link_LaravelDetection(t *testing.T) {
	mgr, siteFS, _, store := newTestSiteManager(t)

	projectPath := t.TempDir()
	publicDir := filepath.Join(projectPath, "public")
	require.NoError(t, os.MkdirAll(publicDir, 0755))

	projectInfo, err := os.Stat(projectPath)
	require.NoError(t, err)
	publicInfo, err := os.Stat(publicDir)
	require.NoError(t, err)

	siteFS.On("Stat", projectPath).Return(projectInfo, nil)
	siteFS.On("Stat", publicDir).Return(publicInfo, nil)

	err = mgr.Link(projectPath, "laravel")
	require.NoError(t, err)

	sites := store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "laravel.test", sites[0].Domain)

	// Read config and assert root points to public/
	confPath := filepath.Join(siteFS.tmpDir, "laravel.test.conf")
	content, err := os.ReadFile(confPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "root "+publicDir+";")
}

func TestManager_Link_AppendsTestSuffix(t *testing.T) {
	mgr, siteFS, _, store := newTestSiteManager(t)

	projectPath := t.TempDir()
	info, err := os.Stat(projectPath)
	require.NoError(t, err)

	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", filepath.Join(projectPath, "public")).Return(nil, errors.New("not found"))

	err = mgr.Link(projectPath, "myapp")
	require.NoError(t, err)

	sites := store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "myapp.test", sites[0].Domain)
}

func TestManager_Link_AlreadyHasTestSuffix(t *testing.T) {
	mgr, siteFS, _, store := newTestSiteManager(t)

	projectPath := t.TempDir()
	info, err := os.Stat(projectPath)
	require.NoError(t, err)

	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", filepath.Join(projectPath, "public")).Return(nil, errors.New("not found"))

	err = mgr.Link(projectPath, "myapp.test")
	require.NoError(t, err)

	sites := store.GetSites()
	require.Len(t, sites, 1)
	// Should NOT be "myapp.test.test"
	assert.Equal(t, "myapp.test", sites[0].Domain)
}

func TestManager_Link_PathNotExist(t *testing.T) {
	mgr, siteFS, _, _ := newTestSiteManager(t)

	projectPath := "/nonexistent/path/that/does/not/exist"
	siteFS.On("Stat", projectPath).Return(nil, os.ErrNotExist)

	err := mgr.Link(projectPath, "myapp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")
}

func TestManager_Link_SetupCAIfNeeded(t *testing.T) {
	// Create a fresh site manager whose cert.Manager has NO CA pre-setup.
	// site.Manager.Link should call SetupCA and TrustCA automatically.
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())
	store := newTestStore(t, paths)

	adminMock := &MockAdminRunner{}
	// Accept any RunWithPrivileges call (TrustCA will call it)
	adminMock.On("RunWithPrivileges", mock.Anything).Return(nil)

	certMgr := cert.NewManager(paths, &realCertFS{}, adminMock)
	// Do NOT call SetupCA here — Link should trigger it automatically

	siteFS := newMockSiteFS(t)
	mgr := sites.NewManager(paths, store, certMgr, siteFS)

	projectPath := t.TempDir()
	info, err := os.Stat(projectPath)
	require.NoError(t, err)
	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", filepath.Join(projectPath, "public")).Return(nil, errors.New("not found"))

	// Link should succeed: SetupCA and TrustCA called internally
	err = mgr.Link(projectPath, "catest")
	require.NoError(t, err, "Link should succeed by auto-setting up CA")

	// CA should now be installed
	assert.True(t, certMgr.IsCAInstalled())
	adminMock.AssertExpectations(t)
}

// --- Unlink tests ---

func TestManager_Unlink_Success(t *testing.T) {
	mgr, siteFS, paths, store := newTestSiteManager(t)

	projectPath := t.TempDir()
	info, err := os.Stat(projectPath)
	require.NoError(t, err)

	// Setup for Link
	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", filepath.Join(projectPath, "public")).Return(nil, errors.New("not found"))

	// Link the site first
	require.NoError(t, mgr.Link(projectPath, "myapp"))

	// Verify site was added
	sites := store.GetSites()
	require.Len(t, sites, 1)

	// Setup Remove mock for unlink
	confPath := filepath.Join(paths.NginxSitesDir(), "myapp.test.conf")
	siteFS.On("Remove", confPath).Return(nil)

	// Unlink
	err = mgr.Unlink("myapp")
	require.NoError(t, err)

	// Store no longer has the site
	sites = store.GetSites()
	assert.Empty(t, sites)

	// Nginx conf remove was called
	siteFS.AssertCalled(t, "Remove", confPath)

	// Cert files should be removed
	_, err = os.Stat(paths.SiteCert("myapp.test"))
	assert.True(t, os.IsNotExist(err), "cert file should be removed after unlink")
	_, err = os.Stat(paths.SiteKey("myapp.test"))
	assert.True(t, os.IsNotExist(err), "key file should be removed after unlink")
}

func TestManager_Unlink_AppendsTestSuffix(t *testing.T) {
	mgr, siteFS, paths, store := newTestSiteManager(t)

	projectPath := t.TempDir()
	info, err := os.Stat(projectPath)
	require.NoError(t, err)

	siteFS.On("Stat", projectPath).Return(info, nil)
	siteFS.On("Stat", filepath.Join(projectPath, "public")).Return(nil, errors.New("not found"))

	require.NoError(t, mgr.Link(projectPath, "myapp"))

	confPath := filepath.Join(paths.NginxSitesDir(), "myapp.test.conf")
	siteFS.On("Remove", confPath).Return(nil)

	// Call Unlink without .test suffix
	err = mgr.Unlink("myapp")
	require.NoError(t, err)

	sites := store.GetSites()
	assert.Empty(t, sites, "site should be removed from store")
}

// --- List tests ---

func TestManager_List_Empty(t *testing.T) {
	mgr, _, _, _ := newTestSiteManager(t)

	sites := mgr.List()
	assert.Empty(t, sites)
}

func TestManager_List_ReturnsSites(t *testing.T) {
	mgr, siteFS, _, store := newTestSiteManager(t)

	// Add sites via Link
	projectPath1 := t.TempDir()
	info1, _ := os.Stat(projectPath1)
	siteFS.On("Stat", projectPath1).Return(info1, nil)
	siteFS.On("Stat", filepath.Join(projectPath1, "public")).Return(nil, errors.New("not found"))
	require.NoError(t, mgr.Link(projectPath1, "siteone"))

	projectPath2 := t.TempDir()
	info2, _ := os.Stat(projectPath2)
	siteFS.On("Stat", projectPath2).Return(info2, nil)
	siteFS.On("Stat", filepath.Join(projectPath2, "public")).Return(nil, errors.New("not found"))
	require.NoError(t, mgr.Link(projectPath2, "sitetwo"))

	_ = store

	sites := mgr.List()
	require.Len(t, sites, 2)

	domains := []string{sites[0].Domain, sites[1].Domain}
	assert.Contains(t, domains, "siteone.test")
	assert.Contains(t, domains, "sitetwo.test")

	// Verify paths are correct
	for _, s := range sites {
		if s.Domain == "siteone.test" {
			assert.Equal(t, projectPath1, s.Path)
		}
		if s.Domain == "sitetwo.test" {
			assert.Equal(t, projectPath2, s.Path)
		}
	}
}

// verifyNginxConfContent is a helper that reads a conf file written by MockSiteFS.Create
// and asserts expected substrings.
func verifyNginxConfContent(t *testing.T, siteFS *MockSiteFS, confPath string, expected ...string) {
	t.Helper()
	realPath := siteFS.CreatedFilePath(confPath)
	require.NotEmpty(t, realPath, "conf file should have been created at %s", confPath)

	content, err := os.ReadFile(realPath)
	require.NoError(t, err)

	confStr := string(content)
	for _, substr := range expected {
		assert.True(t, strings.Contains(confStr, substr),
			"nginx conf should contain %q\n\nActual content:\n%s", substr, confStr)
	}
}
