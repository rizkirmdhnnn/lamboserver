package logger

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// LogFile represents a log file entry with its name, absolute path, and size in bytes.
type LogFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Reader scans the application log directories and reads log file contents.
type Reader struct {
	paths *system.Paths
}

// NewReader creates a Reader that looks for log files under the paths provided.
func NewReader(paths *system.Paths) *Reader {
	return &Reader{paths: paths}
}

// ListLogFiles returns metadata for all .log files found in the application
// logs directory and the Nginx logs directory.
func (r *Reader) ListLogFiles() []LogFile {
	var logs []LogFile

	dirs := []string{r.paths.LogsDir(), r.paths.NginxLogsDir()}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				logs = append(logs, LogFile{
					Name: entry.Name(),
					Path: filepath.Join(dir, entry.Name()),
					Size: info.Size(),
				})
			}
		}
	}
	return logs
}

// ReadLastN reads the given file and returns its last n lines.
func (r *Reader) ReadLastN(filePath string, n int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, scanner.Err()
}
