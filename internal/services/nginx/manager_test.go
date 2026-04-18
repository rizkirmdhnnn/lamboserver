// Package nginx_test contains unit tests for the nginx package.
package nginx_test

import (
	"errors"
	"os"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/services/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- IsDaemonInstalled tests ---

func TestManager_IsDaemonInstalled_True(t *testing.T) {
	mgr, mockFS, _, _, _ := newTestManager(t)
	plistPath := "/Library/LaunchDaemons/" + nginx.ServiceLabel + ".plist"

	mockFS.On("Stat", plistPath).Return(nil, nil)

	assert.True(t, mgr.IsDaemonInstalled())
	mockFS.AssertExpectations(t)
}

func TestManager_IsDaemonInstalled_False(t *testing.T) {
	mgr, mockFS, _, _, _ := newTestManager(t)
	plistPath := "/Library/LaunchDaemons/" + nginx.ServiceLabel + ".plist"

	mockFS.On("Stat", plistPath).Return(nil, os.ErrNotExist)

	assert.False(t, mgr.IsDaemonInstalled())
	mockFS.AssertExpectations(t)
}

// --- Stop test ---

func TestManager_Stop(t *testing.T) {
	mgr, _, mockLaunchd, _, _ := newTestManager(t)

	mockLaunchd.On("Uninstall", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == nginx.ServiceLabel && cfg.Type == system.ServiceDaemon
	})).Return(nil)

	err := mgr.Stop()
	require.NoError(t, err)
	mockLaunchd.AssertExpectations(t)
}

// --- Reload tests ---

func TestManager_Reload(t *testing.T) {
	mgr, _, _, mockHelper, _ := newTestManager(t)

	mockHelper.On("Run", []string{"reload-nginx"}).Return("", nil)

	err := mgr.Reload()
	require.NoError(t, err)
	mockHelper.AssertExpectations(t)
}

func TestManager_Reload_Error(t *testing.T) {
	mgr, _, _, mockHelper, _ := newTestManager(t)

	reloadErr := errors.New("helper: reload-nginx failed")
	mockHelper.On("Run", []string{"reload-nginx"}).Return("", reloadErr)

	err := mgr.Reload()
	require.Error(t, err)
	assert.ErrorIs(t, err, reloadErr)
	mockHelper.AssertExpectations(t)
}

// --- Status tests ---

func TestManager_Status_Running(t *testing.T) {
	mgr, _, mockLaunchd, _, _ := newTestManager(t)

	mockLaunchd.On("IsRunning", nginx.ServiceLabel).Return(true)

	status := mgr.Status()
	assert.True(t, status.Running)
	mockLaunchd.AssertExpectations(t)
}

func TestManager_Status_NotRunning(t *testing.T) {
	mgr, _, mockLaunchd, _, _ := newTestManager(t)

	mockLaunchd.On("IsRunning", nginx.ServiceLabel).Return(false)

	status := mgr.Status()
	assert.False(t, status.Running)
	mockLaunchd.AssertExpectations(t)
}

// --- Start test ---

func TestManager_Start_BinaryExists(t *testing.T) {
	mgr, mockFS, mockLaunchd, _, paths := newTestManager(t)

	// Create the nginx binary file so BinaryLocator.IsInstalled() returns true.
	nginxBin := paths.NginxBin()
	require.NoError(t, os.MkdirAll(paths.BinDir(), 0755))
	require.NoError(t, os.WriteFile(nginxBin, []byte("#!/bin/sh\n"), 0755))

	// Create required directories for EnsureLogFile calls.
	require.NoError(t, os.MkdirAll(paths.LogsDir(), 0755))
	require.NoError(t, os.MkdirAll(paths.NginxLogsDir(), 0755))
	require.NoError(t, os.MkdirAll(paths.DefaultSiteDir(), 0755))

	// Capture all WriteFile calls (nginx.conf, fastcgi_params, etc.).
	captureWriteFile(mockFS)

	// fs.Stat for mime types: return not-exist so writeMimeTypes is called.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, os.ErrNotExist)

	// fs.Stat for default HTML: return nil (exists) so only writeDefaultSiteConf is called.
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, nil)

	// launchd.Install should be called with correct config.
	mockLaunchd.On("Install", mock.MatchedBy(func(cfg system.ServiceConfig) bool {
		return cfg.Label == nginx.ServiceLabel &&
			cfg.Type == system.ServiceDaemon &&
			cfg.KeepAlive == true
	})).Return(nil)

	err := mgr.Start()
	require.NoError(t, err)
	mockLaunchd.AssertExpectations(t)
}

// --- EnsureConfig tests ---

func TestManager_EnsureConfig_MasterConfig(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types already exist, so writeMimeTypes is skipped.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, nil)

	// Default site HTML exists, so only conf is written.
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, nil)

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	confData, ok := writes[paths.NginxConf()]
	require.True(t, ok, "nginx.conf should have been written")
	conf := string(confData)

	assert.Contains(t, conf, "worker_processes auto")
	assert.Contains(t, conf, "error_log "+paths.NginxErrorLog())
	assert.Contains(t, conf, "access_log "+paths.NginxAccessLog())
	assert.Contains(t, conf, "include "+paths.NginxMimeTypes())
	assert.Contains(t, conf, "include "+paths.NginxSitesDir()+"/*.conf")
	assert.Contains(t, conf, "client_max_body_size 512M")
}

func TestManager_EnsureConfig_FastCGIParams(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types exist, default site HTML exists.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, nil)
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, nil)

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	fastcgiData, ok := writes[paths.NginxFastCGIParams()]
	require.True(t, ok, "fastcgi_params should have been written")
	params := string(fastcgiData)

	assert.Contains(t, params, "SCRIPT_NAME")
	assert.Contains(t, params, "REQUEST_URI")
	assert.Contains(t, params, "DOCUMENT_ROOT")
	assert.Contains(t, params, "SERVER_SOFTWARE    LamboServer")
}

func TestManager_EnsureConfig_MimeTypes_WhenMissing(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types do NOT exist — writeMimeTypes should be called.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, os.ErrNotExist)

	// Note: when mime types are missing, writeDefaultSite is NOT called (returns early after writeMimeTypes).

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	mimeData, ok := writes[paths.NginxMimeTypes()]
	require.True(t, ok, "mime.types should have been written")
	mime := string(mimeData)

	assert.Contains(t, mime, "text/html")
	assert.Contains(t, mime, "application/javascript")
	assert.Contains(t, mime, "application/x-httpd-php")
}

func TestManager_EnsureConfig_MimeTypes_WhenExists(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types exist — writeMimeTypes should NOT be called.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, nil)

	// Default site HTML also exists.
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, nil)

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	// Verify mime.types was NOT written.
	_, written := writes[paths.NginxMimeTypes()]
	assert.False(t, written, "mime.types should NOT be written when it already exists")
}

func TestManager_EnsureConfig_DefaultSite(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types exist.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, nil)

	// Default site HTML does NOT exist — both HTML and conf should be written.
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, os.ErrNotExist)

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	htmlPath := paths.DefaultSiteDir() + "/index.html"
	htmlData, ok := writes[htmlPath]
	require.True(t, ok, "default site HTML should have been written")
	assert.Contains(t, string(htmlData), "LamboServer")

	confPath := paths.NginxSitesDir() + "/00-default.conf"
	confData, ok := writes[confPath]
	require.True(t, ok, "00-default.conf should have been written")
	assert.Contains(t, string(confData), "listen 127.0.0.1:80 default_server")
}

func TestManager_EnsureConfig_DefaultSiteExists(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writes := captureWriteFile(mockFS)

	// Mime types exist.
	mockFS.On("Stat", paths.NginxMimeTypes()).Return(nil, nil)

	// Default site HTML exists — HTML should NOT be re-written, but conf should be.
	mockFS.On("Stat", paths.DefaultSiteDir()+"/index.html").Return(nil, nil)

	err := mgr.EnsureConfig()
	require.NoError(t, err)

	// HTML should NOT be re-written.
	htmlPath := paths.DefaultSiteDir() + "/index.html"
	_, htmlWritten := writes[htmlPath]
	assert.False(t, htmlWritten, "default site HTML should NOT be re-written when it exists")

	// 00-default.conf should still be written.
	confPath := paths.NginxSitesDir() + "/00-default.conf"
	confData, ok := writes[confPath]
	require.True(t, ok, "00-default.conf should still be written")
	assert.Contains(t, string(confData), "listen 127.0.0.1:80 default_server")
}

func TestManager_EnsureConfig_WriteError(t *testing.T) {
	mgr, mockFS, _, _, paths := newTestManager(t)

	writeErr := errors.New("disk full")
	mockFS.On("WriteFile", paths.NginxConf(), mock.Anything, mock.Anything).Return(writeErr)

	err := mgr.EnsureConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write nginx.conf")
}
