package nodejs

import (
	"github.com/rizkirmdhnnn/lamboserver/internal/services"
)

// ServiceAdapter wraps the existing Node.js Manager to implement the unified
// services.VersionedService interface.
type ServiceAdapter struct {
	*Manager
}

// NewServiceAdapter creates a ServiceAdapter wrapping the given Manager.
func NewServiceAdapter(m *Manager) *ServiceAdapter {
	return &ServiceAdapter{Manager: m}
}

// Compile-time interface check.
var _ services.VersionedService = (*ServiceAdapter)(nil)

// Start is a no-op for Node.js — it has no daemon process.
// ARCH: Node.js is a version manager only; there is no background service to start.
func (a *ServiceAdapter) Start() error {
	return nil
}

// Stop is a no-op for Node.js — it has no daemon process.
// ARCH: Node.js is a version manager only; there is no background service to stop.
func (a *ServiceAdapter) Stop() error {
	return nil
}

// Restart is a no-op for Node.js.
func (a *ServiceAdapter) Restart() error {
	return nil
}

// Install downloads and installs the given Node.js version.
func (a *ServiceAdapter) Install(version string) error {
	return a.Manager.Install(version)
}

// Status returns StatusRunning if an active version is set, StatusStopped otherwise.
// ARCH: Node.js has no daemon; "running" means an active version is configured.
func (a *ServiceAdapter) Status() services.ServiceStatus {
	if a.store.Get().ActiveNodeVersion != "" {
		return services.StatusRunning
	}
	return services.StatusStopped
}

// Logs returns nil — Node.js has no daemon log.
func (a *ServiceAdapter) Logs() ([]string, error) {
	return nil, nil
}

// Version returns the currently active Node.js version string.
func (a *ServiceAdapter) Version() string {
	return a.store.Get().ActiveNodeVersion
}

// SwitchVersion changes the active Node.js version.
func (a *ServiceAdapter) SwitchVersion(version string) error {
	return a.Manager.SetActive(version)
}

// InstalledVersions returns the version strings of all installed Node.js versions.
func (a *ServiceAdapter) InstalledVersions() []string {
	versions, err := a.Manager.ListInstalled()
	if err != nil {
		return nil
	}
	var result []string
	for _, v := range versions {
		result = append(result, v.Version)
	}
	return result
}

// ActiveVersion returns the currently active Node.js version.
func (a *ServiceAdapter) ActiveVersion() string {
	return a.store.Get().ActiveNodeVersion
}
