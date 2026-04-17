package integration_test

import (
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

// createFakePhpVersion creates a fake PHP version directory with bin/php and sbin/php-fpm stubs.
func createFakePhpVersion(t *testing.T, paths *system.Paths, version string) (phpBin, fpmBin string) {
	t.Helper()
	versionDir := paths.PhpVersionDir(version)
	phpBin = filepath.Join(versionDir, "bin", "php")
	fpmBin = filepath.Join(versionDir, "sbin", "php-fpm")

	require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(fpmBin), 0755))
	require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh\necho PHP/"+version), 0755))
	require.NoError(t, os.WriteFile(fpmBin, []byte("#!/bin/sh\necho php-fpm"), 0755))
	return phpBin, fpmBin
}

// TestPhpVersionSwitch_SymlinkUpdate verifies that switching the active PHP version
// updates symlinks in BinDir and persists the new version in the config store (INTG-02).
func TestPhpVersionSwitch_SymlinkUpdate(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)
	shell := newTestShell(t, paths)

	// Create two fake PHP versions
	php82Bin, php82Fpm := createFakePhpVersion(t, paths, "8.2")
	php83Bin, php83Fpm := createFakePhpVersion(t, paths, "8.3")

	// Set initial active version to 8.2
	require.NoError(t, store.SetActivePhpVersion("8.2"))

	// Wire PHP manager
	rfs := newRealTestFS(paths.Home)
	cmdMock := &MockCommandRunner{}

	// LookPath fails — only manual detection fires
	cmdMock.On("LookPath", mock.Anything).Return("", errors.New("not found"))
	// Version command for 8.2 binary
	cmdMock.On("Run", php82Bin, "-r", "echo PHP_VERSION;").Return("8.2.0", nil)
	// Version command for 8.3 binary
	cmdMock.On("Run", php83Bin, "-r", "echo PHP_VERSION;").Return("8.3.0", nil)

	mgr := php.NewManager(paths, store, rfs, cmdMock)

	// Switch to 8.3: find the version, set active, link
	versions, err := mgr.ListInstalled()
	require.NoError(t, err)

	var target *php.PhpVersion
	for _, v := range versions {
		if strings.HasPrefix(v.Version, "8.3") {
			v := v
			target = &v
			break
		}
	}
	require.NotNil(t, target, "PHP 8.3 should be detected")

	require.NoError(t, mgr.SetActive(target.Version))
	require.NoError(t, shell.LinkPhpVersion(target.Binary, target.FpmBin, target.Path))

	// Config store should reflect the new active version
	assert.Equal(t, target.Version, store.Get().ActivePhpVersion,
		"config store should record the new active PHP version")

	// Symlinks should point to 8.3 binaries
	phpLink := filepath.Join(paths.BinDir(), "php")
	fpmLink := filepath.Join(paths.BinDir(), "php-fpm")

	phpLinkTarget, err := os.Readlink(phpLink)
	require.NoError(t, err, "php symlink should exist")
	assert.Equal(t, php83Bin, phpLinkTarget, "php symlink should point to 8.3 binary")

	fpmLinkTarget, err := os.Readlink(fpmLink)
	require.NoError(t, err, "php-fpm symlink should exist")
	assert.Equal(t, php83Fpm, fpmLinkTarget, "php-fpm symlink should point to 8.3 fpm binary")

	// Sanity: 8.2 binaries are untouched (only the symlink moved)
	_, err = os.Stat(php82Bin)
	assert.NoError(t, err, "8.2 php binary file should still exist after switch to 8.3")
	_, err = os.Stat(php82Fpm)
	assert.NoError(t, err, "8.2 php-fpm binary file should still exist after switch to 8.3")
}

// TestPhpVersionSwitch_ShellPathUpdate verifies the full chain: version switch plus
// shell integration writes the LamboServer PATH block to .zshrc (INTG-02).
func TestPhpVersionSwitch_ShellPathUpdate(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)

	// Shell with a writable temp home containing a .zshrc
	shellHome := t.TempDir()
	zshrc := filepath.Join(shellHome, ".zshrc")
	require.NoError(t, os.WriteFile(zshrc, []byte("# existing content\n"), 0644))
	shell := system.NewIntegration(paths).WithHome(shellHome)

	// Create two fake PHP versions
	_, _ = createFakePhpVersion(t, paths, "8.2")
	php83Bin, php83Fpm := createFakePhpVersion(t, paths, "8.3")

	require.NoError(t, store.SetActivePhpVersion("8.2"))

	rfs := newRealTestFS(paths.Home)
	cmdMock := &MockCommandRunner{}
	cmdMock.On("LookPath", mock.Anything).Return("", errors.New("not found"))

	php82Dir := paths.PhpVersionDir("8.2")
	php83Dir := paths.PhpVersionDir("8.3")
	cmdMock.On("Run", filepath.Join(php82Dir, "bin", "php"), "-r", "echo PHP_VERSION;").Return("8.2.0", nil)
	cmdMock.On("Run", filepath.Join(php83Dir, "bin", "php"), "-r", "echo PHP_VERSION;").Return("8.3.0", nil)

	mgr := php.NewManager(paths, store, rfs, cmdMock)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)

	var target *php.PhpVersion
	for _, v := range versions {
		if strings.HasPrefix(v.Version, "8.3") {
			v := v
			target = &v
			break
		}
	}
	require.NotNil(t, target)

	require.NoError(t, mgr.SetActive(target.Version))
	require.NoError(t, shell.LinkPhpVersion(php83Bin, php83Fpm, target.Path))

	// Mirror the startup behavior: install shell integration after version switch
	require.NoError(t, shell.Install())

	// .zshrc should now contain the LamboServer PATH block
	zshrcData, err := os.ReadFile(zshrc)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(zshrcData), "# Added by LamboServer"),
		".zshrc should contain the LamboServer marker")
	assert.True(t, strings.Contains(string(zshrcData), ".lamboserver/bin"),
		".zshrc should contain the LamboServer PATH export")
}

// TestPhpVersionSwitch_FpmRestart verifies that FpmManager.Restart calls Uninstall then
// Install on the LaunchdService, simulating the PHP-FPM restart triggered on version switch (INTG-02).
func TestPhpVersionSwitch_FpmRestart(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)

	// Create PHP 8.3 with the sbin/php-fpm binary so Start can find it
	php83Fpm := filepath.Join(paths.PhpVersionDir("8.3"), "sbin", "php-fpm")
	require.NoError(t, os.MkdirAll(filepath.Dir(php83Fpm), 0755))
	require.NoError(t, os.WriteFile(php83Fpm, []byte("#!/bin/sh\necho php-fpm"), 0755))

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	mockLaunchd := &MockLaunchdService{}
	rfs := newRealTestFS(paths.Home)

	// FPM manager
	fpmMgr := php.NewFpmManager(paths, store, mockLaunchd, rfs)

	// Mock Uninstall (called by Stop internally in Restart)
	mockLaunchd.On("Uninstall", mock.AnythingOfType("system.ServiceConfig")).Return(nil)
	// Mock Install (called by Start internally in Restart)
	mockLaunchd.On("Install", mock.AnythingOfType("system.ServiceConfig")).Return(nil)
	// IsRunning is not called in Restart itself; it's called by the caller (app_php.go)
	// so we just verify the Uninstall+Install sequence here.

	err := fpmMgr.Restart()
	require.NoError(t, err, "FpmManager.Restart should succeed")

	// Both Uninstall and Install must have been called (stop + start)
	mockLaunchd.AssertCalled(t, "Uninstall", mock.AnythingOfType("system.ServiceConfig"))
	mockLaunchd.AssertCalled(t, "Install", mock.AnythingOfType("system.ServiceConfig"))
}

// TestPhpVersionSwitch_VersionNotFound verifies that when the requested version is not
// installed, ListInstalled returns an empty list and no target is found — matching
// the error path in app_php.go SetActivePhp (INTG-02).
func TestPhpVersionSwitch_VersionNotFound(t *testing.T) {
	paths := newTestPaths(t)
	require.NoError(t, paths.EnsureDirectories())

	store := newTestStore(t, paths)

	// No PHP versions installed — empty PHP dir
	rfs := newRealTestFS(paths.Home)
	cmdMock := &MockCommandRunner{}

	// LookPath fails (no system PHP)
	cmdMock.On("LookPath", mock.Anything).Return("", errors.New("not found"))

	mgr := php.NewManager(paths, store, rfs, cmdMock)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err, "ListInstalled should not error even when no PHP is installed")

	// Try to find version "9.0" in the list
	var target *php.PhpVersion
	for _, v := range versions {
		if v.Version == "9.0" {
			v := v
			target = &v
			break
		}
	}
	assert.Nil(t, target, "version 9.0 should not be found when no PHP is installed")

	// SetActive should also return an error (mirrors SetActivePhp error path)
	err = mgr.SetActive("9.0")
	assert.Error(t, err, "SetActive should return an error for a non-installed version")
	assert.ErrorIs(t, err, php.ErrVersionNotFound,
		"SetActive should wrap ErrVersionNotFound")
}
