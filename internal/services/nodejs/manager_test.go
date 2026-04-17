package nodejs_test

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/nodejs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) (*nodejs.Manager, *MockFileSystem, *MockCommandRunner, *MockFileSystem) {
	t.Helper()
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)
	return mgr, mockFS, mockCmd, nil
}

func TestManager_ListInstalled_Empty(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{
		ReadDirFunc: func(name string) ([]fs.DirEntry, error) {
			return []fs.DirEntry{}, nil
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestManager_ListInstalled_NotExistDir(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{
		ReadDirFunc: func(name string) ([]fs.DirEntry, error) {
			return nil, os.ErrNotExist
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err, "ErrNotExist should be handled gracefully")
	assert.Empty(t, versions)
}

func TestManager_ListInstalled_ReadDirError(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	sentinelErr := errors.New("permission denied")
	mockFS := &MockFileSystem{
		ReadDirFunc: func(name string) ([]fs.DirEntry, error) {
			return nil, sentinelErr
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	_, err := mgr.ListInstalled()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinelErr)
}

func TestManager_ListInstalled_WithVersions(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)

	// Set active version in store
	require.NoError(t, store.SetActiveNodeVersion("v20.0.0"))

	mockFS := &MockFileSystem{
		ReadDirFunc: func(name string) ([]fs.DirEntry, error) {
			return []fs.DirEntry{
				&MockDirEntry{name: "v20.0.0", isDir: true},
				&MockDirEntry{name: "v18.0.0", isDir: true},
				&MockDirEntry{name: "current", isDir: true}, // should be skipped
			}, nil
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Len(t, versions, 2, "current entry should be skipped")

	var v20 *nodejs.NodeVersion
	for i := range versions {
		if versions[i].Version == "v20.0.0" {
			v20 = &versions[i]
		}
	}
	require.NotNil(t, v20, "v20.0.0 should be present")
	assert.True(t, v20.Active, "v20.0.0 should be marked as active")

	for _, v := range versions {
		if v.Version == "v18.0.0" {
			assert.False(t, v.Active, "v18.0.0 should not be active")
		}
	}
}

func TestManager_ListInstalled_NoActive(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t) // active version is empty by default
	mockFS := &MockFileSystem{
		ReadDirFunc: func(name string) ([]fs.DirEntry, error) {
			return []fs.DirEntry{
				&MockDirEntry{name: "v20.0.0", isDir: true},
			}, nil
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	versions, err := mgr.ListInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.False(t, versions[0].Active, "version should not be active when store has no active version")
}

func TestManager_Install_Success(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("v20.0.0")
	require.NoError(t, err)

	require.Len(t, mockFS.MkdirAllCalls, 1, "MkdirAll should be called once")
	require.Len(t, mockCmd.RunCalls, 1, "Run should be called once")

	call := mockCmd.RunCalls[0]
	assert.Equal(t, "sh", call.Name)
	require.Len(t, call.Args, 2)
	assert.Equal(t, "-c", call.Args[0])
	assert.Contains(t, call.Args[1], "curl")
	assert.Contains(t, call.Args[1], "v20.0.0")
}

func TestManager_Install_PrependV(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("20.0.0") // no "v" prefix
	require.NoError(t, err)

	require.Len(t, mockCmd.RunCalls, 1)
	call := mockCmd.RunCalls[0]
	assert.True(t, strings.Contains(call.Args[1], "v20.0.0"),
		"version should have 'v' prepended, got: %s", call.Args[1])
}

func TestManager_Install_MkdirFails(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mkdirErr := errors.New("no space left on device")
	mockFS := &MockFileSystem{
		MkdirAllFunc: func(path string, perm fs.FileMode) error {
			return mkdirErr
		},
	}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("v20.0.0")
	require.Error(t, err)
	assert.ErrorIs(t, err, mkdirErr)
	assert.Empty(t, mockCmd.RunCalls, "Run should not be called when MkdirAll fails")
}

func TestManager_Install_CmdFails(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	cmdErr := errors.New("curl: command not found")
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{
		RunFunc: func(name string, args ...string) (string, error) {
			return "", cmdErr
		},
	}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("v20.0.0")
	require.Error(t, err)
	assert.ErrorIs(t, err, cmdErr)
}

func TestManager_SetActive(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.SetActive("v20.0.0")
	require.NoError(t, err)
	assert.Equal(t, "v20.0.0", store.Get().ActiveNodeVersion)
}

func TestManager_Uninstall(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Uninstall("v20.0.0")
	require.NoError(t, err)

	require.Len(t, mockFS.RemoveAllCalls, 1)
	expectedPath := paths.NodeVersionDir("v20.0.0")
	assert.Equal(t, expectedPath, mockFS.RemoveAllCalls[0])
}

func TestManager_Install_RejectsInvalidVersion(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("v18.0.0; rm -rf /")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Node.js version")
	assert.Empty(t, mockCmd.RunCalls, "Run should not be called for invalid version")
}

func TestManager_Install_RejectsNonNumericVersion(t *testing.T) {
	paths := newTestPaths(t)
	store := newTestStore(t)
	mockFS := &MockFileSystem{}
	mockCmd := &MockCommandRunner{}
	mgr := nodejs.NewManagerWithDeps(paths, store, mockFS, mockCmd)

	err := mgr.Install("latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Node.js version")
	assert.Empty(t, mockCmd.RunCalls, "Run should not be called for non-numeric version")
}
