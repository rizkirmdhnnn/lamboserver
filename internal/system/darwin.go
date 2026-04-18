package system

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// RunWithAdminPrivileges executes a command with macOS admin privileges dialog
func RunWithAdminPrivileges(command string) error {
	escaped := strings.ReplaceAll(command, `"`, `\"`)
	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, escaped)
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("admin command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RunCommand executes a shell command and returns output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// GetArchitecture returns the current CPU architecture
func GetArchitecture() string {
	return runtime.GOARCH
}

// CommandExists checks if a command is available in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
