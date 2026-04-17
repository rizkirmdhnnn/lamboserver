package system

import (
	"os/exec"
	"strings"
)

// RealCmdRunner delegates command execution to os/exec.
// Satisfies CommandRunner interfaces defined in php and node packages.
type RealCmdRunner struct{}

// Run executes the named command with the given arguments and returns combined stdout/stderr output.
func (RealCmdRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// LookPath searches for the named executable in the directories named by PATH.
func (RealCmdRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// RealAdminRunner delegates to system.RunWithAdminPrivileges.
// Satisfies AdminRunner interfaces defined in cert and dns packages.
type RealAdminRunner struct{}

// RunWithPrivileges executes the given shell command with macOS admin privileges
// via an osascript dialog. Used for operations requiring root access.
func (RealAdminRunner) RunWithPrivileges(command string) error {
	return RunWithAdminPrivileges(command)
}
