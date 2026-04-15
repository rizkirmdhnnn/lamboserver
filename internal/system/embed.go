package system

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed embedded/darwin-arm64/*
var embeddedBinaries embed.FS

// ExtractBinary extracts an embedded binary to the target path if it doesn't exist.
func ExtractBinary(name, targetPath string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return nil // already extracted
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	embeddedPath := filepath.Join("embedded", "darwin-"+arch, name)
	data, err := embeddedBinaries.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("embedded binary %s not found for %s: %w", name, arch, err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(targetPath, data, 0755); err != nil {
		return fmt.Errorf("failed to extract %s: %w", name, err)
	}

	return nil
}
