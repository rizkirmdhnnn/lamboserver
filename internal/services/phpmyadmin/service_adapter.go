package phpmyadmin

import (
	"github.com/rizkirmdhnnn/lamboserver/internal/services"
)

// ServiceAdapter wraps the existing phpMyAdmin Manager to implement the unified
// services.WebAdminService interface.
type ServiceAdapter struct {
	*Manager
}

// NewServiceAdapter creates a ServiceAdapter wrapping the given Manager.
func NewServiceAdapter(m *Manager) *ServiceAdapter {
	return &ServiceAdapter{Manager: m}
}

// Compile-time interface check.
var _ services.WebAdminService = (*ServiceAdapter)(nil)

// Install downloads and sets up phpMyAdmin. The version parameter is ignored
// since phpMyAdmin uses a pinned version constant.
func (a *ServiceAdapter) Install(v string) error {
	return a.Manager.Install()
}

// URL returns the local URL where phpMyAdmin is accessible.
func (a *ServiceAdapter) URL() string {
	return "https://phpmyadmin.test"
}

// IsInstalled delegates to the underlying Manager.
// (Already provided by embedded *Manager, but listed for clarity.)

// Version returns the pinned phpMyAdmin version.
func (a *ServiceAdapter) Version() string {
	if !a.Manager.IsInstalled() {
		return ""
	}
	return version
}
