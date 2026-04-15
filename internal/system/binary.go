package system

import (
	"os"
)

// BinaryLocator finds and installs a binary from embedded resources, local path, or system PATH.
type BinaryLocator struct {
	Name      string // binary name (e.g. "nginx", "dnsmasq")
	LocalPath string // expected path in ~/.lamboserver/bin/
}

// Find returns the path to the binary, checking local then system PATH.
func (b *BinaryLocator) Find() string {
	if _, err := os.Stat(b.LocalPath); err == nil {
		return b.LocalPath
	}

	if CommandExists(b.Name) {
		if path, err := RunCommand("which", b.Name); err == nil && path != "" {
			return path
		}
	}

	return ""
}

// IsInstalled returns true if the binary can be found.
func (b *BinaryLocator) IsInstalled() bool {
	return b.Find() != ""
}

// Install extracts the embedded binary to LocalPath.
func (b *BinaryLocator) Install() error {
	if b.IsInstalled() {
		return nil
	}
	return ExtractBinary(b.Name, b.LocalPath)
}
