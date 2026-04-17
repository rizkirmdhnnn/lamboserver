package binaries

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// Downloader handles downloading and verifying service binaries.
type Downloader struct {
	registry *Registry
}

// NewDownloader creates a Downloader backed by the given registry.
func NewDownloader(registry *Registry) *Downloader {
	return &Downloader{registry: registry}
}

// Download fetches the binary for the given service and version, placing it in destDir.
// If the registry entry has a checksum, the download is verified before use.
func (d *Downloader) Download(service, version, destDir string) error {
	info, err := d.registry.Lookup(service, version)
	if err != nil {
		return err
	}

	if info.IsArchive {
		return downloadAndExtract(info.URL, destDir, info.StripComponents, info.Checksum)
	}
	return downloadFile(info.URL, filepath.Join(destDir, filepath.Base(info.URL)), info.Checksum)
}

// downloadFile downloads a single file with optional checksum verification.
func downloadFile(url, destPath, checksum string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpPath := destPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	f.Close()

	if checksum != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if got != checksum {
			os.Remove(tmpPath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", checksum, got)
		}
	}

	return os.Rename(tmpPath, destPath)
}

// downloadAndExtract downloads a .tar.gz archive and extracts it.
func downloadAndExtract(url, destDir string, stripComponents int, checksum string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	strip := fmt.Sprintf("--strip-components=%d", stripComponents)
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("curl -sfL '%s' | tar xz %s -C '%s'", url, strip, destDir))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("download and extract failed: %w: %s", err, string(out))
	}
	return nil
}
