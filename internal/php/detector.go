package php

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

type versionAdder func(ver, path, binary, fpmBin, source string)

// detectAll runs all detection sources and returns found PHP versions.
func detectAll(paths *system.Paths, activeVersion string) []PhpVersion {
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

	detectSystem(add, paths)
	detectManual(paths, add)

	if len(versions) > 0 && activeVersion == "" {
		versions[0].Active = true
	}

	return versions
}

func detectSystem(add versionAdder, paths *system.Paths) {
	phpPath, err := exec.LookPath("php")
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

	ver := getFullVersion(phpPath)
	if ver == "" {
		return
	}

	fpmPath, _ := exec.LookPath("php-fpm")
	add(ver, filepath.Dir(phpPath), phpPath, fpmPath, "system")
}

func detectManual(paths *system.Paths, add versionAdder) {
	entries, err := os.ReadDir(paths.PhpDir())
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "current" {
			continue
		}
		dir := paths.PhpVersionDir(entry.Name())
		phpBin := filepath.Join(dir, "bin", "php")
		if _, err := os.Stat(phpBin); err != nil {
			continue
		}

		ver := getFullVersion(phpBin)
		if ver == "" {
			ver = entry.Name()
		}

		fpmBin := filepath.Join(dir, "sbin", "php-fpm")
		if _, err := os.Stat(fpmBin); err != nil {
			fpmBin = ""
		}
		add(ver, dir, phpBin, fpmBin, "manual")
	}
}

func getFullVersion(phpBin string) string {
	cmd := exec.Command(phpBin, "-r", "echo PHP_VERSION;")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
