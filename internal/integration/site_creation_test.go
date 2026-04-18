// Package integration_test contains integration tests that exercise multiple
// managers working together through real filesystem operations.
package integration_test

import (
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/dnsmasq"
	"github.com/rizkirmdhnnn/lamboserver/internal/sites"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Real filesystem adapters ---

// realCertFS implements cert.FileSystem using real os.* calls.
type realCertFS struct{}

func (r realCertFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r realCertFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r realCertFS) Create(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.Create(name)
}

func (r realCertFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(name, flag, perm)
}

func (r realCertFS) Remove(name string) error {
	return os.Remove(name)
}

// realSiteFS implements sites.FileSystem using real os.* calls.
type realSiteFS struct{}

func (r realSiteFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r realSiteFS) Create(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.Create(name)
}

func (r realSiteFS) Remove(name string) error {
	return os.Remove(name)
}

// --- Mock admin runner ---

// mockAdminRunner implements cert.AdminRunner and records calls.
type mockAdminRunner struct {
	calls []string
}

func (m *mockAdminRunner) RunWithPrivileges(command string) error {
	m.calls = append(m.calls, command)
	return nil
}

// --- Test helpers ---

// siteTestEnv holds the test managers and supporting state.
type siteTestEnv struct {
	siteManager *sites.Manager
	certManager *cert.Manager
	paths       *system.Paths
	store       *config.Store
	adminRunner *mockAdminRunner
}

// newTestSiteManagerWithCA creates all managers and pre-installs the CA.
func newTestSiteManagerWithCA(t *testing.T) *siteTestEnv {
	t.Helper()
	paths := &system.Paths{Home: filepath.Join(t.TempDir(), ".lamboserver")}
	require.NoError(t, paths.EnsureDirectories())

	store := config.NewStore(paths.ConfigFile())
	adminRunner := &mockAdminRunner{}
	certMgr := cert.NewManager(paths, realCertFS{}, adminRunner)
	require.NoError(t, certMgr.SetupCA())

	siteMgr := sites.NewManager(paths, store, certMgr, realSiteFS{})
	return &siteTestEnv{
		siteManager: siteMgr,
		certManager: certMgr,
		paths:       paths,
		store:       store,
		adminRunner: adminRunner,
	}
}

// newTestSiteManagerWithoutCA creates all managers WITHOUT pre-installing the CA.
func newTestSiteManagerWithoutCA(t *testing.T) *siteTestEnv {
	t.Helper()
	paths := &system.Paths{Home: filepath.Join(t.TempDir(), ".lamboserver")}
	require.NoError(t, paths.EnsureDirectories())

	store := config.NewStore(paths.ConfigFile())
	adminRunner := &mockAdminRunner{}
	certMgr := cert.NewManager(paths, realCertFS{}, adminRunner)
	// NOTE: intentionally NOT calling certMgr.SetupCA() here

	siteMgr := sites.NewManager(paths, store, certMgr, realSiteFS{})
	return &siteTestEnv{
		siteManager: siteMgr,
		certManager: certMgr,
		paths:       paths,
		store:       store,
		adminRunner: adminRunner,
	}
}

// readFileContent is a helper to read file content as string.
func readFileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	return string(data)
}

// --- Tests ---

func TestSiteCreation_NginxConfigGenerated(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	err := env.siteManager.Link(projectPath, "myapp")
	require.NoError(t, err)

	confPath := filepath.Join(env.paths.NginxSitesDir(), "myapp.test.conf")
	require.FileExists(t, confPath)

	content := readFileContent(t, confPath)
	assert.Contains(t, content, "server_name myapp.test")
	assert.Contains(t, content, "root "+projectPath)
	assert.Contains(t, content, "ssl_certificate "+env.paths.SiteCert("myapp.test"))
	assert.Contains(t, content, "ssl_certificate_key "+env.paths.SiteKey("myapp.test"))
	assert.Contains(t, content, "fastcgi_pass unix:"+filepath.Join(env.paths.Home, "php-fpm.sock"))
	assert.Contains(t, content, "include "+env.paths.NginxFastCGIParams())
}

func TestSiteCreation_SSLCertGenerated(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	err := env.siteManager.Link(projectPath, "secure.test")
	require.NoError(t, err)

	certPath := env.paths.SiteCert("secure.test")
	keyPath := env.paths.SiteKey("secure.test")

	// Cert file must exist and contain PEM cert header
	require.FileExists(t, certPath)
	certContent := readFileContent(t, certPath)
	assert.Contains(t, certContent, "BEGIN CERTIFICATE")

	// Key file must exist and contain PEM key header
	require.FileExists(t, keyPath)
	keyContent := readFileContent(t, keyPath)
	assert.True(t,
		strings.Contains(keyContent, "BEGIN RSA PRIVATE KEY") ||
			strings.Contains(keyContent, "BEGIN EC PRIVATE KEY"),
		"expected PEM private key header, got: %s", keyContent[:min(len(keyContent), 64)],
	)

	// Parse the certificate and verify its fields
	certPEMData, err := os.ReadFile(certPath)
	require.NoError(t, err)
	certBlock, _ := pem.Decode(certPEMData)
	require.NotNil(t, certBlock, "failed to PEM-decode site cert")
	siteCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	assert.Equal(t, "secure.test", siteCert.Subject.CommonName)
	assert.Contains(t, siteCert.DNSNames, "secure.test")

	// Verify the cert is signed by our CA
	caPEMData, err := os.ReadFile(env.paths.CACert())
	require.NoError(t, err)
	caBlock, _ := pem.Decode(caPEMData)
	require.NotNil(t, caBlock, "failed to PEM-decode CA cert")
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)

	err = siteCert.CheckSignatureFrom(caCert)
	assert.NoError(t, err, "site cert must be signed by local CA")
}

func TestSiteCreation_ConfigStorePersisted(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	err := env.siteManager.Link(projectPath, "stored")
	require.NoError(t, err)

	sites := env.store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "stored.test", sites[0].Domain)
	assert.Equal(t, projectPath, sites[0].Path)
}

func TestSiteCreation_AutoAppendTestTLD(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	err := env.siteManager.Link(projectPath, "noext")
	require.NoError(t, err)

	// Nginx config must be created with .test suffix
	confPath := filepath.Join(env.paths.NginxSitesDir(), "noext.test.conf")
	require.FileExists(t, confPath)

	// Store must have .test suffix
	sites := env.store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "noext.test", sites[0].Domain)
}

func TestSiteCreation_LaravelPublicDetection(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	// Create a "public/" subdirectory to simulate a Laravel project
	publicDir := filepath.Join(projectPath, "public")
	require.NoError(t, os.MkdirAll(publicDir, 0755))

	err := env.siteManager.Link(projectPath, "laravelapp")
	require.NoError(t, err)

	confPath := filepath.Join(env.paths.NginxSitesDir(), "laravelapp.test.conf")
	require.FileExists(t, confPath)

	content := readFileContent(t, confPath)
	// Document root must point to the public/ subdirectory
	assert.Contains(t, content, "root "+publicDir)
	assert.NotContains(t, content, "root "+projectPath+"\n")
}

func TestSiteCreation_AutoSetupCA(t *testing.T) {
	// Build env WITHOUT pre-calling SetupCA
	env := newTestSiteManagerWithoutCA(t)
	projectPath := t.TempDir()

	// CA should not be installed yet
	assert.False(t, env.certManager.IsCAInstalled())

	// Link auto-triggers SetupCA + TrustCA
	err := env.siteManager.Link(projectPath, "autosetup")
	require.NoError(t, err)

	// CA cert and key must now exist
	require.FileExists(t, env.paths.CACert())
	require.FileExists(t, env.paths.CAKey())

	// AdminRunner must have been called (for TrustCA)
	require.NotEmpty(t, env.adminRunner.calls, "expected RunWithPrivileges to be called for TrustCA")
	assert.True(t, strings.Contains(env.adminRunner.calls[0], "security add-trusted-cert"),
		"expected security add-trusted-cert command, got: %s", env.adminRunner.calls[0])

	// The site SSL cert must still be generated correctly
	certPath := env.paths.SiteCert("autosetup.test")
	require.FileExists(t, certPath)
	certContent := readFileContent(t, certPath)
	assert.Contains(t, certContent, "BEGIN CERTIFICATE")
}

func TestSiteCreation_UnlinkCleansUp(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectPath := t.TempDir()

	// First create the site
	err := env.siteManager.Link(projectPath, "cleanup.test")
	require.NoError(t, err)

	confPath := filepath.Join(env.paths.NginxSitesDir(), "cleanup.test.conf")
	require.FileExists(t, confPath, "nginx config must exist before unlink")
	require.FileExists(t, env.paths.SiteCert("cleanup.test"), "cert must exist before unlink")
	require.FileExists(t, env.paths.SiteKey("cleanup.test"), "key must exist before unlink")

	// Now unlink
	err = env.siteManager.Unlink("cleanup.test")
	require.NoError(t, err)

	// Nginx config must be gone
	assert.NoFileExists(t, confPath, "nginx config must be removed after unlink")

	// Cert files must be gone
	assert.NoFileExists(t, env.paths.SiteCert("cleanup.test"), "cert must be removed after unlink")
	assert.NoFileExists(t, env.paths.SiteKey("cleanup.test"), "key must be removed after unlink")

	// Store must not contain the domain
	sites := env.store.GetSites()
	for _, s := range sites {
		assert.NotEqual(t, "cleanup.test", s.Domain, "unlinked domain must not remain in store")
	}
}

func TestSiteCreation_MultipleSites(t *testing.T) {
	env := newTestSiteManagerWithCA(t)
	projectA := t.TempDir()
	projectB := t.TempDir()

	// Link two sites
	require.NoError(t, env.siteManager.Link(projectA, "site-a"))
	require.NoError(t, env.siteManager.Link(projectB, "site-b"))

	confA := filepath.Join(env.paths.NginxSitesDir(), "site-a.test.conf")
	confB := filepath.Join(env.paths.NginxSitesDir(), "site-b.test.conf")

	// Both nginx configs must exist
	require.FileExists(t, confA)
	require.FileExists(t, confB)

	// Both cert pairs must exist
	require.FileExists(t, env.paths.SiteCert("site-a.test"))
	require.FileExists(t, env.paths.SiteKey("site-a.test"))
	require.FileExists(t, env.paths.SiteCert("site-b.test"))
	require.FileExists(t, env.paths.SiteKey("site-b.test"))

	// Store must have 2 entries
	sites := env.store.GetSites()
	require.Len(t, sites, 2)

	// Unlink site-a only
	require.NoError(t, env.siteManager.Unlink("site-a"))

	// site-a config must be gone
	assert.NoFileExists(t, confA, "site-a config must be removed after unlink")

	// site-b config must still exist
	require.FileExists(t, confB, "site-b config must persist after site-a is unlinked")
	require.FileExists(t, env.paths.SiteCert("site-b.test"), "site-b cert must persist")

	// Store must have 1 entry (site-b)
	sites = env.store.GetSites()
	require.Len(t, sites, 1)
	assert.Equal(t, "site-b.test", sites[0].Domain)
}

// dnsTestFS wraps RealTestFS and intercepts ReadFile for privileged system paths
// (e.g. /etc/resolver/test) to simulate them not existing in the test environment.
// This prevents EnsureResolver's early-exit optimisation from triggering when the
// real resolver file already has the expected content on the host machine.
type dnsTestFS struct {
	*RealTestFS
	blockedReads map[string]bool
}

func newDNSTestFS(root string, blockedPaths ...string) *dnsTestFS {
	blocked := make(map[string]bool, len(blockedPaths))
	for _, p := range blockedPaths {
		blocked[p] = true
	}
	return &dnsTestFS{RealTestFS: newRealTestFS(root), blockedReads: blocked}
}

func (d *dnsTestFS) ReadFile(name string) ([]byte, error) {
	if d.blockedReads[name] {
		return nil, os.ErrNotExist
	}
	return d.RealTestFS.ReadFile(name)
}

// TestSiteCreation_DNSResolverEntry verifies that dnsmasq.Manager.EnsureConfig and
// EnsureResolver produce the correct artifacts when called as part of the site
// creation workflow.
func TestSiteCreation_DNSResolverEntry(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	// Block reads of the real resolver file so EnsureResolver always proceeds past
	// its early-exit check, even if /etc/resolver/test already exists on this host.
	dnsFS := newDNSTestFS(paths.Home, paths.ResolverFile())
	mockLaunchd := &MockLaunchdService{}
	mockAdmin := &MockAdminRunner{}

	// Compile-time checks: all three mocks satisfy dns interfaces via structural typing.
	var _ dnsmasq.FileSystem = dnsFS
	var _ dnsmasq.AdminRunner = mockAdmin
	var _ dnsmasq.LaunchdService = mockLaunchd

	dnsMgr := dnsmasq.NewManager(paths, mockLaunchd, dnsFS, mockAdmin)

	t.Run("DnsmasqConfig", func(t *testing.T) {
		err := dnsMgr.EnsureConfig()
		require.NoError(t, err)

		content := readFileContent(t, paths.DnsmasqConf())
		assert.Contains(t, content, "address=/.test/127.0.0.1")
		assert.Contains(t, content, "listen-address=127.0.0.1")
		assert.Contains(t, content, "port=53")
	})

	t.Run("ResolverSetup", func(t *testing.T) {
		mockAdmin.On("RunWithPrivileges", mock.AnythingOfType("string")).Return(nil)

		err := dnsMgr.EnsureResolver()
		require.NoError(t, err)

		// Verify admin was called with the resolver setup command.
		mockAdmin.AssertCalled(t, "RunWithPrivileges", mock.AnythingOfType("string"))
		calls := mockAdmin.Calls
		require.NotEmpty(t, calls, "expected RunWithPrivileges to be called")
		cmd := calls[0].Arguments.String(0)
		assert.Contains(t, cmd, paths.ResolverDir())
		assert.Contains(t, cmd, paths.ResolverFile())

		// Verify the temp resolver file content.
		tmpFile := os.TempDir() + "/lambo-resolver"
		t.Cleanup(func() { os.Remove(tmpFile) })
		data, err := os.ReadFile(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 127.0.0.1\n", string(data))
	})
}

// TestSiteCreation_FullWorkflowWithDNS exercises the complete site creation
// workflow: DNS setup + site link. Verifies all four SC-3 artifacts are present
// after the combined workflow: Nginx config, SSL cert, SSL key, DNS resolver entry.
func TestSiteCreation_FullWorkflowWithDNS(t *testing.T) {
	// --- Setup shared infrastructure ---
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)
	certMgr, adminRunner := newTestCertManager(t, paths)
	adminRunner.On("RunWithPrivileges", mock.AnythingOfType("string")).Return(nil)

	// Setup CA before linking sites.
	require.NoError(t, certMgr.SetupCA())

	// Block reads of the real resolver file so EnsureResolver always proceeds past
	// its early-exit check, even if /etc/resolver/test already exists on this host.
	dnsFS := newDNSTestFS(paths.Home, paths.ResolverFile())
	mockLaunchd := &MockLaunchdService{}

	// Create a fake dnsmasq binary so IsInstalled() returns true.
	require.NoError(t, os.WriteFile(paths.DnsmasqBin(), []byte(""), 0755))

	// Ensure log dir exists so ensureLogFile in dns.Start succeeds.
	require.NoError(t, os.MkdirAll(paths.LogsDir(), 0755))

	// Mock launchd Install and IsRunning for dns.Start.
	mockLaunchd.On("Install", mock.Anything).Return(nil)
	mockLaunchd.On("IsRunning", mock.AnythingOfType("string")).Return(false)

	dnsMgr := dnsmasq.NewManager(paths, mockLaunchd, dnsFS, adminRunner)
	siteMgr := sites.NewManager(paths, store, certMgr, realSiteFS{})

	// --- Execute workflow ---

	// Step 1: Start DNS (EnsureConfig + launchd.Install).
	require.NoError(t, dnsMgr.Start())

	// Step 2: EnsureResolver (separate from Start — simulates full app DNS init).
	require.NoError(t, dnsMgr.EnsureResolver())
	tmpFile := os.TempDir() + "/lambo-resolver"
	t.Cleanup(func() { os.Remove(tmpFile) })

	// Step 3: Link a project site.
	projectPath := t.TempDir()
	require.NoError(t, siteMgr.Link(projectPath, "fulltest"))

	// --- Assert all four SC-3 artifact categories ---

	t.Run("NginxConfig", func(t *testing.T) {
		confPath := filepath.Join(paths.NginxSitesDir(), "fulltest.test.conf")
		require.FileExists(t, confPath)
		content := readFileContent(t, confPath)
		assert.Contains(t, content, "server_name fulltest.test")
	})

	t.Run("SSLCert", func(t *testing.T) {
		certPath := paths.SiteCert("fulltest.test")
		require.FileExists(t, certPath)
		content := readFileContent(t, certPath)
		assert.Contains(t, content, "BEGIN CERTIFICATE")
	})

	t.Run("SSLKey", func(t *testing.T) {
		keyPath := paths.SiteKey("fulltest.test")
		require.FileExists(t, keyPath)
		content := readFileContent(t, keyPath)
		assert.True(t,
			strings.Contains(content, "BEGIN RSA PRIVATE KEY") ||
				strings.Contains(content, "BEGIN EC PRIVATE KEY"),
			"expected PEM private key header",
		)
	})

	t.Run("DNSConfig", func(t *testing.T) {
		require.FileExists(t, paths.DnsmasqConf())
		content := readFileContent(t, paths.DnsmasqConf())
		assert.Contains(t, content, "address=/.test/127.0.0.1")

		// Verify resolver setup command was called with correct paths.
		var resolverCmd string
		for _, call := range adminRunner.Calls {
			cmd := call.Arguments.String(0)
			if strings.Contains(cmd, paths.ResolverDir()) {
				resolverCmd = cmd
				break
			}
		}
		require.NotEmpty(t, resolverCmd, "expected RunWithPrivileges to be called with resolver command")
		assert.Contains(t, resolverCmd, paths.ResolverFile())
	})
}

// min returns the smaller of two ints (Go 1.21 has a builtin, but 1.23 also has it).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
