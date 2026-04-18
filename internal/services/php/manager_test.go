package php_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_ListInstalled(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.3", isDir: true},
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	phpBin83 := paths.PhpVersionDir("8.3") + "/bin/php"
	fpmBin83 := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin83).Return(nil, nil)
	mockFS.On("Stat", fpmBin83).Return(nil, errors.New("no fpm"))
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("8.3.10", nil)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "8.3.10", versions[0].Version)
	assert.Equal(t, "manual", versions[0].Source)
}

func TestManager_SetActive_KnownVersion(t *testing.T) {
	mgr, mockFS, mockCmd, paths, store := newTestManager(t)

	// Setup detection to find "8.3.1"
	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.3", isDir: true},
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	phpBin83 := paths.PhpVersionDir("8.3") + "/bin/php"
	fpmBin83 := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin83).Return(nil, nil)
	mockFS.On("Stat", fpmBin83).Return(nil, errors.New("no fpm"))
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("8.3.1", nil)

	err := mgr.SetActive("8.3.1")
	require.NoError(t, err)
	assert.Equal(t, "8.3.1", store.Get().ActivePhpVersion)
}

func TestManager_SetActive_UnknownVersion(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// No versions detected
	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))
	mockFS.On("ReadDir", paths.PhpDir()).Return([]fs.DirEntry{}, nil)

	err := mgr.SetActive("9.9")
	require.Error(t, err)
	assert.ErrorIs(t, err, php.ErrVersionNotFound)
}

func TestManager_GetActive_ReturnsActiveVersion(t *testing.T) {
	mgr, mockFS, mockCmd, paths, store := newTestManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3.10"))

	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.2", isDir: true},
		&MockDirEntry{name: "8.3", isDir: true},
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	phpBin82 := paths.PhpVersionDir("8.2") + "/bin/php"
	fpmBin82 := paths.PhpVersionDir("8.2") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin82).Return(nil, nil)
	mockFS.On("Stat", fpmBin82).Return(nil, errors.New("no fpm"))
	mockCmd.On("Run", phpBin82, "-r", "echo PHP_VERSION;").Return("8.2.15", nil)

	phpBin83 := paths.PhpVersionDir("8.3") + "/bin/php"
	fpmBin83 := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin83).Return(nil, nil)
	mockFS.On("Stat", fpmBin83).Return(nil, errors.New("no fpm"))
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("8.3.10", nil)

	active, err := mgr.GetActive()
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, "8.3.10", active.Version)
}

func TestManager_GetActive_FallsBackToFirst(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// No active version set in store
	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.2", isDir: true},
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	phpBin82 := paths.PhpVersionDir("8.2") + "/bin/php"
	fpmBin82 := paths.PhpVersionDir("8.2") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin82).Return(nil, nil)
	mockFS.On("Stat", fpmBin82).Return(nil, errors.New("no fpm"))
	mockCmd.On("Run", phpBin82, "-r", "echo PHP_VERSION;").Return("8.2.15", nil)

	active, err := mgr.GetActive()
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, "8.2.15", active.Version)
	assert.True(t, active.Active)
}

func TestManager_GetActive_NoVersions(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))
	mockFS.On("ReadDir", paths.PhpDir()).Return([]fs.DirEntry{}, nil)

	active, err := mgr.GetActive()
	assert.Nil(t, active)
	require.Error(t, err)
	assert.ErrorIs(t, err, php.ErrNoVersionsInstalled)
}

func TestManager_Uninstall_RemovesInactiveVersion(t *testing.T) {
	mgr, mockFS, mockCmd, paths, store := newTestManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	// Stat for dir check in Uninstall
	dir82 := paths.PhpVersionDir("8.2")
	mockFS.On("Stat", dir82).Return(nil, nil)
	mockFS.On("RemoveAll", dir82).Return(nil)

	// detectAll calls also happen within SetActive — but Uninstall doesn't call detectAll
	// It directly checks store active and fs.Stat on the directory
	_ = mockCmd // no cmd calls needed for Uninstall

	err := mgr.Uninstall("8.2")
	require.NoError(t, err)
	mockFS.AssertCalled(t, "RemoveAll", dir82)
}

func TestManager_Uninstall_RejectsActiveVersion(t *testing.T) {
	mgr, _, mockCmd, _, store := newTestManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))
	_ = mockCmd

	err := mgr.Uninstall("8.3")
	require.Error(t, err)
	assert.ErrorIs(t, err, php.ErrVersionActive)
}

func TestManager_Uninstall_RejectsNonexistentVersion(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// Active version is something else, so 8.2 is not protected
	dir82 := paths.PhpVersionDir("8.2")
	mockFS.On("Stat", dir82).Return(nil, os.ErrNotExist)
	_ = mockCmd

	err := mgr.Uninstall("8.2")
	require.Error(t, err)
	assert.ErrorIs(t, err, php.ErrVersionNotFound)
}
