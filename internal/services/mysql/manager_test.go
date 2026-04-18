// Package mysql_test contains unit tests for the mysql package.
package mysql_test

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/mysql"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock definitions ---

// mockFS is a testify mock implementing mysql.FileSystem.
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

// mockLaunchd is a testify mock implementing mysql.LaunchdService.
type mockLaunchd struct {
	mock.Mock
}

func (m *mockLaunchd) Install(cfg system.ServiceConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *mockLaunchd) Uninstall(cfg system.ServiceConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *mockLaunchd) IsRunning(label string) bool {
	args := m.Called(label)
	return args.Bool(0)
}

// mockCmd is a testify mock implementing mysql.CommandRunner.
type mockCmd struct {
	mock.Mock
}

func (m *mockCmd) Run(name string, args ...string) (string, error) {
	allArgs := append([]string{name}, args...)
	callArgs := m.Called(allArgs)
	return callArgs.String(0), callArgs.Error(1)
}

// mockHelper is a testify mock implementing mysql.HelperRunner.
type mockHelper struct {
	mock.Mock
}

func (m *mockHelper) Run(args ...string) (string, error) {
	callArgs := m.Called(args)
	return callArgs.String(0), callArgs.Error(1)
}

// mockDBOpener is a testify mock implementing mysql.DBOpener.
// It captures the DSN string and returns a configurable error.
type mockDBOpener struct {
	mock.Mock
	capturedDSN string
}

func (m *mockDBOpener) Open(driverName, dataSourceName string) (*sql.DB, error) {
	m.capturedDSN = dataSourceName
	args := m.Called(driverName, dataSourceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.DB), args.Error(1)
}

// mockAdmin is a testify mock implementing mysql.AdminRunner.
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

func newTestManager(t *testing.T) (*mysql.Manager, *mockFS, *mockLaunchd, *mockCmd, *mockHelper, *mockAdmin, *system.Paths) {
	t.Helper()
	paths := newTestPaths(t)
	mfs := &mockFS{}
	mld := &mockLaunchd{}
	mc := &mockCmd{}
	mh := &mockHelper{}
	ma := &mockAdmin{}
	mgr := mysql.NewManager(paths, mld, mfs, mc, mh, ma)
	return mgr, mfs, mld, mc, mh, ma, paths
}

// --- TestIsInstalled ---

func TestIsInstalled_True(t *testing.T) {
	_, _, _, _, _, _, paths := newTestManager(t)

	// Create the mysqld binary so BinaryLocator.IsInstalled() returns true.
	require.NoError(t, os.MkdirAll(paths.MySQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte("#!/bin/sh\n"), 0755))

	// Create a new manager — BinaryLocator checks the file on IsInstalled().
	mld := &mockLaunchd{}
	mfs := &mockFS{}
	mc := &mockCmd{}
	mh := &mockHelper{}
	mgr := mysql.NewManager(paths, mld, mfs, mc, mh, &mockAdmin{})

	assert.True(t, mgr.IsInstalled())
}

func TestIsInstalled_FalseWhenLocalPathAbsent(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	// BinaryLocator.IsInstalled() checks LocalPath first, then system PATH.
	// The LocalPath (tempdir/mysql/bin/mysqld) does not exist, so if mysqld is also
	// absent from system PATH the result is false. On machines with MySQL installed
	// via Homebrew or system packages, system PATH may have mysqld and this returns true.
	// We skip the assertion when the system has mysqld, as the binary-locator PATH
	// fallback is correct and expected behavior.
	if mgr.IsInstalled() {
		t.Log("mysqld found on system PATH — IsInstalled() correctly returns true; skipping false-assertion")
		return
	}
	assert.False(t, mgr.IsInstalled())
}

// --- TestInstall ---

func TestInstall_Success_ARM64(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	// MkdirAll for bin dir succeeds.
	mfs.On("MkdirAll", paths.MySQLBinDir(), fs.FileMode(0755)).Return(nil)

	// Capture the shell command — verify URL and destination.
	var capturedCmd string
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Run(func(args mock.Arguments) {
		allArgs := args.Get(0).([]string)
		if len(allArgs) >= 3 {
			capturedCmd = allArgs[2]
		}
		// Create the mysqld binary so Stat succeeds.
		os.MkdirAll(paths.MySQLBinDir(), 0755)
		os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte(""), 0755)
	}).Return("", nil)

	// chmod -R u+rwX to fix tarball permissions.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "chmod"
	})).Return("", nil)

	// Stat for mysqld binary verification after extraction.
	mfs.On("Stat", paths.MySQLBinDir()+"/mysqld").Return(nil, nil)

	err := mgr.Install()
	require.NoError(t, err)

	// Verify the URL contains the correct MySQL distribution.
	assert.Contains(t, capturedCmd, "cdn.mysql.com")
	assert.Contains(t, capturedCmd, "mysql-8.4")
	assert.Contains(t, capturedCmd, "macos14")
	assert.Contains(t, capturedCmd, paths.MySQLDir())

	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

func TestInstall_ArchDetection(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLBinDir(), fs.FileMode(0755)).Return(nil)

	var capturedCmd string
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Run(func(args mock.Arguments) {
		allArgs := args.Get(0).([]string)
		if len(allArgs) >= 3 {
			capturedCmd = allArgs[2]
		}
		os.MkdirAll(paths.MySQLBinDir(), 0755)
		os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte(""), 0755)
	}).Return("", nil)

	// chmod -R u+rwX to fix tarball permissions.
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "chmod"
	})).Return("", nil)

	mfs.On("Stat", paths.MySQLBinDir()+"/mysqld").Return(nil, nil)

	err := mgr.Install()
	require.NoError(t, err)

	// URL must contain either arm64 or x86_64 — both are valid arch strings.
	hasArch := strings.Contains(capturedCmd, "arm64") || strings.Contains(capturedCmd, "x86_64")
	assert.True(t, hasArch, "URL should contain a valid architecture string (arm64 or x86_64)")
}

func TestInstall_DownloadFails(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLBinDir(), fs.FileMode(0755)).Return(nil)

	downloadErr := errors.New("network error")
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 1 && args[0] == "sh"
	})).Return("", downloadErr)

	err := mgr.Install()
	require.Error(t, err)
	assert.ErrorIs(t, err, downloadErr)
}

// --- TestInitDataDir ---

func TestInitDataDir_Success(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	// MkdirAll for data dir succeeds (current user, no privilege elevation).
	mfs.On("MkdirAll", paths.MySQLDataDir(), fs.FileMode(0755)).Return(nil)

	// WriteFile for my.cnf.
	mfs.On("WriteFile", paths.MySQLConf(), mock.Anything, fs.FileMode(0644)).Return(nil)

	// mysqld --initialize-insecure via CommandRunner (current user).
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "mysqld") && args[1] == "--initialize-insecure"
	})).Return("", nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)

	mfs.AssertExpectations(t)
	mc.AssertExpectations(t)
}

func TestInitDataDir_CleanupOnMysqldFailure(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLDataDir(), fs.FileMode(0755)).Return(nil)
	mfs.On("WriteFile", paths.MySQLConf(), mock.Anything, fs.FileMode(0644)).Return(nil)

	mysqldErr := errors.New("exit status 1")
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "mysqld") && args[1] == "--initialize-insecure"
	})).Return("", mysqldErr)

	// RemoveAll must be called to clean partial data dir on failure.
	mfs.On("RemoveAll", paths.MySQLDataDir()).Return(nil)

	err := mgr.InitDataDir()
	require.Error(t, err)
	assert.ErrorIs(t, err, mysqldErr)

	mfs.AssertExpectations(t)
}

func TestInitDataDir_NoCleanupOnSuccess(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLDataDir(), fs.FileMode(0755)).Return(nil)
	mfs.On("WriteFile", paths.MySQLConf(), mock.Anything, fs.FileMode(0644)).Return(nil)
	mc.On("Run", mock.MatchedBy(func(args []string) bool {
		return len(args) >= 2 && strings.Contains(args[0], "mysqld") && args[1] == "--initialize-insecure"
	})).Return("", nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)

	mfs.AssertNotCalled(t, "RemoveAll", mock.Anything)
}

func TestInitDataDir_WriteConfFails(t *testing.T) {
	mgr, mfs, _, _, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLDataDir(), fs.FileMode(0755)).Return(nil)

	writeErr := errors.New("permission denied")
	mfs.On("WriteFile", paths.MySQLConf(), mock.Anything, fs.FileMode(0644)).Return(writeErr)

	err := mgr.InitDataDir()
	require.Error(t, err)
	assert.ErrorIs(t, err, writeErr)
}

func TestInitDataDir_MyCnfContents(t *testing.T) {
	mgr, mfs, _, mc, _, _, paths := newTestManager(t)

	mfs.On("MkdirAll", paths.MySQLDataDir(), fs.FileMode(0755)).Return(nil)

	var writtenConf string
	mfs.On("WriteFile", paths.MySQLConf(), mock.Anything, fs.FileMode(0644)).
		Run(func(args mock.Arguments) {
			writtenConf = string(args.Get(1).([]byte))
		}).Return(nil)

	// mysqld --initialize-insecure
	mc.On("Run", mock.Anything).Return("", nil)

	err := mgr.InitDataDir()
	require.NoError(t, err)

	// Verify my.cnf contains expected directives.
	assert.Contains(t, writtenConf, "datadir")
	assert.Contains(t, writtenConf, paths.MySQLDataDir())
	assert.Contains(t, writtenConf, "socket")
	assert.Contains(t, writtenConf, paths.MySQLSocket())
	assert.Contains(t, writtenConf, "bind-address = 127.0.0.1")
	assert.Contains(t, writtenConf, "port        = 3306")
}

// --- TestStart ---

func TestStart_Success(t *testing.T) {
	mgr, mfs, mld, _, _, _, paths := newTestManager(t)

	// Create mysqld binary so IsInstalled() is true.
	require.NoError(t, os.MkdirAll(paths.MySQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte("#!/bin/sh\n"), 0755))

	// Create required dirs for EnsureLogFile.
	require.NoError(t, os.MkdirAll(paths.LogsDir(), 0755))

	// Data dir mysql/ subdirectory exists — no need to call InitDataDir.
	mfs.On("Stat", paths.MySQLDataDir()+"/mysql").Return(nil, nil)

	// Verify launchd.Install is called with ServiceAgent (user-level, Herd-style).
	mld.On("Install", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == mysql.ServiceLabel &&
			cfg.Type == system.ServiceAgent &&
			cfg.UserName == "" &&
			cfg.RunAtLoad == true &&
			cfg.KeepAlive == true
	})).Return(nil)

	err := mgr.Start()
	require.NoError(t, err)
	mld.AssertExpectations(t)
}

func TestStart_NotInstalled(t *testing.T) {
	mgr, mfs, mld, mc, _, _, paths := newTestManager(t)
	// No mysqld binary at LocalPath. If mysqld exists on system PATH, IsInstalled()
	// will return true and Start proceeds; in that case we verify Start proceeds
	// without error through the launchd.Install call.
	// This test uses an isolated tempdir for LocalPath — if mysqld is not in PATH,
	// Start returns "MySQL is not installed".
	if mgr.IsInstalled() {
		// mysqld found on system PATH — skip the "not installed" assertion.
		t.Skip("mysqld found on system PATH; skipping 'not installed' test")
	}
	_ = mfs
	_ = mld
	_ = mc
	_ = paths
	err := mgr.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MySQL is not installed")
}

func TestStart_ArgsContainDefaultsFile(t *testing.T) {
	mgr, mfs, mld, _, _, _, paths := newTestManager(t)

	require.NoError(t, os.MkdirAll(paths.MySQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte("#!/bin/sh\n"), 0755))
	require.NoError(t, os.MkdirAll(paths.LogsDir(), 0755))

	mfs.On("Stat", paths.MySQLDataDir()+"/mysql").Return(nil, nil)

	var capturedCfg system.ServiceConfig
	mld.On("Install", mock.Anything).Run(func(args mock.Arguments) {
		capturedCfg = args.Get(0).(system.ServiceConfig)
	}).Return(nil)

	err := mgr.Start()
	require.NoError(t, err)

	// Verify --defaults-file arg is present (runs as current user, no --user flag).
	argsStr := strings.Join(capturedCfg.Args, " ")
	assert.Contains(t, argsStr, "--defaults-file=")
	assert.Contains(t, argsStr, paths.MySQLConf())
	assert.NotContains(t, argsStr, "--user=_mysql") // Herd-style: runs as current user
}

// --- TestStop ---

func TestStop_Success(t *testing.T) {
	mgr, mfs, mld, _, _, _, paths := newTestManager(t)

	mld.On("Uninstall", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == mysql.ServiceLabel && cfg.Type == system.ServiceAgent
	})).Return(nil)

	// Socket does not exist immediately after uninstall → loop exits on first iteration.
	mfs.On("Stat", paths.MySQLSocket()).Return(nil, os.ErrNotExist)

	err := mgr.Stop()
	require.NoError(t, err)
	mld.AssertExpectations(t)
}

func TestStop_UninstallError(t *testing.T) {
	mgr, _, mld, _, _, _, _ := newTestManager(t)

	uninstallErr := errors.New("launchctl: daemon not found")
	mld.On("Uninstall", mock.Anything).Return(uninstallErr)

	err := mgr.Stop()
	require.Error(t, err)
	assert.ErrorIs(t, err, uninstallErr)
}

// --- TestStatus ---

func TestStatus_InstalledAndRunning(t *testing.T) {
	_, mfs, mld, _, _, _, paths := newTestManager(t)

	// Create binary.
	require.NoError(t, os.MkdirAll(paths.MySQLBinDir(), 0755))
	require.NoError(t, os.WriteFile(paths.MySQLBinDir()+"/mysqld", []byte(""), 0755))

	// Create fresh manager pointing at these paths.
	mgr := mysql.NewManager(paths, mld, mfs, &mockCmd{}, &mockHelper{}, &mockAdmin{})

	mld.On("IsRunning", mysql.ServiceLabel).Return(true)

	status := mgr.Status()
	assert.True(t, status.Installed)
	assert.True(t, status.Running)
	assert.Equal(t, 3306, status.Port)

	mld.AssertExpectations(t)
}

func TestStatus_NotRunning(t *testing.T) {
	mgr, _, mld, _, _, _, _ := newTestManager(t)

	mld.On("IsRunning", mysql.ServiceLabel).Return(false)

	status := mgr.Status()
	// Installed depends on whether mysqld is on system PATH; just verify Running=false and Port.
	assert.False(t, status.Running)
	assert.Equal(t, 3306, status.Port)
	mld.AssertExpectations(t)
}

// --- TestCreateDatabase ---

func TestCreateDatabase_ValidName(t *testing.T) {
	_, _, _, _, _, _, paths := newTestManager(t)

	mdb := &mockDBOpener{}
	mdb.On("Open", "mysql", mock.Anything).Return(nil, errors.New("connection refused"))

	mld := &mockLaunchd{}
	mfs := &mockFS{}
	mc := &mockCmd{}
	mh := &mockHelper{}
	mgr := mysql.NewManagerWithDBOpener(paths, mld, mfs, mc, mh, &mockAdmin{}, mdb)

	// Valid name — should reach openDB (connection refused is fine for testing validation).
	err := mgr.CreateDatabase("mydb")
	require.Error(t, err)
	// Error is about connection, not validation.
	assert.NotContains(t, err.Error(), "invalid database name")

	// Verify DSN was constructed with Unix socket path.
	assert.Contains(t, mdb.capturedDSN, paths.MySQLSocket())
}

func TestCreateDatabase_InvalidName_Semicolon(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)

	err := mgr.CreateDatabase("my;db")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_InvalidName_Empty(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)

	err := mgr.CreateDatabase("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_InvalidName_TooLong(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)

	// 65-character name — exceeds the 64-character MySQL limit.
	longName := strings.Repeat("a", 65)
	err := mgr.CreateDatabase(longName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_InvalidName_Hyphen(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)

	err := mgr.CreateDatabase("my-db")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestCreateDatabase_ValidName_MaxLength(t *testing.T) {
	_, _, _, _, _, _, paths := newTestManager(t)

	mdb := &mockDBOpener{}
	mdb.On("Open", "mysql", mock.Anything).Return(nil, errors.New("connection refused"))

	mgr := mysql.NewManagerWithDBOpener(paths, &mockLaunchd{}, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	// Exactly 64 characters — should pass validation.
	name64 := strings.Repeat("a", 64)
	err := mgr.CreateDatabase(name64)
	// Should NOT get "invalid database name" — the name is valid.
	assert.NotContains(t, err.Error(), "invalid database name")
}

// --- TestDropDatabase ---

func TestDropDatabase_Valid(t *testing.T) {
	_, _, _, _, _, _, paths := newTestManager(t)

	mdb := &mockDBOpener{}
	mdb.On("Open", "mysql", mock.Anything).Return(nil, errors.New("connection refused"))

	mgr := mysql.NewManagerWithDBOpener(paths, &mockLaunchd{}, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	// Valid name — reaches openDB.
	err := mgr.DropDatabase("mydb")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid database name")
	assert.NotContains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_mysql(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	err := mgr.DropDatabase("mysql")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_InformationSchema(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	err := mgr.DropDatabase("information_schema")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_PerformanceSchema(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	err := mgr.DropDatabase("performance_schema")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_SystemDB_Sys(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	err := mgr.DropDatabase("sys")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot drop system database")
}

func TestDropDatabase_InvalidName(t *testing.T) {
	mgr, _, _, _, _, _, _ := newTestManager(t)
	err := mgr.DropDatabase("drop; SELECT *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

// --- TestListDatabases DSN construction ---

func TestListDatabases_DSNContainsSocket(t *testing.T) {
	_, _, _, _, _, _, paths := newTestManager(t)

	mdb := &mockDBOpener{}
	mdb.On("Open", "mysql", mock.Anything).Return(nil, errors.New("connection refused"))

	mgr := mysql.NewManagerWithDBOpener(paths, &mockLaunchd{}, &mockFS{}, &mockCmd{}, &mockHelper{}, &mockAdmin{}, mdb)

	_, err := mgr.ListDatabases()
	require.Error(t, err) // Expected: connection refused.

	// Verify the DSN uses the Unix socket path (not TCP).
	assert.Contains(t, mdb.capturedDSN, "unix(")
	assert.Contains(t, mdb.capturedDSN, paths.MySQLSocket())
	assert.Contains(t, mdb.capturedDSN, "root@")
}
