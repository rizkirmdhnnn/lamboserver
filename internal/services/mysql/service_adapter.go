package mysql

import (
	"bufio"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/services"
)

// ServiceAdapter wraps the existing MySQL Manager to implement the unified
// services.Service interface without modifying the original Manager's method signatures.
type ServiceAdapter struct {
	*Manager
}

// NewServiceAdapter creates a ServiceAdapter wrapping the given Manager.
func NewServiceAdapter(m *Manager) *ServiceAdapter {
	return &ServiceAdapter{Manager: m}
}

// Compile-time interface check.
var _ services.Service = (*ServiceAdapter)(nil)

// Restart stops then starts the MySQL service.
func (a *ServiceAdapter) Restart() error {
	if err := a.Stop(); err != nil {
		return err
	}
	return a.Start()
}

// Install downloads and installs MySQL. The version parameter is ignored since
// MySQL uses a pinned version constant.
func (a *ServiceAdapter) Install(version string) error {
	return a.Manager.Install()
}

// Status returns the unified service status for MySQL.
// Returns "not_installed" when the mysqld binary is absent so the frontend
// can distinguish "not installed" from "installed but stopped".
func (a *ServiceAdapter) Status() services.ServiceStatus {
	if !a.IsInstalled() {
		return services.StatusNotInstalled
	}
	s := a.Manager.Status()
	if s.Running {
		return services.StatusRunning
	}
	return services.StatusStopped
}

// Logs reads the last 100 lines from the MySQL error log.
func (a *ServiceAdapter) Logs() ([]string, error) {
	logPath := a.paths.LogsDir() + "/mysql-error.log"
	return readLastNLines(logPath, 100)
}

// Version returns the pinned MySQL version.
func (a *ServiceAdapter) Version() string {
	if !a.IsInstalled() {
		return ""
	}
	return mysqlVersion
}

// ensureLogFile creates an empty file at path if it does not already exist.
func ensureLogFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte{}, 0644)
	}
}

func readLastNLines(filePath string, n int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
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
