package php_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAll_NoVersionsFound(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))
	mockFS.On("ReadDir", paths.PhpDir()).Return([]fs.DirEntry{}, nil)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestDetectAll_SystemPhpFound(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	mockCmd.On("LookPath", "php").Return("/usr/bin/php", nil)
	// Run for getFullVersion
	mockCmd.On("Run", "/usr/bin/php", "-r", "echo PHP_VERSION;").Return("8.3.1", nil)
	// LookPath for php-fpm (not found is fine)
	mockCmd.On("LookPath", "php-fpm").Return("", errors.New("not found"))
	// detectManual - ReadDir returns empty
	mockFS.On("ReadDir", paths.PhpDir()).Return([]fs.DirEntry{}, nil)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "8.3.1", versions[0].Version)
	assert.Equal(t, "system", versions[0].Source)
}

func TestDetectAll_ManualVersionsFound(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// No system PHP
	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.2", isDir: true},
		&MockDirEntry{name: "8.3", isDir: true},
		&MockDirEntry{name: "current", isDir: true}, // should be skipped
		&MockDirEntry{name: "readme.txt", isDir: false}, // not a dir, should be skipped
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	// 8.2: binary exists, getFullVersion returns a version
	phpBin82 := paths.PhpVersionDir("8.2") + "/bin/php"
	fpmBin82 := paths.PhpVersionDir("8.2") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin82).Return(nil, nil) // nil FileInfo is fine since we only check err
	mockFS.On("Stat", fpmBin82).Return(nil, nil)
	mockCmd.On("Run", phpBin82, "-r", "echo PHP_VERSION;").Return("8.2.15", nil)

	// 8.3: binary exists
	phpBin83 := paths.PhpVersionDir("8.3") + "/bin/php"
	fpmBin83 := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", phpBin83).Return(nil, nil)
	mockFS.On("Stat", fpmBin83).Return(nil, nil)
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("8.3.10", nil)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 2)

	sources := make([]string, len(versions))
	for i, v := range versions {
		sources[i] = v.Source
	}
	assert.Contains(t, sources, "manual")
}

func TestDetectAll_ManualVersionSkippedWhenNoBinary(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	mockCmd.On("LookPath", "php").Return("", errors.New("not found"))

	entries := []fs.DirEntry{
		&MockDirEntry{name: "8.2", isDir: true},
	}
	mockFS.On("ReadDir", paths.PhpDir()).Return(entries, nil)

	// binary stat fails
	phpBin82 := paths.PhpVersionDir("8.2") + "/bin/php"
	mockFS.On("Stat", phpBin82).Return(nil, errors.New("not found"))

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestDetectAll_ActiveVersionMarked(t *testing.T) {
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

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 2)

	var active []php.PhpVersion
	for _, v := range versions {
		if v.Active {
			active = append(active, v)
		}
	}
	require.Len(t, active, 1)
	assert.Equal(t, "8.3.10", active[0].Version)
}

func TestDetectAll_FirstVersionActiveWhenNoActiveSet(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// store active is empty by default
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

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.True(t, versions[0].Active, "first version should be marked active when no active set")
}

func TestDetectAll_DuplicateVersionDeduped(t *testing.T) {
	mgr, mockFS, mockCmd, paths, _ := newTestManager(t)

	// System PHP reports 8.3.10
	mockCmd.On("LookPath", "php").Return("/usr/bin/php", nil)
	mockCmd.On("Run", "/usr/bin/php", "-r", "echo PHP_VERSION;").Return("8.3.10", nil)
	mockCmd.On("LookPath", "php-fpm").Return("", errors.New("not found"))

	// Manual also has 8.3 which resolves to 8.3.10
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
	// Duplicate should be deduplicated — only one entry for 8.3.10
	assert.Len(t, versions, 1)
	assert.Equal(t, "8.3.10", versions[0].Version)
}

func TestGetFullVersion_Success(t *testing.T) {
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
	// Version string with trailing newline (trimmed by getFullVersion)
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("8.3.1\n", nil)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "8.3.1", versions[0].Version)
}

func TestGetFullVersion_Error(t *testing.T) {
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
	// Run fails — getFullVersion returns "", so falls back to entry name
	mockCmd.On("Run", phpBin83, "-r", "echo PHP_VERSION;").Return("", errors.New("exec failed"))

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	// Falls back to entry name "8.3"
	require.Len(t, versions, 1)
	assert.Equal(t, "8.3", versions[0].Version)
}
