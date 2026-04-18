// Package binaries provides a centralized registry of downloadable service
// binaries and a downloader that handles HTTP fetching with checksum verification.
package binaries

import (
	"fmt"
	"runtime"
)

// BinaryInfo holds download metadata for a specific service version.
type BinaryInfo struct {
	// URL is the download URL for the binary or archive.
	URL string

	// Checksum is the expected SHA256 checksum of the downloaded file (hex-encoded).
	// Empty string skips verification.
	Checksum string

	// StripComponents is the number of leading path components to strip when extracting
	// a tar archive (like tar --strip-components). 0 for non-archive downloads.
	StripComponents int

	// IsArchive indicates whether the download is a .tar.gz that needs extraction.
	IsArchive bool
}

// Registry maps service names and versions to their download information.
// It serves as the single source of truth for where to find binaries.
type Registry struct {
	entries map[string]func(version, os, arch string) (*BinaryInfo, error)
}

// NewRegistry creates a Registry pre-populated with known service download sources.
func NewRegistry() *Registry {
	r := &Registry{
		entries: make(map[string]func(version, os, arch string) (*BinaryInfo, error)),
	}
	r.registerDefaults()
	return r
}

// Lookup returns the BinaryInfo for the given service, version, and current platform.
func (r *Registry) Lookup(service, version string) (*BinaryInfo, error) {
	fn, ok := r.entries[service]
	if !ok {
		return nil, fmt.Errorf("no download source registered for service %q", service)
	}
	return fn(version, runtime.GOOS, runtime.GOARCH)
}

// Register adds a custom download source for a service.
func (r *Registry) Register(service string, fn func(version, os, arch string) (*BinaryInfo, error)) {
	r.entries[service] = fn
}

func (r *Registry) registerDefaults() {
	r.entries["php"] = phpDownloadURL
	r.entries["nodejs"] = nodejsDownloadURL
	r.entries["mysql"] = mysqlDownloadURL
	r.entries["postgresql"] = postgresDownloadURL
}

func phpDownloadURL(version, goos, arch string) (*BinaryInfo, error) {
	if goos != "darwin" {
		return nil, fmt.Errorf("PHP download not supported on %s", goos)
	}
	// herdphp.com format: php{XY}-{arch} where XY is major+minor without dot
	return &BinaryInfo{
		URL:       fmt.Sprintf("https://download.herdphp.com/%s/php%s-%s", version, version, arch),
		IsArchive: false,
	}, nil
}

func nodejsDownloadURL(version, goos, arch string) (*BinaryInfo, error) {
	if goos != "darwin" && goos != "linux" {
		return nil, fmt.Errorf("Node.js download not yet supported on %s", goos)
	}
	nodeArch := arch
	if nodeArch == "amd64" {
		nodeArch = "x64"
	}
	return &BinaryInfo{
		URL:             fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-%s-%s.tar.gz", version, version, goos, nodeArch),
		IsArchive:       true,
		StripComponents: 1,
	}, nil
}

func mysqlDownloadURL(version, goos, arch string) (*BinaryInfo, error) {
	if goos != "darwin" {
		return nil, fmt.Errorf("MySQL download not yet supported on %s", goos)
	}
	return &BinaryInfo{
		URL:             fmt.Sprintf("https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-%s-macos14-%s.tar.gz", version, arch),
		IsArchive:       true,
		StripComponents: 1,
	}, nil
}

func postgresDownloadURL(version, goos, arch string) (*BinaryInfo, error) {
	if goos != "darwin" {
		return nil, fmt.Errorf("PostgreSQL download not yet supported on %s", goos)
	}
	return &BinaryInfo{
		URL:             fmt.Sprintf("https://get.enterprisedb.com/postgresql/postgresql-%s-1-osx-binaries.zip", version),
		IsArchive:       true,
		StripComponents: 1,
	}, nil
}
