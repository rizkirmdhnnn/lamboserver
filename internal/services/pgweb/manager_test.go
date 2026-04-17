// Package pgweb_test contains unit tests for the pgweb package.
package pgweb_test

import (
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/pgweb"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock definitions ---

// mockFS is a testify mock implementing pgweb.FileSystem.
type mockFS struct {
	mock.Mock
}

func (m *mockFS) Stat(name string) (fs.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(fs.FileInfo), args.Error(1)
}

func (m *mockFS) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

// mockCmd is a testify mock implementing pgweb.CommandRunner.
type mockCmd struct {
	mock.Mock
}

func (m *mockCmd) Run(name string, args ...string) (string, error) {
	allArgs := append([]string{name}, args...)
	callArgs := m.Called(allArgs)
	return callArgs.String(0), callArgs.Error(1)
}

// --- Test helpers ---

func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

func newTestManager(t *testing.T) (*pgweb.Manager, *mockFS, *mockCmd, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mfs := &mockFS{}
	mc := &mockCmd{}
	mgr := pgweb.NewManager(paths, mfs, mc)
	return mgr, mfs, mc, paths
}

// --- Tests ---

func TestIsInstalled_FalseWhenBinaryAbsent(t *testing.T) {
	mgr, mfs, _, paths := newTestManager(t)
	binaryPath := paths.PgwebDir() + "/pgweb"
	mfs.On("Stat", binaryPath).Return(nil, os.ErrNotExist)

	assert.False(t, mgr.IsInstalled())
	mfs.AssertExpectations(t)
}

func TestIsInstalled_TrueWhenBinaryPresent(t *testing.T) {
	mgr, mfs, _, paths := newTestManager(t)
	binaryPath := paths.PgwebDir() + "/pgweb"

	// Create a real file so we can get a real FileInfo.
	require.NoError(t, os.MkdirAll(paths.PgwebDir(), 0755))
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0755))
	fi, err := os.Stat(binaryPath)
	require.NoError(t, err)

	mfs.On("Stat", binaryPath).Return(fi, nil)

	assert.True(t, mgr.IsInstalled())
	mfs.AssertExpectations(t)
}

func TestInstall_Success(t *testing.T) {
	mgr, mfs, mc, paths := newTestManager(t)
	dir := paths.PgwebDir()
	binaryPath := dir + "/pgweb"
	zipPath := dir + "/pgweb.zip"

	mfs.On("MkdirAll", dir, fs.FileMode(0755)).Return(nil)

	// curl download
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Return("", nil).Once()

	// unzip
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Return("", nil).Once()

	// mv (rename extracted binary)
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "mv"
	})).Return("", nil)

	// chmod
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "chmod"
	})).Return("", nil)

	// rm cleanup (best-effort)
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && args[0] == "rm" && args[1] == "-f"
	})).Return("", nil).Maybe()

	_ = zipPath // referenced implicitly through paths

	// Post-install Stat success
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0755))
	fi, err := os.Stat(binaryPath)
	require.NoError(t, err)
	mfs.On("Stat", binaryPath).Return(fi, nil)

	err = mgr.Install()
	require.NoError(t, err)
}

func TestInstall_DownloadFails(t *testing.T) {
	mgr, mfs, mc, paths := newTestManager(t)
	dir := paths.PgwebDir()

	mfs.On("MkdirAll", dir, fs.FileMode(0755)).Return(nil)

	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Return("", os.ErrDeadlineExceeded)

	err := mgr.Install()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download pgweb")
}

func TestStopIdempotent(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)
	// proc is nil on a fresh manager — Stop must not panic and must return nil.
	err := mgr.Stop()
	require.NoError(t, err)
}

func TestURL(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)
	assert.Equal(t, "http://127.0.0.1:8081", mgr.URL())
}

func TestVersion(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)
	assert.Equal(t, "0.17.0", mgr.Version())
}

func TestStatus_NotInstalled(t *testing.T) {
	mgr, mfs, _, paths := newTestManager(t)
	binaryPath := paths.PgwebDir() + "/pgweb"
	mfs.On("Stat", binaryPath).Return(nil, os.ErrNotExist)

	status := mgr.Status()
	assert.False(t, status.Installed)
	assert.False(t, status.Running)
	assert.Equal(t, 8081, status.Port)
	mfs.AssertExpectations(t)
}

// --- PGW-02 gap tests: Start, IsRunning, checkPort ---

// writeFakeBinary writes a minimal shell-script stub to the given path so that
// exec.Command can successfully launch it during tests.
func writeFakeBinary(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	// A long-running stub that ignores all flags pgweb.Start passes.
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\nsleep 30\n"), 0755))
}

// TestIsRunning_FalseWhenProcNil verifies that IsRunning returns false on a
// freshly-constructed manager where no process has been started yet.
func TestIsRunning_FalseWhenProcNil(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)
	assert.False(t, mgr.IsRunning())
}

// TestIsRunning_TrueWhenProcessAlive verifies that IsRunning returns true after
// Start() successfully launches a background process.
func TestIsRunning_TrueWhenProcessAlive(t *testing.T) {
	mgr, _, _, paths := newTestManager(t)
	binaryPath := paths.PgwebDir() + "/pgweb"
	writeFakeBinary(t, binaryPath)

	err := mgr.Start()
	require.NoError(t, err)
	t.Cleanup(func() { mgr.Stop() }) //nolint:errcheck

	assert.True(t, mgr.IsRunning())
}

// TestStart_IdempotentWhenAlreadyRunning verifies that calling Start() a second
// time while pgweb is already running returns nil without launching a second process.
func TestStart_IdempotentWhenAlreadyRunning(t *testing.T) {
	mgr, _, _, paths := newTestManager(t)
	binaryPath := paths.PgwebDir() + "/pgweb"
	writeFakeBinary(t, binaryPath)

	require.NoError(t, mgr.Start())
	t.Cleanup(func() { mgr.Stop() }) //nolint:errcheck

	// Second Start() call must be idempotent.
	err := mgr.Start()
	assert.NoError(t, err)
}

// TestStart_PortConflictReturnsError verifies that Start() refuses to launch pgweb
// when port 8081 is already occupied, returning the descriptive T-06-02 error.
func TestStart_PortConflictReturnsError(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)

	// Occupy port 8081 for the duration of this test.
	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		t.Skip("port 8081 is already occupied by the host environment; skipping conflict test")
	}
	defer ln.Close()

	startErr := mgr.Start()
	require.Error(t, startErr)
	assert.Contains(t, startErr.Error(), "port 8081 is already in use")
}

// TestStart_MissingBinaryReturnsWrappedError verifies that when the pgweb binary
// does not exist, Start() returns an error wrapped with "failed to start pgweb".
// This confirms exec.Command (non-blocking cmd.Start) is used, not cmd.Run.
func TestStart_MissingBinaryReturnsWrappedError(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)
	// Binary does not exist in temp paths — exec.Command.Start() will fail.

	err := mgr.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start pgweb")
}

// TestCheckPort_FreePortAllowsStart verifies that checkPort passes (port is free)
// and Start proceeds to attempt exec.Command.  The binary is absent so the exec
// fails, but the error is "failed to start pgweb" — NOT a port-conflict error —
// confirming checkPort itself returned nil.
func TestCheckPort_FreePortAllowsStart(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)

	err := mgr.Start()
	// Port is free so checkPort succeeds; exec.Command fails due to missing binary.
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "port 8081 is already in use",
		"error should come from exec failure, not from port check")
	assert.Contains(t, err.Error(), "failed to start pgweb")
}

// TestCheckPort_OccupiedPortErrors verifies checkPort returns an error when the
// port is already bound, exercising T-06-02 DoS mitigation via net.Listen probe.
func TestCheckPort_OccupiedPortErrors(t *testing.T) {
	mgr, _, _, _ := newTestManager(t)

	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		t.Skip("port 8081 is already occupied by the host environment; skipping")
	}
	defer ln.Close()

	startErr := mgr.Start()
	require.Error(t, startErr)
	assert.Contains(t, startErr.Error(), "port 8081 is already in use by another process")
}
