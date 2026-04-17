// Package postgresql_test contains unit tests for the postgresql package.
package postgres_test

import (
	"database/sql"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/postgres"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock definitions ---

// mockFS is a testify mock implementing postgres.FileSystem.
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

func (m *mockFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	args := m.Called(name, data, perm)
	return args.Error(0)
}

func (m *mockFS) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *mockFS) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *mockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]fs.DirEntry), args.Error(1)
}

func (m *mockFS) ReadFile(name string) ([]byte, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// mockCmd is a testify mock implementing postgres.CommandRunner.
type mockCmd struct {
	mock.Mock
}

func (m *mockCmd) Run(name string, args ...string) (string, error) {
	allArgs := append([]string{name}, args...)
	callArgs := m.Called(allArgs)
	return callArgs.String(0), callArgs.Error(1)
}

// mockHelper is a testify mock implementing postgres.HelperRunner.
type mockHelper struct {
	mock.Mock
}

func (m *mockHelper) Run(args ...string) (string, error) {
	callArgs := m.Called(args)
	return callArgs.String(0), callArgs.Error(1)
}

// mockAdmin is a testify mock implementing postgres.AdminRunner.
type mockAdmin struct {
	mock.Mock
}

func (m *mockAdmin) RunWithPrivileges(command string) error {
	args := m.Called(command)
	return args.Error(0)
}

// --- Test helpers ---

func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}

func newTestManager(t *testing.T) (*postgres.Manager, *mockFS, *mockCmd, *mockHelper, *mockAdmin, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mfs := &mockFS{}
	mc := &mockCmd{}
	mh := &mockHelper{}
	ma := &mockAdmin{}
	mgr := postgres.NewManager(paths, mfs, mc, mh, ma)
	return mgr, mfs, mc, mh, ma, paths
}

// port5432IsFree returns true if port 5432 can be bound (i.e., nothing is using it).
func port5432IsFree() bool {
	ln, err := net.Listen("tcp", "127.0.0.1:5432")
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// containsStr returns true if any element of slice equals target.
func containsStr(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

// --- TestIsInstalled ---

func TestIsInstalled_True(t *testing.T) {
	_, _, _, _, _, paths := newTestManager(t)

	// Create the pg_ctl binary so BinaryLocator.IsInstalled() returns true.
	require.NoError(t, os.MkdirAll(paths.PostgreSQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.PostgreSQLBinDir()+"/pg_ctl", []byte("#!/bin/sh\n"), 0755))

	// Create a new manager — BinaryLocator checks the file on IsInstalled().
	mgr := postgres.NewManager(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{})
	assert.True(t, mgr.IsInstalled())
}

func TestIsInstalled_FalseWhenLocalPathAbsent(t *testing.T) {
	mgr, _, _, _, _, _ := newTestManager(t)
	// BinaryLocator.IsInstalled() checks LocalPath first, then system PATH.
	// The LocalPath (tempdir/postgresql/bin/pg_ctl) does not exist, so if pg_ctl is also
	// absent from system PATH the result is false. On machines with PostgreSQL installed
	// via Homebrew or system packages, system PATH may have pg_ctl and this returns true.
	if mgr.IsInstalled() {
		t.Log("pg_ctl found on system PATH — IsInstalled() correctly returns true; skipping false-assertion")
		return
	}
	assert.False(t, mgr.IsInstalled())
}

// --- TestInstall ---

func TestInstall_ArchDetection(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// MkdirAll for PostgreSQL root dir.
	mfs.On("MkdirAll", paths.PostgreSQLDir(), fs.FileMode(0755)).Return(nil)

	var capturedCmd string
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Run(func(args mock.Arguments) {
		allArgs := args.Get(0).([]string)
		if len(allArgs) >= 3 {
			capturedCmd = allArgs[2]
		}
		// Create pg_ctl binary so Stat succeeds after extraction.
		os.MkdirAll(paths.PostgreSQLBinDir(), 0755)
		os.WriteFile(paths.PostgreSQLBinDir()+"/pg_ctl", []byte(""), 0755)
	}).Return("", nil)

	// Stat for pg_ctl binary verification.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLBinDir(), "pg_ctl")).Return(nil, nil)

	err := mgr.Install()
	require.NoError(t, err)

	// PostgreSQL uses "aarch64" not "arm64" — both aarch64 and x86_64 are valid.
	hasArch := strings.Contains(capturedCmd, "aarch64") || strings.Contains(capturedCmd, "x86_64")
	assert.True(t, hasArch, "URL should contain aarch64 or x86_64 (NOT arm64)")
	assert.Contains(t, capturedCmd, "theseus-rs/postgresql-binaries")
	assert.Contains(t, capturedCmd, "17.9.0")

	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

func TestInstall_CurlFailure(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.PostgreSQLDir(), fs.FileMode(0755)).Return(nil)

	downloadErr := errors.New("network error")
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Return("", downloadErr)

	err := mgr.Install()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download")
}

// --- TestInitDataDir ---

func TestInitDataDir_Success(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// PG_VERSION absent — not yet initialized.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLDataDir(), "PG_VERSION")).
		Return(nil, os.ErrNotExist)
	// initdb command.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "initdb")
	})).Return("", nil)
	// WriteFile for postgres.conf.
	mfs.On("WriteFile", paths.PostgreSQLConfFile(), mock.Anything, fs.FileMode(0644)).Return(nil)
	// WriteFile for pg_hba.conf.
	mfs.On("WriteFile", paths.PostgreSQLHbaFile(), mock.Anything, fs.FileMode(0644)).Return(nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)
	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

func TestInitDataDir_SkipsWhenAlreadyInitialized(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// PG_VERSION present — already initialized.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLDataDir(), "PG_VERSION")).
		Return(nil, nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)
	// cmd.Run should NOT have been called (no initdb invocation).
	mc.AssertNotCalled(t, "Run")
}

func TestInitDataDir_HbaContainsLocalAllAllTrust(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// PG_VERSION absent — not yet initialized.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLDataDir(), "PG_VERSION")).
		Return(nil, os.ErrNotExist)
	// initdb command.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "initdb")
	})).Return("", nil)
	// WriteFile for postgres.conf.
	mfs.On("WriteFile", paths.PostgreSQLConfFile(), mock.Anything, fs.FileMode(0644)).Return(nil)

	// Capture pg_hba.conf content.
	var capturedHba string
	mfs.On("WriteFile", paths.PostgreSQLHbaFile(), mock.Anything, fs.FileMode(0644)).
		Run(func(args mock.Arguments) {
			capturedHba = string(args.Get(1).([]byte))
		}).Return(nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)

	// D-12: pg_hba.conf must use trust auth for all local connections.
	assert.Contains(t, capturedHba, "local   all       all")
	assert.Contains(t, capturedHba, "host    all       all   127.0.0.1/32  trust")
	assert.Contains(t, capturedHba, "host    all       all   ::1/128       trust")
	assert.Contains(t, capturedHba, "trust")
	// Verify it does NOT restrict to postgres user only.
	assert.NotContains(t, capturedHba, "local   all   postgres")
}

func TestInitDataDir_CleanupOnInitdbFailure(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// PG_VERSION absent — not yet initialized.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLDataDir(), "PG_VERSION")).
		Return(nil, os.ErrNotExist)

	initdbErr := errors.New("initdb: fatal error")
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "initdb")
	})).Return("", initdbErr)

	// RemoveAll must be called to clean partial data dir on failure (D-09).
	mfs.On("RemoveAll", paths.PostgreSQLDataDir()).Return(nil)

	err := mgr.InitDataDir()
	require.Error(t, err)

	mfs.AssertCalled(t, "RemoveAll", paths.PostgreSQLDataDir())
	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

func TestInitDataDir_NoCleanupOnSuccess(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	// PG_VERSION absent — not yet initialized.
	mfs.On("Stat", filepath.Join(paths.PostgreSQLDataDir(), "PG_VERSION")).
		Return(nil, os.ErrNotExist)

	// initdb succeeds.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "initdb")
	})).Return("", nil)

	// WriteFile for postgres.conf.
	mfs.On("WriteFile", paths.PostgreSQLConfFile(), mock.Anything, fs.FileMode(0644)).Return(nil)
	// WriteFile for pg_hba.conf.
	mfs.On("WriteFile", paths.PostgreSQLHbaFile(), mock.Anything, fs.FileMode(0644)).Return(nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)

	// RemoveAll must NOT be called when initdb succeeds.
	mfs.AssertNotCalled(t, "RemoveAll", mock.Anything)
	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

// --- TestStart ---

func TestStart_Success(t *testing.T) {
	_, _, _, _, _, paths := newTestManager(t)

	// Create pg_ctl binary so IsInstalled() is true.
	require.NoError(t, os.MkdirAll(paths.PostgreSQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.PostgreSQLBinDir()+"/pg_ctl", []byte("#!/bin/sh\n"), 0755))

	mfs := &mockFS{}
	mc := &mockCmd{}

	// Recreate manager with real binary on disk.
	mgr := postgres.NewManager(paths, mfs, mc, &mockHelper{}, &mockAdmin{})

	// Note: checkPort uses real net.Listen — port 5432 must be free.
	if !port5432IsFree() {
		t.Skip("port 5432 is in use — skipping TestStart_Success")
	}

	// postmaster.pid absent — no stale PID.
	mfs.On("ReadFile", filepath.Join(paths.PostgreSQLDataDir(), "postmaster.pid")).
		Return(nil, os.ErrNotExist)
	// MkdirAll for socket dir.
	mfs.On("MkdirAll", paths.PostgreSQLSocketDir(), fs.FileMode(0755)).Return(nil)
	// MkdirAll for log dir.
	mfs.On("MkdirAll", filepath.Join(paths.PostgreSQLDataDir(), "log"), fs.FileMode(0755)).Return(nil)
	// pg_ctl start command.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "pg_ctl") && args[len(args)-1] == "start"
	})).Return("", nil)

	err := mgr.Start()
	require.NoError(t, err)
	mc.AssertExpectations(t)
}

func TestStart_FailsWhenNotInstalled(t *testing.T) {
	mgr, _, _, _, _, _ := newTestManager(t)
	if mgr.IsInstalled() {
		t.Skip("pg_ctl found on system PATH; skipping 'not installed' test")
	}
	err := mgr.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestStart_RemovesStalePID(t *testing.T) {
	_, _, _, _, _, paths := newTestManager(t)

	// Create pg_ctl binary so IsInstalled() is true.
	require.NoError(t, os.MkdirAll(paths.PostgreSQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.PostgreSQLBinDir()+"/pg_ctl", []byte("#!/bin/sh\n"), 0755))

	mfs := &mockFS{}
	mc := &mockCmd{}

	// Recreate manager with real binary on disk.
	mgr := postgres.NewManager(paths, mfs, mc, &mockHelper{}, &mockAdmin{})

	// Note: checkPort uses real net.Listen — port 5432 must be free.
	if !port5432IsFree() {
		t.Skip("port 5432 is in use — skipping TestStart_RemovesStalePID")
	}

	pidFile := filepath.Join(paths.PostgreSQLDataDir(), "postmaster.pid")

	// PID 99999 is almost certainly dead (ESRCH) — stale PID scenario.
	mfs.On("ReadFile", pidFile).
		Return([]byte("99999\ndata_dir\nextra\n"), nil)
	// RemoveAll for the stale pid file.
	mfs.On("RemoveAll", pidFile).Return(nil)
	// MkdirAll for socket dir.
	mfs.On("MkdirAll", paths.PostgreSQLSocketDir(), fs.FileMode(0755)).Return(nil)
	// MkdirAll for log dir.
	mfs.On("MkdirAll", filepath.Join(paths.PostgreSQLDataDir(), "log"), fs.FileMode(0755)).Return(nil)
	// pg_ctl start command.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "pg_ctl") && args[len(args)-1] == "start"
	})).Return("", nil)

	err := mgr.Start()
	require.NoError(t, err)
	// RemoveAll must have been called on the stale pid file.
	mfs.AssertCalled(t, "RemoveAll", pidFile)
}

// --- TestStop ---

func TestStop_Success(t *testing.T) {
	mgr, mfs, mc, _, _, paths := newTestManager(t)

	pidFile := filepath.Join(paths.PostgreSQLDataDir(), "postmaster.pid")

	// pg_ctl stop command succeeds.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "pg_ctl") && containsStr(args, "stop")
	})).Return("", nil)

	// postmaster.pid does not exist immediately → poll loop exits on first check.
	mfs.On("Stat", pidFile).Return(nil, os.ErrNotExist)

	err := mgr.Stop()
	require.NoError(t, err)
	mc.AssertExpectations(t)
}

func TestStop_CmdFailure(t *testing.T) {
	mgr, _, mc, _, _, _ := newTestManager(t)

	stopErr := errors.New("pg_ctl: no data directory specified")
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "pg_ctl")
	})).Return("", stopErr)

	err := mgr.Stop()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stop")
}

// --- TestStatus ---

func TestStatus_NotInstalled(t *testing.T) {
	mgr, mfs, _, _, _, _ := newTestManager(t)

	// postmaster.pid absent — not running.
	mfs.On("ReadFile", mock.Anything).Return(nil, os.ErrNotExist)

	status := mgr.Status()
	// Port must always be 5432.
	assert.Equal(t, 5432, status.Port)
	// IsRunning is false when postmaster.pid is absent.
	assert.False(t, status.Running)
	// Installed depends on system PATH; just verify it matches IsInstalled().
	assert.Equal(t, mgr.IsInstalled(), status.Installed)
}

// --- TestIsRunning ---

func TestIsRunning_FalseWhenNoPidFile(t *testing.T) {
	mgr, mfs, _, _, _, paths := newTestManager(t)

	pidFile := filepath.Join(paths.PostgreSQLDataDir(), "postmaster.pid")
	mfs.On("ReadFile", pidFile).Return(nil, os.ErrNotExist)

	assert.False(t, mgr.IsRunning())
	mfs.AssertExpectations(t)
}

// --- mockDBOpener ---

// mockDBOpener is a testify mock implementing postgres.DBOpener.
type mockDBOpener struct {
	mock.Mock
	capturedDriver string
	capturedDSN    string
}

func (m *mockDBOpener) Open(driverName, dataSourceName string) (*sql.DB, error) {
	m.capturedDriver = driverName
	m.capturedDSN = dataSourceName
	args := m.Called(driverName, dataSourceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.DB), args.Error(1)
}

// --- Database CRUD Tests ---

func TestCreateDatabase_ValidName(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mfs := &mockFS{}
	mc := &mockCmd{}
	mh := &mockHelper{}
	mdb := &mockDBOpener{}
	mdb.On("Open", "pgx", mock.Anything).Return(nil, errors.New("connection refused"))

	mgr := postgres.NewManagerWithDBOpener(paths, mfs, mc, mh, &mockAdmin{}, mdb)

	err := mgr.CreateDatabase("mydb")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid database name")
	assert.Equal(t, "pgx", mdb.capturedDriver)
	assert.Contains(t, mdb.capturedDSN, "host=")
	assert.Contains(t, mdb.capturedDSN, paths.PostgreSQLSocketDir())
	assert.Contains(t, mdb.capturedDSN, "user=postgres")
}

func TestCreateDatabase_InvalidName_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.CreateDatabase("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_InvalidName_Semicolon(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.CreateDatabase("test;drop")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_InvalidName_TooLong(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	longName := strings.Repeat("a", 64) // 64 chars exceeds PostgreSQL's 63-char limit
	err := mgr.CreateDatabase(longName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_ValidName_MaxLength(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mdb := &mockDBOpener{}
	mdb.On("Open", "pgx", mock.Anything).Return(nil, errors.New("connection refused"))
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	maxName := strings.Repeat("a", 63) // Exactly 63 chars — max valid length
	err := mgr.CreateDatabase(maxName)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid database name") // Passes validation, fails at connection
}

func TestDropDatabase_SystemDB_postgres(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.DropDatabase("postgres")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_template0(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.DropDatabase("template0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_template1(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.DropDatabase("template1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, &mockDBOpener{})

	err := mgr.DropDatabase("test;drop")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDropDatabase_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mdb := &mockDBOpener{}
	mdb.On("Open", "pgx", mock.Anything).Return(nil, errors.New("connection refused"))
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	err := mgr.DropDatabase("mydb")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid database name")
	assert.NotContains(t, err.Error(), "cannot drop system database")
}

func TestListDatabases_DSNContainsSocketDir(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &system.Paths{Home: tmpDir}
	mdb := &mockDBOpener{}
	mdb.On("Open", "pgx", mock.Anything).Return(nil, errors.New("connection refused"))
	mgr := postgres.NewManagerWithDBOpener(paths, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	_, err := mgr.ListDatabases()
	require.Error(t, err)
	assert.Equal(t, "pgx", mdb.capturedDriver)
	assert.Contains(t, mdb.capturedDSN, "host="+paths.PostgreSQLSocketDir())
	assert.Contains(t, mdb.capturedDSN, "sslmode=disable")
}
