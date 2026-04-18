package dnsmasq

import (
	"bufio"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/services"
)

// ServiceAdapter wraps the existing DNS Manager to implement the unified
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

// Restart stops then starts the dnsmasq service.
func (a *ServiceAdapter) Restart() error {
	if err := a.Stop(); err != nil {
		return err
	}
	return a.Start()
}

// Install installs the dnsmasq binary. The version parameter is ignored.
func (a *ServiceAdapter) Install(version string) error {
	return a.binary.Install()
}

// Status returns the unified service status for dnsmasq.
func (a *ServiceAdapter) Status() services.ServiceStatus {
	s := a.Manager.Status()
	if s.Running {
		return services.StatusRunning
	}
	return services.StatusStopped
}

// Logs reads the last 100 lines from the dnsmasq stderr log.
func (a *ServiceAdapter) Logs() ([]string, error) {
	logPath := a.paths.LogsDir() + "/dnsmasq-stderr.log"
	return readLastNLines(logPath, 100)
}

// Version returns "bundled" since dnsmasq uses an embedded binary.
func (a *ServiceAdapter) Version() string {
	if a.binary.Find() == "" {
		return ""
	}
	return "bundled"
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
