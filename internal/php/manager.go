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

var knownPhpSeries = []string{"7.4", "8.0", "8.1", "8.2", "8.3", "8.4", "8.5"}

type PhpVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Binary  string `json:"binary"`
	FpmBin  string `json:"fpm_bin"`
	Active  bool   `json:"active"`
	Source  string `json:"source"` // "system", "manual"
}

type AvailablePhpVersion struct {
	Version         string `json:"version"`
	Series          string `json:"series"`
	InternalVersion int    `json:"internal_version"`
	Installed       bool   `json:"installed"`
}

type Manager struct {
	paths *system.Paths
	store *config.Store
}

func NewManager(paths *system.Paths, store *config.Store) *Manager {
	return &Manager{paths: paths, store: store}
}

// ListInstalled detects PHP versions from all known sources.
func (m *Manager) ListInstalled() ([]PhpVersion, error) {
	activeVersion := m.store.Get().ActivePhpVersion
	return detectAll(m.paths, activeVersion), nil
}

// SetActive switches the active PHP version.
func (m *Manager) SetActive(version string) error {
	versions, _ := m.ListInstalled()
	for _, v := range versions {
		if v.Version == version {
			return m.store.SetActivePhpVersion(version)
		}
	}
	return fmt.Errorf("PHP version %s not found", version)
}

// GetActive returns the currently active PHP version.
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
	return nil, fmt.Errorf("no PHP versions installed")
}

// ListAvailable fetches downloadable PHP versions.
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

// Install downloads a PHP version.
// Download URLs: https://download.herdphp.com/{series}/php{XY}-{arch}
//                https://download.herdphp.com/{series}/php{XY}-fpm-{arch}
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
		return fmt.Errorf("PHP version %s not available for download", version)
	}

	destDir := m.paths.PhpVersionDir(version)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	arch := getArch()
	seriesTag := strings.ReplaceAll(target.Series, ".", "") // "8.4" -> "84"

	// Download PHP CLI
	cliURL := fmt.Sprintf("https://download.herdphp.com/%s/php%s-%s", target.Series, seriesTag, arch)
	cliBin := destDir + "/bin/php"
	if err := os.MkdirAll(destDir+"/bin", 0755); err != nil {
		return err
	}
	if err := system.DownloadFile(cliURL, cliBin); err != nil {
		return fmt.Errorf("failed to download PHP CLI: %w", err)
	}
	os.Chmod(cliBin, 0755)

	// Download PHP FPM
	fpmURL := fmt.Sprintf("https://download.herdphp.com/%s/php%s-fpm-%s", target.Series, seriesTag, arch)
	fpmBin := destDir + "/sbin/php-fpm"
	if err := os.MkdirAll(destDir+"/sbin", 0755); err != nil {
		return err
	}
	if err := system.DownloadFile(fpmURL, fpmBin); err != nil {
		return fmt.Errorf("failed to download PHP FPM: %w", err)
	}
	os.Chmod(fpmBin, 0755)

	return nil
}

// Uninstall removes a manually installed PHP version.
func (m *Manager) Uninstall(version string) error {
	if m.store.Get().ActivePhpVersion == version {
		return fmt.Errorf("cannot uninstall active PHP version %s, switch to another version first", version)
	}

	dir := m.paths.PhpVersionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("PHP version %s is not manually installed", version)
	}
	return os.RemoveAll(dir)
}

func getArch() string {
	arch := system.GetArchitecture()
	if arch == "amd64" {
		return "x86_64"
	}
	return "arm64"
}
