package services

import (
	"fmt"
	"maps"
)

// Manager is the central orchestrator that holds references to all registered
// services. It provides lookup by name so that app.go can dispatch generic
// calls like StartService("nginx") without importing individual packages.
type Manager struct {
	services  map[string]Service
	versioned map[string]VersionedService
	webAdmin  map[string]WebAdminService
}

// NewManager creates an empty Manager ready for service registration.
func NewManager() *Manager {
	return &Manager{
		services:  make(map[string]Service),
		versioned: make(map[string]VersionedService),
		webAdmin:  make(map[string]WebAdminService),
	}
}

// Register adds a Service to the registry under the given name.
// If the service also implements VersionedService, it is registered
// in both the services and versioned maps.
func (m *Manager) Register(name string, svc Service) {
	m.services[name] = svc
	if vs, ok := svc.(VersionedService); ok {
		m.versioned[name] = vs
	}
}

// RegisterWebAdmin adds a WebAdminService to the registry under the given name.
func (m *Manager) RegisterWebAdmin(name string, svc WebAdminService) {
	m.webAdmin[name] = svc
}

// Get returns the Service registered under the given name.
// Returns an error if the name is not found.
func (m *Manager) Get(name string) (Service, error) {
	svc, ok := m.services[name]
	if !ok {
		return nil, fmt.Errorf("service %q not found", name)
	}
	return svc, nil
}

// GetVersioned returns the VersionedService registered under the given name.
// Returns an error if the name is not found or is not a versioned service.
func (m *Manager) GetVersioned(name string) (VersionedService, error) {
	svc, ok := m.versioned[name]
	if !ok {
		return nil, fmt.Errorf("versioned service %q not found", name)
	}
	return svc, nil
}

// GetWebAdmin returns the WebAdminService registered under the given name.
// Returns an error if the name is not found.
func (m *Manager) GetWebAdmin(name string) (WebAdminService, error) {
	svc, ok := m.webAdmin[name]
	if !ok {
		return nil, fmt.Errorf("web admin service %q not found", name)
	}
	return svc, nil
}

// All returns a copy of the service registry map.
func (m *Manager) All() map[string]Service {
	result := make(map[string]Service, len(m.services))
	maps.Copy(result, m.services)
	return result
}

// AllVersioned returns a copy of the versioned service registry map.
func (m *Manager) AllVersioned() map[string]VersionedService {
	result := make(map[string]VersionedService, len(m.versioned))
	maps.Copy(result, m.versioned)
	return result
}

// AllWebAdmin returns a copy of the web admin service registry map.
func (m *Manager) AllWebAdmin() map[string]WebAdminService {
	result := make(map[string]WebAdminService, len(m.webAdmin))
	maps.Copy(result, m.webAdmin)
	return result
}

// Names returns the names of all registered services (both daemon and web admin).
func (m *Manager) Names() []string {
	seen := make(map[string]bool)
	var names []string
	for name := range m.services {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	for name := range m.webAdmin {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}
