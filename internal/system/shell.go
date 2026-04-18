package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

)

const shellMarker = "# Added by LamboServer"
const shellBlock = `
# Added by LamboServer
export PATH="$HOME/.lamboserver/bin:$PATH"
`

// Integration manages PATH injection into the user's shell rc files and maintains
// symlinks in ~/.lamboserver/bin/ for active PHP and Node.js binaries.
type Integration struct {
	paths   *Paths
	homeDir string
}

// NewIntegration creates a new Integration with the given path configuration.
func NewIntegration(paths *Paths) *Integration {
	return &Integration{paths: paths}
}

// WithHome returns a copy of the Integration with a custom home directory.
// Used in tests to redirect shell rc file operations to a temp directory.
func (s *Integration) WithHome(home string) *Integration {
	return &Integration{paths: s.paths, homeDir: home}
}

// IsInstalled checks if PATH injection exists in any shell rc file
func (s *Integration) IsInstalled() bool {
	for _, rc := range s.shellRCFiles() {
		if s.fileContains(rc, shellMarker) {
			return true
		}
	}
	return false
}

// Install adds PATH injection to the user's shell rc files
func (s *Integration) Install() error {
	rcFiles := s.shellRCFiles()
	installed := false

	for _, rc := range rcFiles {
		if s.fileContains(rc, shellMarker) {
			installed = true // already installed in this file
			continue
		}

		// Only install in files that exist, plus always ensure .zshrc (macOS default)
		_, err := os.Stat(rc)
		isZshrc := strings.HasSuffix(rc, ".zshrc")

		if err != nil && !isZshrc {
			continue
		}

		if err := s.appendToFile(rc, shellBlock); err != nil {
			continue
		}
		installed = true
	}

	if !installed {
		// Fallback: create .zshrc with the block
		home := s.homeDir
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		zshrc := filepath.Join(home, ".zshrc")
		return s.appendToFile(zshrc, shellBlock)
	}

	return nil
}

// Uninstall removes PATH injection from all shell rc files
func (s *Integration) Uninstall() error {
	for _, rc := range s.shellRCFiles() {
		if s.fileContains(rc, shellMarker) {
			s.removeBlock(rc)
		}
	}
	return nil
}

// CreateSymlink creates or updates a symlink in ~/.lamboserver/bin/
func (s *Integration) CreateSymlink(name, target string) error {
	linkPath := filepath.Join(s.paths.BinDir(), name)

	// Remove existing symlink or file
	os.Remove(linkPath)

	return os.Symlink(target, linkPath)
}

// RemoveSymlink removes a symlink from ~/.lamboserver/bin/
func (s *Integration) RemoveSymlink(name string) error {
	linkPath := filepath.Join(s.paths.BinDir(), name)
	return os.Remove(linkPath)
}

// LinkPhpVersion creates symlinks for a specific PHP version.
// phpBinary is the full path to the php executable.
// fpmBinary is the full path to php-fpm, can be empty.
// basePath is the directory containing the binaries.
func (s *Integration) LinkPhpVersion(phpBinary, fpmBinary, basePath string) error {
	// Always link the main php binary
	if phpBinary != "" {
		s.CreateSymlink("php", phpBinary)
	}

	// Link php-fpm
	if fpmBinary != "" {
		s.CreateSymlink("php-fpm", fpmBinary)
	}

	// Link additional binaries from standard locations
	for _, sub := range []string{"bin", "sbin", ""} {
		dir := basePath
		if sub != "" {
			dir = filepath.Join(basePath, sub)
		}
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		extras := []string{"phpize", "php-config", "php-cgi", "pear", "pecl", "composer"}
		for _, name := range extras {
			srcPath := filepath.Join(dir, name)
			if _, err := os.Stat(srcPath); err == nil {
				s.CreateSymlink(name, srcPath)
			}
		}
	}

	return nil
}

// LinkNodeBinaries creates symlinks for node, npm, npx from a given Node path
func (s *Integration) LinkNodeBinaries(nodeBasePath string) error {
	binDir := filepath.Join(nodeBasePath, "bin")

	nodeBins := []string{"node", "npm", "npx", "corepack"}

	for _, name := range nodeBins {
		srcPath := filepath.Join(binDir, name)
		if _, err := os.Stat(srcPath); err == nil {
			s.CreateSymlink(name, srcPath)
		}
	}

	return nil
}

func (s *Integration) shellRCFiles() []string {
	home := s.homeDir
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
	}
}

func (s *Integration) fileContains(path, needle string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			return true
		}
	}
	return false
}

func (s *Integration) appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func (s *Integration) removeBlock(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	skip := false

	for _, line := range lines {
		if strings.Contains(line, shellMarker) {
			skip = true
			continue
		}
		if skip {
			// Skip the export PATH line after the marker
			if strings.Contains(line, ".lamboserver/bin") {
				skip = false
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			skip = false
		}
		result = append(result, line)
	}

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}
