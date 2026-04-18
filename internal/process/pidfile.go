package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// PIDFileManager implements ProcessManager using PID files for services
// that manage their own processes (e.g., PostgreSQL via pg_ctl).
// It tracks running processes by writing PID files to a configured directory.
type PIDFileManager struct {
	pidDir string
}

// NewPIDFileManager creates a PIDFileManager that stores PID files in the given directory.
func NewPIDFileManager(pidDir string) *PIDFileManager {
	return &PIDFileManager{pidDir: pidDir}
}

func (p *PIDFileManager) pidPath(name string) string {
	return fmt.Sprintf("%s/%s.pid", p.pidDir, name)
}

// Start launches the process described by cfg and writes a PID file.
func (p *PIDFileManager) Start(name string, cfg ProcessConfig) error {
	cmd := exec.Command(cfg.Program, cfg.Args...)
	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	pidFile := cfg.PIDFile
	if pidFile == "" {
		pidFile = p.pidPath(name)
	}
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file for %s: %w", name, err)
	}

	return nil
}

// Stop reads the PID file and sends SIGTERM to the process.
func (p *PIDFileManager) Stop(name string) error {
	pidFile := p.pidPath(name)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read PID file for %s: %w", name, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("invalid PID in file for %s: %w", name, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidFile)
		return nil
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		os.Remove(pidFile)
		return nil
	}

	os.Remove(pidFile)
	return nil
}

// Restart stops then starts the service.
func (p *PIDFileManager) Restart(name string, cfg ProcessConfig) error {
	if err := p.Stop(name); err != nil {
		return err
	}
	return p.Start(name, cfg)
}

// IsRunning checks if the process from the PID file is still alive.
func (p *PIDFileManager) IsRunning(name string) bool {
	pidFile := p.pidPath(name)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, Signal(0) checks if process exists without sending a signal.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
