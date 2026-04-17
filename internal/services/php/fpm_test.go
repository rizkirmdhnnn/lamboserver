package php_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFpmManager_SocketPath(t *testing.T) {
	fpm, _, _, paths, _ := newTestFpmManager(t)
	expected := filepath.Join(paths.Home, "php-fpm.sock")
	assert.Equal(t, expected, fpm.SocketPath())
}

func TestFpmManager_Start_Success(t *testing.T) {
	fpm, mockFS, mockLaunchd, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	fpmBin := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", fpmBin).Return(nil, nil)
	mockFS.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLaunchd.On("Install", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == php.FpmServiceLabel &&
			cfg.Type == system.ServiceAgent &&
			cfg.RunAtLoad == true
	})).Return(nil)

	err := fpm.Start()
	require.NoError(t, err)
	mockLaunchd.AssertCalled(t, "Install", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == php.FpmServiceLabel
	}))
}

func TestFpmManager_Start_NoActiveVersion(t *testing.T) {
	fpm, _, _, _, _ := newTestFpmManager(t)
	// Store active is empty by default

	err := fpm.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active PHP version")
}

func TestFpmManager_Start_FpmBinaryMissing(t *testing.T) {
	fpm, mockFS, _, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	fpmBin := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", fpmBin).Return(nil, errors.New("no such file"))

	err := fpm.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "php-fpm binary not found")
}

func TestFpmManager_Start_ConfigWriteFails(t *testing.T) {
	fpm, mockFS, _, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	fpmBin := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", fpmBin).Return(nil, nil)
	mockFS.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("disk full"))

	err := fpm.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate fpm config")
}

func TestFpmManager_Stop(t *testing.T) {
	fpm, _, mockLaunchd, _, _ := newTestFpmManager(t)

	mockLaunchd.On("Uninstall", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == php.FpmServiceLabel && cfg.Type == system.ServiceAgent
	})).Return(nil)

	err := fpm.Stop()
	require.NoError(t, err)
	mockLaunchd.AssertCalled(t, "Uninstall", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == php.FpmServiceLabel
	}))
}

func TestFpmManager_Restart(t *testing.T) {
	fpm, mockFS, mockLaunchd, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	fpmBin := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", fpmBin).Return(nil, nil)
	mockFS.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLaunchd.On("Uninstall", mock.Anything).Return(nil)
	mockLaunchd.On("Install", mock.Anything).Return(nil)

	err := fpm.Restart()
	require.NoError(t, err)
}

func TestFpmManager_Status_Running(t *testing.T) {
	fpm, _, mockLaunchd, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	mockLaunchd.On("IsRunning", php.FpmServiceLabel).Return(true)

	status := fpm.Status()
	assert.True(t, status.Running)
	assert.Equal(t, "8.3", status.Version)
	assert.Equal(t, filepath.Join(paths.Home, "php-fpm.sock"), status.Socket)
}

func TestFpmManager_Status_NotRunning(t *testing.T) {
	fpm, _, mockLaunchd, _, _ := newTestFpmManager(t)

	mockLaunchd.On("IsRunning", php.FpmServiceLabel).Return(false)

	status := fpm.Status()
	assert.False(t, status.Running)
}

func TestFpmManager_GenerateConfig_Content(t *testing.T) {
	fpm, mockFS, mockLaunchd, paths, store := newTestFpmManager(t)

	require.NoError(t, store.SetActivePhpVersion("8.3"))

	fpmBin := paths.PhpVersionDir("8.3") + "/sbin/php-fpm"
	mockFS.On("Stat", fpmBin).Return(nil, nil)

	// Capture the written config content
	var capturedConfig []byte
	mockFS.On("WriteFile", mock.Anything, mock.MatchedBy(func(data []byte) bool {
		capturedConfig = data
		return true
	}), mock.Anything).Return(nil)

	mockLaunchd.On("Install", mock.Anything).Return(nil)

	err := fpm.Start()
	require.NoError(t, err)
	require.NotNil(t, capturedConfig)

	configStr := string(capturedConfig)
	assert.Contains(t, configStr, "[lamboserver]")
	assert.Contains(t, configStr, "listen = "+filepath.Join(paths.Home, "php-fpm.sock"))
	assert.Contains(t, configStr, "pm = dynamic")
	assert.True(t, strings.Contains(configStr, "pm.max_children = 10"), "config should contain pm.max_children = 10")
}
