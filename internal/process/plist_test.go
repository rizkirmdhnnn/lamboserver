package process_test

import (
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePlist_DaemonDefaultsToRoot(t *testing.T) {
	cfg := process.PlistConfig{
		Label:    "com.test.daemon",
		Program:  "/usr/bin/test",
		IsDaemon: true,
		// UserName intentionally empty
	}
	plist, err := process.GeneratePlist(cfg)
	require.NoError(t, err)
	assert.Contains(t, plist, "<key>UserName</key>")
	assert.Contains(t, plist, "<string>root</string>")
}

func TestGeneratePlist_DaemonCustomUserName(t *testing.T) {
	cfg := process.PlistConfig{
		Label:    "com.test.mariadb",
		Program:  "/usr/local/bin/mysqld",
		IsDaemon: true,
		UserName: "_mysql",
	}
	plist, err := process.GeneratePlist(cfg)
	require.NoError(t, err)
	assert.Contains(t, plist, "<key>UserName</key>")
	assert.Contains(t, plist, "<string>_mysql</string>")
	assert.NotContains(t, plist, "<string>root</string>")
}

func TestGeneratePlist_AgentOmitsUserName(t *testing.T) {
	cfg := process.PlistConfig{
		Label:    "com.test.agent",
		Program:  "/usr/bin/test",
		IsDaemon: false,
	}
	plist, err := process.GeneratePlist(cfg)
	require.NoError(t, err)
	assert.NotContains(t, plist, "<key>UserName</key>")
}

func TestGeneratePlist_AgentIgnoresUserNameField(t *testing.T) {
	cfg := process.PlistConfig{
		Label:    "com.test.agent",
		Program:  "/usr/bin/test",
		IsDaemon: false,
		UserName: "_mysql",
	}
	plist, err := process.GeneratePlist(cfg)
	require.NoError(t, err)
	// UserName should NOT appear because IsDaemon is false
	assert.NotContains(t, plist, "<key>UserName</key>")
}
