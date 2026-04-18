package system

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadFile downloads a URL to a local file path
func DownloadFile(url, destPath string) error {
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

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	f.Close()

	return os.Rename(tmpPath, destPath)
}

// DownloadAndExtract downloads a .tar.gz and extracts to destDir
func DownloadAndExtract(url, destDir string, stripComponents int) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	strip := fmt.Sprintf("--strip-components=%d", stripComponents)
	_, err := RunCommand("sh", "-c",
		fmt.Sprintf("curl -sfL '%s' | tar xz %s -C '%s'", url, strip, destDir))
	if err != nil {
		return fmt.Errorf("download and extract failed: %w", err)
	}
	return nil
}
