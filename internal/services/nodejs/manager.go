package nodejs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// versionRe validates Node.js version strings, accepting only the form vX.Y.Z or X.Y.Z.
var versionRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

// NodeVersion represents an installed Node.js version with its directory path and active state.
type NodeVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Active  bool   `json:"active"`
}

// osFileSystem is the production implementation of FileSystem using os package.
type osFileSystem struct{}

func (osFileSystem) ReadDir(name string) ([]fs.DirEntry, error)      { return os.ReadDir(name) }
func (osFileSystem) MkdirAll(path string, perm fs.FileMode) error     { return os.MkdirAll(path, perm) }
func (osFileSystem) RemoveAll(path string) error                       { return os.RemoveAll(path) }

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
	return system.RunCommand(name, args...)
}

// Manager handles Node.js version installation, detection, and activation.
// Installed versions are stored as subdirectories under ~/.lamboserver/node/.
// The active version is persisted in the config store.
type Manager struct {
	paths *system.Paths
	store *config.Store
	fs    FileSystem
	cmd   CommandRunner
}

// NewManager creates a Manager with injected dependencies.
func NewManager(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{
		paths: paths,
		store: store,
		fs:    fs,
		cmd:   cmd,
	}
}

// newManagerWithDeps creates a Manager with injected dependencies (used in tests).
func newManagerWithDeps(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{
		paths: paths,
		store: store,
		fs:    fs,
		cmd:   cmd,
	}
}

// ListInstalled returns all Node.js versions installed under ~/.lamboserver/node/.
// The Active field is set for the version that matches the store's ActiveNodeVersion.
// Returns an empty slice (not an error) if the node directory does not exist yet.
func (m *Manager) ListInstalled() ([]NodeVersion, error) {
	var versions []NodeVersion
	activeVersion := m.store.Get().ActiveNodeVersion

	entries, err := m.fs.ReadDir(m.paths.NodeDir())
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

// ListAvailable fetches recent LTS Node.js versions from https://nodejs.org/dist/index.json.
// It returns up to 5 unique major LTS versions (e.g. v22.x, v20.x, v18.x).
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

// Install downloads and extracts the specified Node.js version from nodejs.org into
// ~/.lamboserver/node/{version}/. A "v" prefix is added to the version string if absent.
// The tarball is streamed and extracted via curl and tar in a single shell command.
func (m *Manager) Install(version string) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !versionRe.MatchString(version) {
		return fmt.Errorf("invalid Node.js version: %q", version)
	}

	arch := system.GetArchitecture()
	archStr := "arm64"
	if arch == "amd64" {
		archStr = "x64"
	}

	url := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-darwin-%s.tar.gz", version, version, archStr)
	destDir := m.paths.NodeVersionDir(version)

	if err := m.fs.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Download and extract
	_, err := m.cmd.Run("sh", "-c",
		fmt.Sprintf("curl -sL %s | tar xz --strip-components=1 -C '%s'", url, filepath.Clean(destDir)))
	return err
}

// SetActive persists the specified Node.js version as active in the config store.
func (m *Manager) SetActive(version string) error {
	return m.store.SetActiveNodeVersion(version)
}

// Uninstall removes the specified Node.js version directory from ~/.lamboserver/node/.
func (m *Manager) Uninstall(version string) error {
	return m.fs.RemoveAll(m.paths.NodeVersionDir(version))
}
