package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	helperScriptName = "lambo-helper"
	sudoersFile      = "/etc/sudoers.d/lamboserver"
)

// Helper manages privileged operations without repeated password prompts.
// On first setup, it installs a sudoers entry that allows the current user
// to run lambo-helper as root without a password.
type Helper struct {
	paths *Paths
}

func NewHelper(paths *Paths) *Helper {
	return &Helper{paths: paths}
}

// IsInstalled checks if the helper and sudoers entry are set up
func (h *Helper) IsInstalled() bool {
	_, err1 := os.Stat(h.helperPath())
	_, err2 := os.Stat(sudoersFile)
	return err1 == nil && err2 == nil
}

// Install sets up the helper script and sudoers entry. Asks admin password ONCE.
func (h *Helper) Install() error {
	// Always rewrite helper script to keep it up to date
	if err := h.writeHelperScript(); err != nil {
		return fmt.Errorf("failed to write helper script: %w", err)
	}

	if h.IsInstalled() {
		return nil
	}

	// Install sudoers entry (requires admin password - only time we ask)
	username := os.Getenv("USER")
	if username == "" {
		return fmt.Errorf("could not determine username")
	}

	sudoersContent := fmt.Sprintf(
		"%s ALL=(root) NOPASSWD: %s\n",
		username, h.helperPath(),
	)

	tmpFile := os.TempDir() + "/lamboserver-sudoers"
	if err := os.WriteFile(tmpFile, []byte(sudoersContent), 0644); err != nil {
		return err
	}

	// Install sudoers file + resolver in one admin prompt
	cmd := fmt.Sprintf(
		"cp %s %s && chmod 0440 %s && chown root:wheel %s && mkdir -p /etc/resolver && echo 'nameserver 127.0.0.1' > /etc/resolver/test",
		tmpFile, sudoersFile, sudoersFile, sudoersFile,
	)
	if err := RunWithAdminPrivileges(cmd); err != nil {
		return fmt.Errorf("failed to install sudoers entry: %w", err)
	}

	// Validate sudoers file
	if _, err := RunCommand("sudo", "-n", h.helperPath(), "status"); err != nil {
		// If validation fails, the sudoers entry might be invalid - remove it
		RunWithAdminPrivileges("rm -f " + sudoersFile)
		return fmt.Errorf("sudoers validation failed: %w", err)
	}

	return nil
}

// Uninstall removes the helper script and sudoers entry
func (h *Helper) Uninstall() error {
	RunWithAdminPrivileges("rm -f " + sudoersFile)
	os.Remove(h.helperPath())
	return nil
}

// UpdateScript rewrites the helper script to pick up code changes
func (h *Helper) UpdateScript() error {
	return h.writeHelperScript()
}

// Run executes a privileged command via the helper (no password needed after setup)
func (h *Helper) Run(args ...string) (string, error) {
	cmdArgs := append([]string{"-n", h.helperPath()}, args...)
	cmd := exec.Command("sudo", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("helper command failed: %w\nOutput: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (h *Helper) helperPath() string {
	return h.paths.BinDir() + "/" + helperScriptName
}

func (h *Helper) writeHelperScript() error {
	script := fmt.Sprintf(`#!/bin/bash
# LamboServer privileged helper
# This script runs as root via sudoers for service management

LAMBO_HOME="%s"
LAUNCH_DAEMONS="/Library/LaunchDaemons"

case "$1" in
    install-daemon)
        PLIST_SRC="$2"
        PLIST_NAME=$(basename "$PLIST_SRC")
        PLIST_DST="$LAUNCH_DAEMONS/$PLIST_NAME"
        # Unload old daemon if exists
        launchctl unload "$PLIST_DST" 2>/dev/null
        cp "$PLIST_SRC" "$PLIST_DST"
        chown root:wheel "$PLIST_DST"
        chmod 644 "$PLIST_DST"
        launchctl load -w "$PLIST_DST"
        ;;
    uninstall-daemon)
        LABEL="$2"
        PLIST="$LAUNCH_DAEMONS/$LABEL.plist"
        if [ -f "$PLIST" ]; then
            launchctl unload "$PLIST" 2>/dev/null
            rm -f "$PLIST"
        fi
        ;;
    reload-nginx)
        "$LAMBO_HOME/bin/nginx" -s reload -c "$LAMBO_HOME/nginx/nginx.conf" 2>/dev/null
        ;;
    status)
        echo "ok"
        ;;
    *)
        echo "Usage: lambo-helper {install-daemon|uninstall-daemon|reload-nginx|status}" >&2
        exit 1
        ;;
esac
`, h.paths.Home)

	if err := os.WriteFile(h.helperPath(), []byte(script), 0755); err != nil {
		return err
	}
	return nil
}
