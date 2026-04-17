package integration_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestStartupSequence_DirectoryCreation verifies that EnsureDirectories creates all
// required subdirectories under the temp home with correct permissions (INTG-01).
func TestStartupSequence_DirectoryCreation(t *testing.T) {
	paths := newTestPaths(t)

	err := paths.EnsureDirectories()
	require.NoError(t, err, "EnsureDirectories should not return an error")

	expectedDirs := []string{
		paths.Home,
		paths.BinDir(),
		paths.PhpDir(),
		paths.NodeDir(),
		paths.NginxDir(),
		paths.NginxSitesDir(),
		paths.NginxLogsDir(),
		paths.CertsDir(),
		paths.LogsDir(),
		paths.ServicesDir(),
		paths.DnsmasqDir(),
		paths.DefaultSiteDir(),
	}

	for _, dir := range expectedDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			info, err := os.Stat(dir)
			require.NoError(t, err, "directory %s should exist", dir)
			assert.True(t, info.IsDir(), "%s should be a directory", dir)
			// Check permissions (mask to lower 9 bits)
			assert.Equal(t, os.FileMode(0755), info.Mode().Perm(), "%s should have 0755 permissions", dir)
		})
	}
}

// TestStartupSequence_CASetup verifies that SetupCA generates a CA cert and key,
// IsCAInstalled reports correctly, and SetupCA is idempotent (INTG-01).
func TestStartupSequence_CASetup(t *testing.T) {
	paths := newTestPaths(t)
	certMgr, _ := newTestCertManager(t, paths)

	// Before setup
	assert.False(t, certMgr.IsCAInstalled(), "CA should not be installed before SetupCA")

	// First call creates the CA
	err := certMgr.SetupCA()
	require.NoError(t, err, "SetupCA should succeed")
	assert.True(t, certMgr.IsCAInstalled(), "CA should be installed after SetupCA")

	// CA cert file should contain PEM data
	certData, err := os.ReadFile(paths.CACert())
	require.NoError(t, err, "CA cert file should be readable")
	assert.True(t, bytes.Contains(certData, []byte("-----BEGIN CERTIFICATE-----")),
		"CA cert should contain PEM CERTIFICATE block")

	// CA key file should contain PEM data
	keyData, err := os.ReadFile(paths.CAKey())
	require.NoError(t, err, "CA key file should be readable")
	assert.True(t, bytes.Contains(keyData, []byte("-----BEGIN RSA PRIVATE KEY-----")),
		"CA key should contain PEM RSA PRIVATE KEY block")

	// Record file modification times before second call
	certInfo1, _ := os.Stat(paths.CACert())
	keyInfo1, _ := os.Stat(paths.CAKey())

	// Second call is idempotent — no error, files unchanged
	err = certMgr.SetupCA()
	require.NoError(t, err, "second SetupCA call should be idempotent (no error)")

	certInfo2, _ := os.Stat(paths.CACert())
	keyInfo2, _ := os.Stat(paths.CAKey())

	assert.Equal(t, certInfo1.ModTime(), certInfo2.ModTime(),
		"CA cert file should not be modified on second SetupCA call")
	assert.Equal(t, keyInfo1.ModTime(), keyInfo2.ModTime(),
		"CA key file should not be modified on second SetupCA call")
}

// TestStartupSequence_CASetupWithTrust verifies that TrustCA delegates to AdminRunner
// with a command containing "security add-trusted-cert" (INTG-01).
func TestStartupSequence_CASetupWithTrust(t *testing.T) {
	paths := newTestPaths(t)
	certMgr, adminMock := newTestCertManager(t, paths)

	// Set up mock to capture the command
	adminMock.On("RunWithPrivileges", mock.MatchedBy(func(cmd string) bool {
		return strings.Contains(cmd, "security add-trusted-cert")
	})).Return(nil)

	require.NoError(t, certMgr.SetupCA())
	require.NoError(t, certMgr.TrustCA())

	adminMock.AssertExpectations(t)
}

// TestStartupSequence_SymlinkRestore_PHP verifies that LinkPhpVersion creates the
// correct symlinks in BinDir for the active PHP version (INTG-01).
func TestStartupSequence_SymlinkRestore_PHP(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)
	shell := newTestShell(t, paths)

	// Create fake PHP 8.3 directory structure
	phpVersionDir := paths.PhpVersionDir("8.3")
	phpBin := filepath.Join(phpVersionDir, "bin", "php")
	fpmBin := filepath.Join(phpVersionDir, "sbin", "php-fpm")

	require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(fpmBin), 0755))
	require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh\necho PHP/8.3.0"), 0755))
	require.NoError(t, os.WriteFile(fpmBin, []byte("#!/bin/sh\necho php-fpm"), 0755))

	// Set active version in config
	require.NoError(t, store.SetActivePhpVersion("8.3"))

	// Create PHP manager with RealTestFS and MockCommandRunner
	rfs := newRealTestFS(paths.Home)
	cmdMock := &MockCommandRunner{}

	// LookPath fails so only manual detection fires
	cmdMock.On("LookPath", mock.Anything).Return("", errors.New("not found"))
	// Run for version detection on the php binary returns "8.3.0"
	cmdMock.On("Run", phpBin, "-r", "echo PHP_VERSION;").Return("8.3.0", nil)

	mgr := php.NewManager(paths, store, rfs, cmdMock)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.NotEmpty(t, versions, "should detect at least one PHP version")

	// Find 8.3
	var target *php.PhpVersion
	for _, v := range versions {
		if v.Version == "8.3.0" || v.Version == "8.3" {
			v := v // capture range variable
			target = &v
			break
		}
	}
	require.NotNil(t, target, "PHP 8.3 should be detected")

	// Link the version
	err = shell.LinkPhpVersion(target.Binary, target.FpmBin, target.Path)
	require.NoError(t, err)

	// Assert symlinks exist in BinDir
	phpLink := filepath.Join(paths.BinDir(), "php")
	fpmLink := filepath.Join(paths.BinDir(), "php-fpm")

	phpTarget, err := os.Readlink(phpLink)
	require.NoError(t, err, "php symlink should exist in BinDir")
	assert.Equal(t, target.Binary, phpTarget, "php symlink should point to 8.3 binary")

	fpmTarget, err := os.Readlink(fpmLink)
	require.NoError(t, err, "php-fpm symlink should exist in BinDir")
	assert.Equal(t, target.FpmBin, fpmTarget, "php-fpm symlink should point to 8.3 fpm binary")
}

// TestStartupSequence_SymlinkRestore_Node verifies that LinkNodeBinaries creates the
// correct symlinks in BinDir for a given Node version directory (INTG-01).
func TestStartupSequence_SymlinkRestore_Node(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	shell := newTestShell(t, paths)

	// Create fake Node 20.0.0 directory
	nodeVersionDir := paths.NodeVersionDir("20.0.0")
	nodeBin := filepath.Join(nodeVersionDir, "bin", "node")
	npmBin := filepath.Join(nodeVersionDir, "bin", "npm")

	require.NoError(t, os.MkdirAll(filepath.Join(nodeVersionDir, "bin"), 0755))
	require.NoError(t, os.WriteFile(nodeBin, []byte("#!/bin/sh\necho node"), 0755))
	require.NoError(t, os.WriteFile(npmBin, []byte("#!/bin/sh\necho npm"), 0755))

	err := shell.LinkNodeBinaries(nodeVersionDir)
	require.NoError(t, err)

	// Assert symlinks exist in BinDir
	nodeLink := filepath.Join(paths.BinDir(), "node")
	npmLink := filepath.Join(paths.BinDir(), "npm")

	nodeTarget, err := os.Readlink(nodeLink)
	require.NoError(t, err, "node symlink should exist in BinDir")
	assert.Equal(t, nodeBin, nodeTarget, "node symlink should point to 20.0.0 binary")

	npmTarget, err := os.Readlink(npmLink)
	require.NoError(t, err, "npm symlink should exist in BinDir")
	assert.Equal(t, npmBin, npmTarget, "npm symlink should point to 20.0.0 npm")
}

// TestStartupSequence_ShellIntegration verifies that Shell.Install adds the LamboServer
// PATH block to an existing .zshrc and IsInstalled reports true afterwards (INTG-01).
func TestStartupSequence_ShellIntegration(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	// Create a Shell with WithHome pointing to a temp dir that has a .zshrc
	shellHome := t.TempDir()
	zshrc := filepath.Join(shellHome, ".zshrc")
	require.NoError(t, os.WriteFile(zshrc, []byte("# existing zshrc content\n"), 0644))

	shell := system.NewIntegration(paths).WithHome(shellHome)

	// Should not be installed yet
	assert.False(t, shell.IsInstalled(), "shell integration should not be installed before Install()")

	err := shell.Install()
	require.NoError(t, err, "Install should succeed")
	assert.True(t, shell.IsInstalled(), "shell integration should be installed after Install()")

	// .zshrc should contain the LamboServer PATH block
	zshrcContent, err := os.ReadFile(zshrc)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(zshrcContent), "# Added by LamboServer"),
		".zshrc should contain the LamboServer marker comment")
	assert.True(t, strings.Contains(string(zshrcContent), ".lamboserver/bin"),
		".zshrc should contain the LamboServer PATH export")
}
