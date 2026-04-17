package php

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// versionAdder is a callback used by detection functions to register a found PHP version.
// It deduplicates by version string so each version is reported only once.
type versionAdder func(ver, path, binary, fpmBin, source string)

// detectAll runs all detection sources and returns found PHP versions.
// It queries the system PATH (detectSystem) and the manually installed directory
// (detectManual). The first discovered version is marked active when no explicit
// active version is configured in the store.
func detectAll(paths *system.Paths, activeVersion string, fs FileSystem, cmd CommandRunner) []PhpVersion {
	var versions []PhpVersion
	seen := make(map[string]bool)

	add := func(ver, path, binary, fpmBin, source string) {
		if seen[ver] {
			return
		}
		seen[ver] = true
		versions = append(versions, PhpVersion{
			Version: ver,
			Path:    path,
			Binary:  binary,
			FpmBin:  fpmBin,
			Active:  ver == activeVersion,
			Source:  source,
		})
	}

	detectSystem(add, paths, cmd)
	detectManual(paths, add, fs, cmd)

	if len(versions) > 0 && activeVersion == "" {
		versions[0].Active = true
	}

	return versions
}

// detectSystem detects a PHP installation on the system PATH. It skips any binary
// that resolves back to LamboServer's own managed directory to avoid self-detection.
func detectSystem(add versionAdder, paths *system.Paths, cmd CommandRunner) {
	phpPath, err := cmd.LookPath("php")
	if err != nil {
		return
	}

	realPath, _ := filepath.EvalSymlinks(phpPath)
	if realPath == "" {
		realPath = phpPath
	}

	// Skip if it points to our own lamboserver bin
	home, _ := os.UserHomeDir()
	if strings.Contains(realPath, filepath.Join(home, ".lamboserver")) {
		return
	}

	ver := getFullVersion(phpPath, cmd)
	if ver == "" {
		return
	}

	fpmPath, _ := cmd.LookPath("php-fpm")
	add(ver, filepath.Dir(phpPath), phpPath, fpmPath, "system")
}

// detectManual scans ~/.lamboserver/php/ for manually installed PHP versions.
// Each subdirectory (excluding "current") is checked for a bin/php binary; if found,
// the version string is obtained by running `php -r "echo PHP_VERSION;"`.
func detectManual(paths *system.Paths, add versionAdder, fs FileSystem, cmd CommandRunner) {
	entries, err := fs.ReadDir(paths.PhpDir())
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "current" {
			continue
		}
		dir := paths.PhpVersionDir(entry.Name())
		phpBin := filepath.Join(dir, "bin", "php")
		if _, err := fs.Stat(phpBin); err != nil {
			continue
		}

		ver := getFullVersion(phpBin, cmd)
		if ver == "" {
			ver = entry.Name()
		}

		fpmBin := filepath.Join(dir, "sbin", "php-fpm")
		if _, err := fs.Stat(fpmBin); err != nil {
			fpmBin = ""
		}
		add(ver, dir, phpBin, fpmBin, "manual")
	}
}

// getFullVersion runs the given PHP binary with `-r "echo PHP_VERSION;"` and returns
// the trimmed output. Returns an empty string if the binary fails to execute.
func getFullVersion(phpBin string, cmd CommandRunner) string {
	output, err := cmd.Run(phpBin, "-r", "echo PHP_VERSION;")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}
