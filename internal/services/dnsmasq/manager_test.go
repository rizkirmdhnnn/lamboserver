package dnsmasq_test

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/dnsmasq"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) (*dnsmasq.Manager, *MockFileSystem, *MockLaunchdService, *MockAdminRunner, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{}
	mockLaunchd := &MockLaunchdService{}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, mockAdmin)
	return mgr, mockFS, mockLaunchd, mockAdmin, paths
}

func TestManager_IsDaemonInstalled_True(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		StatFunc: func(name string) (fs.FileInfo, error) {
			// Return success — plist exists
			return &mockFileInfo{name: "com.lamboserver.dnsmasq.plist"}, nil
		},
	}
	mockLaunchd := &MockLaunchdService{}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, mockAdmin)

	assert.True(t, mgr.IsDaemonInstalled())
}

func TestManager_IsDaemonInstalled_False(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		StatFunc: func(name string) (fs.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}
	mockLaunchd := &MockLaunchdService{}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, mockAdmin)

	assert.False(t, mgr.IsDaemonInstalled())
}

func TestManager_EnsureConfig(t *testing.T) {
	_, mockFS, _, _, paths := newTestManager(t)
	mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, &MockAdminRunner{})

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	require.Len(t, mockFS.WriteFileCalls, 1)
	call := mockFS.WriteFileCalls[0]
	assert.Equal(t, paths.DnsmasqConf(), call.Name)

	content := string(call.Data)
	assert.Contains(t, content, "address=/.test/127.0.0.1")
	assert.Contains(t, content, "listen-address=127.0.0.1")
	assert.Contains(t, content, "port=53")
}

func TestManager_EnsureResolver_CreatesNew(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		ReadFileFunc: func(name string) ([]byte, error) {
			// File doesn't exist
			return nil, os.ErrNotExist
		},
	}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, mockAdmin)

	err := mgr.EnsureResolver()
	require.NoError(t, err)

	// WriteFile called for temp file
	require.NotEmpty(t, mockFS.WriteFileCalls)

	// RunWithPrivileges called with mkdir + cp command
	require.Len(t, mockAdmin.RunWithPrivilegesCalls, 1)
	cmd := mockAdmin.RunWithPrivilegesCalls[0]
	assert.Contains(t, cmd, "mkdir -p")
	assert.Contains(t, cmd, "cp")
}

func TestManager_EnsureResolver_AlreadyCorrect(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		ReadFileFunc: func(name string) ([]byte, error) {
			// Return the exact expected content
			return []byte("nameserver 127.0.0.1\n"), nil
		},
	}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, mockAdmin)

	err := mgr.EnsureResolver()
	require.NoError(t, err)

	// No WriteFile or RunWithPrivileges should be called (early return)
	assert.Empty(t, mockFS.WriteFileCalls, "WriteFile should not be called when content is already correct")
	assert.Empty(t, mockAdmin.RunWithPrivilegesCalls, "RunWithPrivileges should not be called when content is already correct")
}

func TestManager_EnsureResolver_AdminFails(t *testing.T) {
	paths := newTestPaths(t)
	adminErr := errors.New("admin authentication failed")
	mockFS := &MockFileSystem{
		ReadFileFunc: func(name string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
	}
	mockAdmin := &MockAdminRunner{
		RunWithPrivilegesFunc: func(command string) error {
			return adminErr
		},
	}
	mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, mockAdmin)

	err := mgr.EnsureResolver()
	require.Error(t, err)
	assert.ErrorIs(t, err, adminErr)
}

func TestManager_Stop(t *testing.T) {
	_, mockFS, mockLaunchd, _, paths := newTestManager(t)
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, &MockAdminRunner{})

	err := mgr.Stop()
	require.NoError(t, err)

	require.Len(t, mockLaunchd.UninstallCalls, 1)
	cfg := mockLaunchd.UninstallCalls[0]
	assert.Equal(t, "com.lamboserver.dnsmasq", cfg.Label)
	assert.Equal(t, system.ServiceDaemon, cfg.Type)
}

func TestManager_Status_AllTrue(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		StatFunc: func(name string) (fs.FileInfo, error) {
			// Resolver file exists
			return &mockFileInfo{name: "test"}, nil
		},
	}
	mockLaunchd := &MockLaunchdService{
		IsRunningFunc: func(label string) bool {
			return true
		},
	}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, &MockAdminRunner{})

	status := mgr.Status()
	assert.True(t, status.Running)
	assert.True(t, status.Resolver)
}

func TestManager_Status_AllFalse(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		StatFunc: func(name string) (fs.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}
	mockLaunchd := &MockLaunchdService{
		IsRunningFunc: func(label string) bool {
			return false
		},
	}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, &MockAdminRunner{})

	status := mgr.Status()
	assert.False(t, status.Running)
	assert.False(t, status.Resolver)
}

func TestManager_Start_BinaryExists(t *testing.T) {
	paths := newTestPaths(t)

	// Create the dnsmasq binary file at paths.DnsmasqBin() so BinaryLocator.IsInstalled() returns true
	binPath := paths.DnsmasqBin()
	require.NoError(t, os.MkdirAll(strings.TrimSuffix(binPath, "/dnsmasq"), 0755))
	f, err := os.Create(binPath)
	require.NoError(t, err)
	f.Close()
	require.NoError(t, os.Chmod(binPath, 0755))

	mockFS := &MockFileSystem{}
	mockLaunchd := &MockLaunchdService{}
	mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, &MockAdminRunner{})

	err = mgr.Start()
	require.NoError(t, err)

	require.Len(t, mockLaunchd.InstallCalls, 1)
	cfg := mockLaunchd.InstallCalls[0]
	assert.Equal(t, "com.lamboserver.dnsmasq", cfg.Label)
	assert.Equal(t, system.ServiceDaemon, cfg.Type)
	assert.True(t, cfg.KeepAlive)
	assert.True(t, cfg.RunAtLoad)
}

func TestManager_EnsureResolver_ShellQuotesPaths(t *testing.T) {
	paths := newTestPaths(t)
	mockFS := &MockFileSystem{
		ReadFileFunc: func(name string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
	}
	mockAdmin := &MockAdminRunner{}
	mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, mockAdmin)

	err := mgr.EnsureResolver()
	require.NoError(t, err)

	require.Len(t, mockAdmin.RunWithPrivilegesCalls, 1)
	cmd := mockAdmin.RunWithPrivilegesCalls[0]

	// Verify all three path arguments are wrapped in single quotes
	assert.Contains(t, cmd, "mkdir -p '")
	assert.Contains(t, cmd, "cp '")
	// Count single quotes: should have at least 6 (2 per path × 3 paths)
	quoteCount := strings.Count(cmd, "'")
	assert.GreaterOrEqual(t, quoteCount, 6, "each of the 3 paths should be wrapped in single quotes")
}
