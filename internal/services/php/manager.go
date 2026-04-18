// Package php provides PHP version management and FPM lifecycle control.
package php

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// Ensure system.RealFS and system.RealCmdRunner satisfy the package interfaces at compile time.
var _ FileSystem = system.RealFS{}
var _ CommandRunner = system.RealCmdRunner{}

var knownPhpSeries = []string{"7.4", "8.0", "8.1", "8.2", "8.3", "8.4", "8.5"}

// PhpVersion represents a detected PHP installation with its binary paths and active state.
type PhpVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Binary  string `json:"binary"`
	FpmBin  string `json:"fpm_bin"`
	Active  bool   `json:"active"`
	Source  string `json:"source"` // "system", "manual"
}

// AvailablePhpVersion represents a PHP version available for download from herdphp.com.
type AvailablePhpVersion struct {
	Version         string `json:"version"`
	Series          string `json:"series"`
	InternalVersion int    `json:"internal_version"`
	Installed       bool   `json:"installed"`
}

// Manager handles PHP version detection, installation, and activation.
// It discovers versions from system PATH and manually installed directories,
// downloads new versions from herdphp.com, and persists the active selection
// in the config store.
type Manager struct {
	paths *system.Paths
	store *config.Store
	fs    FileSystem
	cmd   CommandRunner
}

// NewManager creates a Manager with injected filesystem and command-runner dependencies.
// paths provides all LamboServer directory locations; store provides config persistence.
func NewManager(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{paths: paths, store: store, fs: fs, cmd: cmd}
}

// ListInstalled detects and returns all PHP versions found across system PATH and manually
// installed directories under ~/.lamboserver/php/. The Active field is set based on the
// currently configured active version in the config store.
func (m *Manager) ListInstalled() ([]PhpVersion, error) {
	activeVersion := m.store.Get().ActivePhpVersion
	return detectAll(m.paths, activeVersion, m.fs, m.cmd), nil
}

// SetActive switches the active PHP version by persisting the selection to the config store.
// Returns ErrVersionNotFound if the requested version is not currently installed.
func (m *Manager) SetActive(version string) error {
	versions, _ := m.ListInstalled()
	for _, v := range versions {
		if v.Version == version {
			return m.store.SetActivePhpVersion(version)
		}
	}
	return fmt.Errorf("SetActive %s: %w", version, ErrVersionNotFound)
}

// GetActive returns the currently active PHP version. If no version is explicitly set,
// it returns the first detected version. Returns ErrNoVersionsInstalled when no PHP
// installations are found.
func (m *Manager) GetActive() (*PhpVersion, error) {
	versions, _ := m.ListInstalled()
	for _, v := range versions {
		if v.Active {
			return &v, nil
		}
	}
	if len(versions) > 0 {
		return &versions[0], nil
	}
	return nil, fmt.Errorf("GetActive: %w", ErrNoVersionsInstalled)
}

// ListAvailable fetches downloadable PHP versions from herdphp.com for all known series
// (7.4 through 8.5). The Installed field on each entry is set to true if a matching
// major.minor version is already present on the system.
func (m *Manager) ListAvailable() ([]AvailablePhpVersion, error) {
	installed, _ := m.ListInstalled()
	installedSet := make(map[string]bool)
	for _, v := range installed {
		parts := strings.SplitN(v.Version, ".", 3)
		if len(parts) >= 2 {
			installedSet[parts[0]+"."+parts[1]] = true
		}
	}

	arch := getArch()

	var available []AvailablePhpVersion
	for _, series := range knownPhpSeries {
		ver, internalVer, err := fetchLatestForSeries(series, arch)
		if err != nil || ver == "" {
			continue
		}
		available = append(available, AvailablePhpVersion{
			Version:         ver,
			Series:          series,
			InternalVersion: internalVer,
			Installed:       installedSet[series],
		})
	}
	return available, nil
}

// phpUpdateResponse matches the Herd php-update API response.
type phpUpdateResponse struct {
	Version         string `json:"version"`
	InternalVersion int    `json:"internalVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// fetchLatestForSeries checks the latest PHP version for a given series.
// API: GET https://herdphp.com/api/php-update/{series}/{arch}/0/
// Using internalVersion=0 to always get the latest info.
func fetchLatestForSeries(series, arch string) (string, int, error) {
	url := fmt.Sprintf("https://herdphp.com/api/php-update/%s/%s/0/", series, arch)
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("HTTP %d for series %s", resp.StatusCode, series)
	}

	var result phpUpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}
	return result.Version, result.InternalVersion, nil
}

// Install downloads and installs the specified PHP version from herdphp.com.
// It downloads both the PHP CLI binary and the PHP-FPM binary into
// ~/.lamboserver/php/{version}/bin/ and ~/.lamboserver/php/{version}/sbin/ respectively.
// Returns ErrVersionNotAvailable if the version is not in the downloadable list.
func (m *Manager) Install(version string) error {
	available, err := m.ListAvailable()
	if err != nil {
		return fmt.Errorf("failed to fetch available versions: %w", err)
	}

	var target *AvailablePhpVersion
	for _, v := range available {
		if v.Version == version {
			target = &v
			break
		}
	}
	if target == nil {
		return fmt.Errorf("Install %s: %w", version, ErrVersionNotAvailable)
	}

	destDir := m.paths.PhpVersionDir(version)
	if err := m.fs.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	arch := getArch()
	seriesTag := strings.ReplaceAll(target.Series, ".", "") // "8.4" -> "84"

	// Download PHP CLI
	cliURL := fmt.Sprintf("https://download.herdphp.com/%s/php%s-%s", target.Series, seriesTag, arch)
	cliBin := destDir + "/bin/php"
	if err := m.fs.MkdirAll(destDir+"/bin", 0755); err != nil {
		return err
	}
	if err := system.DownloadFile(cliURL, cliBin); err != nil {
		return fmt.Errorf("failed to download PHP CLI: %w", err)
	}
	m.fs.Chmod(cliBin, 0755) //nolint: errcheck — chmod failure is non-fatal after download

	// Download PHP FPM
	fpmURL := fmt.Sprintf("https://download.herdphp.com/%s/php%s-fpm-%s", target.Series, seriesTag, arch)
	fpmBin := destDir + "/sbin/php-fpm"
	if err := m.fs.MkdirAll(destDir+"/sbin", 0755); err != nil {
		return err
	}
	if err := system.DownloadFile(fpmURL, fpmBin); err != nil {
		return fmt.Errorf("failed to download PHP FPM: %w", err)
	}
	m.fs.Chmod(fpmBin, 0755) //nolint: errcheck — chmod failure is non-fatal after download

	return nil
}

// Uninstall removes a manually installed PHP version by deleting its directory under
// ~/.lamboserver/php/{version}/. Returns ErrVersionActive if the version is currently
// active, and ErrVersionNotFound if it is not installed.
func (m *Manager) Uninstall(version string) error {
	if m.store.Get().ActivePhpVersion == version {
		return fmt.Errorf("Uninstall %s: %w", version, ErrVersionActive)
	}

	dir := m.paths.PhpVersionDir(version)
	if _, err := m.fs.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("Uninstall %s: %w", version, ErrVersionNotFound)
	}
	return m.fs.RemoveAll(dir)
}

func getArch() string {
	arch := system.GetArchitecture()
	if arch == "amd64" {
		return "x86_64"
	}
	return "arm64"
}
