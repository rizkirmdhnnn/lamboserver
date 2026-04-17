package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// SiteConfig holds the configuration for a single virtual host site.
type SiteConfig struct {
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	PhpVersion string `json:"php_version,omitempty"`
	SSLEnabled bool   `json:"ssl_enabled"`
	CreatedAt  string `json:"created_at"`
}

// AppConfig is the top-level application configuration persisted to
// ~/.lamboserver/config.json. It tracks active runtime versions, site
// mappings, Nginx ports, and debug mode.
type AppConfig struct {
	ActivePhpVersion  string       `json:"active_php_version"`
	ActiveNodeVersion string       `json:"active_node_version"`
	Sites             []SiteConfig `json:"sites"`
	FirstRunComplete  bool         `json:"first_run_complete"`
	NginxPort         int          `json:"nginx_port"`
	NginxSSLPort      int          `json:"nginx_ssl_port"`
	DebugMode         bool         `json:"debug_mode"`
	MySQLEnabled      bool         `json:"mysql_enabled"`
	PostgreSQLEnabled bool         `json:"postgresql_enabled"`
}

// Store provides thread-safe, persistent storage for AppConfig. It caches the
// configuration in memory (protected by sync.RWMutex) and writes changes to the
// JSON file atomically. Get reads from the in-memory cache without file I/O;
// all mutating methods acquire a write lock before modifying and persisting.
type Store struct {
	mu       sync.RWMutex
	config   AppConfig
	filePath string
}

// NewStore creates a Store backed by the given file path, loading any existing
// configuration from disk. If the file does not exist, a default config is written.
func NewStore(filePath string) *Store {
	s := &Store{
		filePath: filePath,
		config: AppConfig{
			NginxPort:    80,
			NginxSSLPort: 443,
			Sites:        []SiteConfig{},
		},
	}
	s.Load()
	return s
}

// Load reads the configuration from disk into memory. If the file does not
// exist, it creates the file with the current (default) configuration.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.saveLocked()
		}
		return err
	}
	return json.Unmarshal(data, &s.config)
}

// Save persists the current in-memory configuration to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// Get returns the current in-memory AppConfig without file I/O. Safe for concurrent use.
func (s *Store) Get() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetActivePhpVersion updates the active PHP version in config and persists the change.
func (s *Store) SetActivePhpVersion(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.ActivePhpVersion = version
	return s.saveLocked()
}

// SetActiveNodeVersion updates the active Node.js version in config and persists the change.
func (s *Store) SetActiveNodeVersion(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.ActiveNodeVersion = version
	return s.saveLocked()
}

// AddSite appends a new site entry for the given domain and path. If the domain
// already exists in the config, the operation is a no-op.
func (s *Store) AddSite(domain, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, site := range s.config.Sites {
		if site.Domain == domain {
			return nil // already exists
		}
	}

	s.config.Sites = append(s.config.Sites, SiteConfig{
		Domain:     domain,
		Path:       path,
		SSLEnabled: true,
		CreatedAt:  time.Now().Format(time.RFC3339),
	})
	return s.saveLocked()
}

// RemoveSite removes the site entry for the given domain and persists the change.
func (s *Store) RemoveSite(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sites := make([]SiteConfig, 0, len(s.config.Sites))
	for _, site := range s.config.Sites {
		if site.Domain != domain {
			sites = append(sites, site)
		}
	}
	s.config.Sites = sites
	return s.saveLocked()
}

// GetSites returns a copy of the current site list. Safe for concurrent use.
func (s *Store) GetSites() []SiteConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SiteConfig, len(s.config.Sites))
	copy(result, s.config.Sites)
	return result
}

// SetFirstRunComplete marks the initial setup as done and persists the change.
func (s *Store) SetFirstRunComplete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.FirstRunComplete = true
	return s.saveLocked()
}

// SetDebugMode enables or disables debug logging in config and persists the change.
func (s *Store) SetDebugMode(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.DebugMode = enabled
	return s.saveLocked()
}

// SetMySQLEnabled enables or disables MySQL auto-start in config and persists the change.
func (s *Store) SetMySQLEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.MySQLEnabled = enabled
	return s.saveLocked()
}

// SetPostgreSQLEnabled enables or disables PostgreSQL auto-start in config and persists the change.
func (s *Store) SetPostgreSQLEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.PostgreSQLEnabled = enabled
	return s.saveLocked()
}
