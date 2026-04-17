// Package process provides a platform-neutral abstraction for managing service
// processes. It defines the ProcessManager interface and ProcessConfig struct
// that services use instead of directly calling launchd, systemd, or other
// OS-specific process managers.
package process

// ProcessConfig holds all parameters needed to start and manage a service process.
type ProcessConfig struct {
	// Label is a unique identifier for the service (e.g., "com.my-dev-env.nginx").
	Label string

	// Program is the absolute path to the service binary.
	Program string

	// Args are command-line arguments passed to the program.
	Args []string

	// WorkingDir is the working directory for the process. Empty means inherit.
	WorkingDir string

	// LogDir is the directory where stdout/stderr logs are written.
	LogDir string

	// Privileged indicates whether the process needs root/admin privileges.
	// On macOS, privileged services run as LaunchDaemons; unprivileged run as LaunchAgents.
	Privileged bool

	// PIDFile is the path where the process PID is stored.
	// Used by PID-file-based managers (e.g., PostgreSQL via pg_ctl).
	PIDFile string

	// KeepAlive indicates whether the process should be restarted automatically if it exits.
	KeepAlive bool

	// RunAtLoad indicates whether the process should start automatically when loaded.
	RunAtLoad bool

	// Env holds environment variables to set for the process.
	Env map[string]string
}

// ProcessManager is the interface for platform-specific process lifecycle management.
// Implementations exist for macOS launchd and direct PID-file-based management.
type ProcessManager interface {
	// Start launches the service process with the given configuration.
	Start(name string, cfg ProcessConfig) error

	// Stop terminates the service process identified by name.
	Stop(name string) error

	// Restart stops and then starts the service.
	Restart(name string, cfg ProcessConfig) error

	// IsRunning reports whether the service identified by name is currently running.
	IsRunning(name string) bool
}
