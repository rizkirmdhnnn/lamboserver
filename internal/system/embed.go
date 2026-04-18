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

// codesignBinary ad-hoc codesigns a binary at path.
// Returns an error if codesign fails; the caller decides whether to block or warn.
func codesignBinary(path string) error {
	_, err := RunCommand("/usr/bin/codesign", "--force", "--sign", "-", path)
	return err
}

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

	// D-07: Ad-hoc codesign immediately after extraction (SIGN-03).
	// D-08: Non-fatal — log warning but do not block service startup.
	if err := codesignBinary(targetPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: codesign failed for %s: %v\n", name, err)
	}
	return nil
}
