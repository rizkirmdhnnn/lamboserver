// Package services defines the unified interfaces for all managed services.
// Every service in the application must implement one of the three interfaces
// defined here: Service, VersionedService, or WebAdminService.
//
// Interface assignment:
//
//	Nginx        → Service
//	DNSMasq      → Service
//	PHP          → VersionedService
//	Node.js      → VersionedService
//	MySQL        → Service
//	PostgreSQL   → Service
//	phpMyAdmin   → WebAdminService
package services

// ServiceStatus represents the current state of a managed service.
type ServiceStatus string

const (
	StatusNotInstalled ServiceStatus = "not_installed"
	StatusStopped      ServiceStatus = "stopped"
	StatusStarting     ServiceStatus = "starting"
	StatusRunning      ServiceStatus = "running"
	StatusError        ServiceStatus = "error"
)

// Service is the base interface for all daemon-backed services.
// Services that run as background processes (Nginx, DNSMasq, MySQL,
// PostgreSQL) implement this interface.
type Service interface {
	// Install downloads and sets up the service binary for the given version.
	// For unversioned services, version may be empty.
	Install(version string) error

	// Start launches the service daemon.
	Start() error

	// Stop terminates the service daemon.
	Stop() error

	// Restart performs a stop followed by start, or a graceful reload
	// if the service supports it.
	Restart() error

	// Status returns the current running state of the service.
	Status() ServiceStatus

	// Logs returns the most recent log lines from the service log file.
	Logs() ([]string, error)

	// Version returns the currently installed or active version string.
	Version() string
}

// VersionedService extends Service for runtimes that support multiple
// simultaneously installed versions (PHP, Node.js). Only one version
// is "active" at a time, but multiple can be installed.
type VersionedService interface {
	Service

	// SwitchVersion changes the active version to the specified one.
	// The version must already be installed.
	SwitchVersion(version string) error

	// InstalledVersions returns a list of all locally installed version strings.
	InstalledVersions() []string

	// ActiveVersion returns the currently active version string.
	ActiveVersion() string
}

// WebAdminService is the interface for web-based administration tools
// that are served through an existing web server (e.g., phpMyAdmin via
// Nginx + PHP-FPM). These services have no daemon of their own.
type WebAdminService interface {
	// Install downloads and extracts the web admin tool to the binaries directory.
	Install(version string) error

	// URL returns the full URL where the admin tool is accessible
	// (e.g., "https://phpmyadmin.test:8088").
	URL() string

	// IsInstalled reports whether the admin tool has been installed.
	IsInstalled() bool

	// Version returns the installed version string, or empty if not installed.
	Version() string
}
