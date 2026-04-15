package debug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	mu      sync.Mutex
	enabled bool
	file    *os.File
	logger  *log.Logger
	logPath string
}

func NewLogger(logDir string) *Logger {
	logPath := filepath.Join(logDir, "debug.log")
	return &Logger{
		logPath: logPath,
	}
}

func (l *Logger) Enable() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.enabled {
		return nil
	}

	f, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open debug log: %w", err)
	}

	l.file = f
	l.logger = log.New(f, "", 0)
	l.enabled = true
	l.write("DEBUG", "Debug mode enabled")
	return nil
}

func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled {
		return
	}

	l.write("DEBUG", "Debug mode disabled")
	l.enabled = false
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	l.logger = nil
}

func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

func (l *Logger) LogPath() string {
	return l.logPath
}

func (l *Logger) write(level, msg string) {
	if l.logger == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	l.logger.Printf("[%s] [%s] %s", timestamp, level, msg)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.enabled {
		return
	}
	l.write("INFO", fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.enabled {
		return
	}
	l.write("ERROR", fmt.Sprintf(format, args...))
}

// Action logs a user-initiated action with its result
func (l *Logger) Action(action string, err error) {
	if err != nil {
		l.Error("ACTION %s -> FAILED: %v", action, err)
	} else {
		l.Info("ACTION %s -> OK", action)
	}
}

// ClearLog truncates the debug log file
func (l *Logger) ClearLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Truncate(0)
	}
	return os.Truncate(l.logPath, 0)
}
