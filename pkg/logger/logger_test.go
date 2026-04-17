package logger_test

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rizkirmdhnnn/lamboserver/pkg/logger"
)

// TestLogger_NewLogger_LogPath verifies that the log path returned by LogPath()
// equals the directory joined with "debug.log".
func TestLogger_NewLogger_LogPath(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	assert.Equal(t, dir+"/debug.log", l.LogPath())
}

// TestLogger_Enable_CreatesLogFile verifies that Enable() creates the log file
// and sets the enabled state to true.
func TestLogger_Enable_CreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	assert.True(t, l.IsEnabled())
	_, err := os.Stat(l.LogPath())
	assert.NoError(t, err, "log file should exist after Enable()")
}

// TestLogger_Enable_Idempotent verifies that calling Enable() twice does not
// return an error and leaves the logger in an enabled state.
func TestLogger_Enable_Idempotent(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	require.NoError(t, l.Enable()) // second call is no-op, must not error
	assert.True(t, l.IsEnabled())
}

// TestLogger_Enable_ErrorOnBadPath verifies that Enable() returns an error when
// the log directory does not exist, and that IsEnabled() remains false.
func TestLogger_Enable_ErrorOnBadPath(t *testing.T) {
	l := logger.NewLogger("/nonexistent/path/that/cannot/be/created")
	err := l.Enable()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open debug log")
	assert.False(t, l.IsEnabled())
}

// TestLogger_Disable_SetsEnabledFalse verifies that Disable() sets enabled to false
// after Enable() was called.
func TestLogger_Disable_SetsEnabledFalse(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Disable()
	assert.False(t, l.IsEnabled())
}

// TestLogger_Disable_WhenAlreadyDisabled verifies that Disable() is a no-op
// when the logger is already disabled and does not panic.
func TestLogger_Disable_WhenAlreadyDisabled(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	l.Disable() // should not panic
	assert.False(t, l.IsEnabled())
}

// TestLogger_Info_WritesToFile verifies that Info() writes a formatted INFO
// message to the log file when the logger is enabled.
func TestLogger_Info_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Info("hello %s", "world")
	l.Disable() // flush / close

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "[INFO] hello world")
}

// TestLogger_Error_WritesToFile verifies that Error() writes a formatted ERROR
// message to the log file when the logger is enabled.
func TestLogger_Error_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Error("bad %s", "thing")
	l.Disable() // flush / close

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "[ERROR] bad thing")
}

// TestLogger_Info_NoOpWhenDisabled verifies that Info() is a no-op when the
// logger is disabled: no file is created.
func TestLogger_Info_NoOpWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	l.Info("should not write") // disabled by default — no panic, no file created
	_, err := os.Stat(l.LogPath())
	assert.True(t, os.IsNotExist(err))
}

// TestLogger_Error_NoOpWhenDisabled verifies that Error() is a no-op when the
// logger is disabled: no file is created.
func TestLogger_Error_NoOpWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	l.Error("should not write")
	_, err := os.Stat(l.LogPath())
	assert.True(t, os.IsNotExist(err))
}

// TestLogger_Action_OKPath verifies that Action() with a nil error writes
// "ACTION <name> -> OK" to the log file.
func TestLogger_Action_OKPath(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Action("InstallPhp", nil)
	l.Disable()

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "ACTION InstallPhp -> OK")
}

// TestLogger_Action_FailedPath verifies that Action() with a non-nil error
// writes "ACTION <name> -> FAILED: <err>" to the log file.
func TestLogger_Action_FailedPath(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Action("InstallPhp", fmt.Errorf("disk full"))
	l.Disable()

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "ACTION InstallPhp -> FAILED: disk full")
}

// TestLogger_ClearLog_TruncatesFile verifies that ClearLog() truncates the log
// file to zero bytes when the logger is enabled.
func TestLogger_ClearLog_TruncatesFile(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Info("some content")
	require.NoError(t, l.ClearLog())

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Empty(t, data)
}

// TestLogger_ClearLog_WhenDisabled_TruncatesViaTruncate verifies that ClearLog()
// truncates via os.Truncate when the logger is not enabled (file handle is nil).
func TestLogger_ClearLog_WhenDisabled_TruncatesViaTruncate(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	// Write some content directly to create the file
	require.NoError(t, os.WriteFile(l.LogPath(), []byte("old content"), 0644))
	require.NoError(t, l.ClearLog()) // logger.file is nil — uses os.Truncate branch
	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	assert.Empty(t, data)
}

// TestLogger_ClearLog_WhenDisabled_FileNotExist verifies that ClearLog() returns
// an error when the log file does not exist and the logger is disabled.
func TestLogger_ClearLog_WhenDisabled_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	// Do not create the file; logger is disabled (file handle nil)
	err := l.ClearLog()
	require.Error(t, err)
}

// TestLogger_LogFormat_HasTimestamp verifies that log entries include a timestamp
// prefix with brackets and the correct level tag.
func TestLogger_LogFormat_HasTimestamp(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())
	l.Info("test")
	l.Disable()

	data, err := os.ReadFile(l.LogPath())
	require.NoError(t, err)
	content := string(data)
	// Verify timestamp format: [2026-... (year bracket prefix)
	assert.True(t, strings.Contains(content, "[20"), "expected timestamp starting with [20xx")
	// Verify level tag present
	assert.Contains(t, content, "[INFO]")
}

// TestLogger_ConcurrentInfoCalls verifies that concurrent calls to Info() do not
// trigger race conditions. Run with -race flag to enable detection.
func TestLogger_ConcurrentInfoCalls(t *testing.T) {
	dir := t.TempDir()
	l := logger.NewLogger(dir)
	require.NoError(t, l.Enable())

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			l.Info("msg %d", n)
		}(i)
	}
	wg.Wait()
	l.Disable()
	// -race flag will catch any mutex violations above
}
