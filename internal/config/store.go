package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type SiteConfig struct {
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	PhpVersion string `json:"php_version,omitempty"`
	SSLEnabled bool   `json:"ssl_enabled"`
	CreatedAt  string `json:"created_at"`
}

type AppConfig struct {
	ActivePhpVersion  string       `json:"active_php_version"`
	ActiveNodeVersion string       `json:"active_node_version"`
	Sites             []SiteConfig `json:"sites"`
	FirstRunComplete  bool         `json:"first_run_complete"`
	NginxPort         int          `json:"nginx_port"`
	NginxSSLPort      int          `json:"nginx_ssl_port"`
	DebugMode         bool         `json:"debug_mode"`
}

type Store struct {
	mu       sync.RWMutex
	config   AppConfig
	filePath string
}

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

func (s *Store) Get() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Store) SetActivePhpVersion(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.ActivePhpVersion = version
	return s.saveLocked()
}

func (s *Store) SetActiveNodeVersion(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.ActiveNodeVersion = version
	return s.saveLocked()
}

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

func (s *Store) GetSites() []SiteConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SiteConfig, len(s.config.Sites))
	copy(result, s.config.Sites)
	return result
}

func (s *Store) SetFirstRunComplete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.FirstRunComplete = true
	return s.saveLocked()
}

func (s *Store) SetDebugMode(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.DebugMode = enabled
	return s.saveLocked()
}
