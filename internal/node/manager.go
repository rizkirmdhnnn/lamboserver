package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

type NodeVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Active  bool   `json:"active"`
}

type Manager struct {
	paths *system.Paths
	store *config.Store
}

func NewManager(paths *system.Paths, store *config.Store) *Manager {
	return &Manager{paths: paths, store: store}
}

func (m *Manager) ListInstalled() ([]NodeVersion, error) {
	var versions []NodeVersion
	activeVersion := m.store.Get().ActiveNodeVersion

	entries, err := os.ReadDir(m.paths.NodeDir())
	if err != nil {
		if os.IsNotExist(err) {
			return versions, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			versions = append(versions, NodeVersion{
				Version: entry.Name(),
				Path:    m.paths.NodeVersionDir(entry.Name()),
				Active:  entry.Name() == activeVersion,
			})
		}
	}
	return versions, nil
}

type nodeDistVersion struct {
	Version string `json:"version"`
	LTS     any    `json:"lts"`
}

func (m *Manager) ListAvailable() ([]string, error) {
	resp, err := http.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node versions: %w", err)
	}
	defer resp.Body.Close()

	var distVersions []nodeDistVersion
	if err := json.NewDecoder(resp.Body).Decode(&distVersions); err != nil {
		return nil, err
	}

	// Return LTS versions only, limit to recent ones
	var versions []string
	seen := make(map[string]bool)
	for _, v := range distVersions {
		if v.LTS != nil && v.LTS != false {
			major := strings.Split(strings.TrimPrefix(v.Version, "v"), ".")[0]
			if !seen[major] {
				seen[major] = true
				versions = append(versions, v.Version)
				if len(versions) >= 5 {
					break
				}
			}
		}
	}
	return versions, nil
}

func (m *Manager) Install(version string) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	arch := system.GetArchitecture()
	archStr := "arm64"
	if arch == "amd64" {
		archStr = "x64"
	}

	url := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-darwin-%s.tar.gz", version, version, archStr)
	destDir := m.paths.NodeVersionDir(version)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Download and extract
	_, err := system.RunCommand("sh", "-c",
		fmt.Sprintf("curl -sL %s | tar xz --strip-components=1 -C %s", url, destDir))
	return err
}

func (m *Manager) SetActive(version string) error {
	return m.store.SetActiveNodeVersion(version)
}

func (m *Manager) Uninstall(version string) error {
	return os.RemoveAll(m.paths.NodeVersionDir(version))
}
