package system_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_IsInstalled_False_WhenNotInstalled(t *testing.T) {
	integration, _ := newTestIntegration(t)
	assert.False(t, integration.IsInstalled())
}

func TestIntegration_Install_AddsPathToRCFiles(t *testing.T) {
	integration, tmpHome := newTestIntegration(t)

	require.NoError(t, integration.Install())
	assert.True(t, integration.IsInstalled())

	data, err := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "# Added by LamboServer")
	assert.Contains(t, content, ".lamboserver/bin")
}

func TestIntegration_Install_Idempotent(t *testing.T) {
	integration, tmpHome := newTestIntegration(t)

	require.NoError(t, integration.Install())
	require.NoError(t, integration.Install())

	data, err := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	require.NoError(t, err)
	count := strings.Count(string(data), "# Added by LamboServer")
	assert.Equal(t, 1, count, "marker should appear exactly once after two installs")
}

func TestIntegration_Install_ExistingBashrc(t *testing.T) {
	integration, tmpHome := newTestIntegration(t)

	// Pre-create .bashrc
	bashrc := filepath.Join(tmpHome, ".bashrc")
	require.NoError(t, os.WriteFile(bashrc, []byte("# existing bashrc\n"), 0644))

	require.NoError(t, integration.Install())

	// Both .zshrc and .bashrc should contain the marker
	for _, rcFile := range []string{".zshrc", ".bashrc"} {
		data, err := os.ReadFile(filepath.Join(tmpHome, rcFile))
		require.NoError(t, err, "rc file should exist: %s", rcFile)
		assert.Contains(t, string(data), "# Added by LamboServer", "marker should be in %s", rcFile)
	}
}

func TestIntegration_Uninstall_RemovesBlock(t *testing.T) {
	integration, tmpHome := newTestIntegration(t)

	require.NoError(t, integration.Install())
	require.True(t, integration.IsInstalled())

	require.NoError(t, integration.Uninstall())
	assert.False(t, integration.IsInstalled())

	data, err := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "# Added by LamboServer")
}

func TestIntegration_Uninstall_PreservesOtherContent(t *testing.T) {
	integration, tmpHome := newTestIntegration(t)

	zshrc := filepath.Join(tmpHome, ".zshrc")
	require.NoError(t, os.WriteFile(zshrc, []byte("export FOO=bar\n"), 0644))

	require.NoError(t, integration.Install())
	require.NoError(t, integration.Uninstall())

	data, err := os.ReadFile(zshrc)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "export FOO=bar", "original content should be preserved")
	assert.NotContains(t, content, "# Added by LamboServer", "marker should be removed")
}

func TestIntegration_CreateSymlink(t *testing.T) {
	integ, tmpHome := newTestIntegration(t)
	binDir := filepath.Join(tmpHome, ".lamboserver", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "php-real")
	require.NoError(t, os.WriteFile(target, []byte("#!/bin/sh"), 0755))

	require.NoError(t, integ.CreateSymlink("php", target))

	linkPath := filepath.Join(binDir, "php")
	info, err := os.Lstat(linkPath)
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), info.Mode()&os.ModeSymlink, "should be a symlink")

	resolved, err := os.Readlink(linkPath)
	require.NoError(t, err)
	assert.Equal(t, target, resolved)
}

func TestIntegration_RemoveSymlink(t *testing.T) {
	integ, tmpHome := newTestIntegration(t)
	binDir := filepath.Join(tmpHome, ".lamboserver", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "php-real")
	require.NoError(t, os.WriteFile(target, []byte("#!/bin/sh"), 0755))

	require.NoError(t, integ.CreateSymlink("php", target))

	linkPath := filepath.Join(binDir, "php")
	_, err := os.Lstat(linkPath)
	require.NoError(t, err, "symlink should exist before removal")

	require.NoError(t, integ.RemoveSymlink("php"))

	_, err = os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(err), "symlink should not exist after removal")
}

func TestIntegration_CreateSymlink_OverwritesExisting(t *testing.T) {
	integ, tmpHome := newTestIntegration(t)
	binDir := filepath.Join(tmpHome, ".lamboserver", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	targetDir := t.TempDir()
	target1 := filepath.Join(targetDir, "php-v1")
	target2 := filepath.Join(targetDir, "php-v2")
	require.NoError(t, os.WriteFile(target1, []byte("#!/bin/sh"), 0755))
	require.NoError(t, os.WriteFile(target2, []byte("#!/bin/sh"), 0755))

	require.NoError(t, integ.CreateSymlink("php", target1))
	require.NoError(t, integ.CreateSymlink("php", target2))

	resolved, err := os.Readlink(filepath.Join(binDir, "php"))
	require.NoError(t, err)
	assert.Equal(t, target2, resolved)
}
