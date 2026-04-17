package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ServiceType distinguishes between system-level daemons and user-level agents.
type ServiceType int

const (
	ServiceDaemon ServiceType = iota // /Library/LaunchDaemons (root)
	ServiceAgent                     // ~/Library/LaunchAgents (user)
)

// ServiceConfig holds all parameters needed to generate and install a launchd plist.
type ServiceConfig struct {
	Label                string
	Program              string
	Args                 []string
	RunAtLoad            bool
	KeepAlive            bool
	WorkingDir           string
	StdoutPath           string
	StderrPath           string
	UserName             string // optional; defaults to "root" for daemons if empty
	Type                 ServiceType
	EnvironmentVariables map[string]string
}

// IsDaemon returns true if this service runs as a system daemon (root).
func (c ServiceConfig) IsDaemon() bool {
	return c.Type == ServiceDaemon
}

// ServiceStatus reports the current running state of a launchd service.
type ServiceStatus struct {
	Label   string `json:"label"`
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
}

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
{{- if .IsDaemon}}
	<key>UserName</key>
	<string>{{if .UserName}}{{.UserName}}{{else}}root{{end}}</string>
{{- end}}
	<key>ProgramArguments</key>
	<array>
		<string>{{.Program}}</string>
{{- range .Args}}
		<string>{{.}}</string>
{{- end}}
	</array>
	<key>RunAtLoad</key>
	{{if .RunAtLoad}}<true/>{{else}}<false/>{{end}}
	<key>KeepAlive</key>
	{{if .KeepAlive}}<true/>{{else}}<false/>{{end}}
{{- if .WorkingDir}}
	<key>WorkingDirectory</key>
	<string>{{.WorkingDir}}</string>
{{- end}}
{{- if .StdoutPath}}
	<key>StandardOutPath</key>
	<string>{{.StdoutPath}}</string>
{{- end}}
{{- if .StderrPath}}
	<key>StandardErrorPath</key>
	<string>{{.StderrPath}}</string>
{{- end}}
{{- if .EnvironmentVariables}}
	<key>EnvironmentVariables</key>
	<dict>
	{{- range $key, $value := .EnvironmentVariables}}
		<key>{{$key}}</key>
		<string>{{$value}}</string>
	{{- end}}
	</dict>
{{- end}}
</dict>
</plist>`

// LaunchdManager installs, uninstalls, and controls macOS launchd services
// (both system-level LaunchDaemons and user-level LaunchAgents). It delegates
// privileged daemon operations to the Helper to avoid repeated password prompts.
type LaunchdManager struct {
	paths  *Paths
	helper *Helper
}

// NewLaunchdManager creates a LaunchdManager with the given path configuration.
func NewLaunchdManager(paths *Paths) *LaunchdManager {
	return &LaunchdManager{
		paths:  paths,
		helper: NewHelper(paths),
	}
}

// Helper returns the privileged helper for setup
func (m *LaunchdManager) Helper() *Helper {
	return m.helper
}

func (m *LaunchdManager) plistPath(cfg ServiceConfig) string {
	if cfg.Type == ServiceDaemon {
		return filepath.Join(m.paths.LaunchDaemonsDir(), cfg.Label+".plist")
	}
	return filepath.Join(m.paths.LaunchAgentsDir(), cfg.Label+".plist")
}

// Install generates a plist from cfg, writes it to the appropriate launchd directory,
// and loads the service. Daemons are installed via the privileged helper; agents are
// installed directly without admin privileges.
func (m *LaunchdManager) Install(cfg ServiceConfig) error {
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse plist template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("failed to execute plist template: %w", err)
	}

	plistContent := buf.String()

	if cfg.Type == ServiceDaemon {
		// Write plist to temp, then use helper to install as daemon (no password)
		tmpFile := filepath.Join(os.TempDir(), cfg.Label+".plist")
		if err := os.WriteFile(tmpFile, []byte(plistContent), 0644); err != nil {
			return err
		}
		_, err := m.helper.Run("install-daemon", tmpFile)
		os.Remove(tmpFile)
		return err
	}

	// User agent - no admin needed
	plistPath := m.plistPath(cfg)
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return err
	}
	_, err = RunCommand("launchctl", "load", plistPath)
	return err
}

// Uninstall unloads and removes the plist for the given service configuration.
// Daemons are uninstalled via the privileged helper; agents are removed directly.
func (m *LaunchdManager) Uninstall(cfg ServiceConfig) error {
	if cfg.Type == ServiceDaemon {
		_, err := m.helper.Run("uninstall-daemon", cfg.Label)
		return err
	}

	plistPath := m.plistPath(cfg)
	RunCommand("launchctl", "unload", plistPath)
	return os.Remove(plistPath)
}

// Start sends a start signal to the launchd service identified by label.
func (m *LaunchdManager) Start(label string) error {
	_, err := RunCommand("launchctl", "start", label)
	return err
}

// Stop sends a stop signal to the launchd service identified by label.
func (m *LaunchdManager) Stop(label string) error {
	_, err := RunCommand("launchctl", "stop", label)
	return err
}

// IsRunning reports whether the launchd service identified by label is currently
// running. It checks user-level agents via launchctl list and system-level daemons
// via plist existence combined with pgrep.
func (m *LaunchdManager) IsRunning(label string) bool {
	// Check user-level agents
	output, err := RunCommand("launchctl", "list")
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, label) {
				fields := strings.Fields(line)
				if len(fields) >= 1 && fields[0] != "-" {
					return true
				}
			}
		}
	}

	// Check system-level daemons via plist existence + process name
	daemonPlist := filepath.Join(m.paths.LaunchDaemonsDir(), label+".plist")
	if _, err := os.Stat(daemonPlist); err == nil {
		parts := strings.Split(label, ".")
		binaryName := parts[len(parts)-1]
		_, err := RunCommand("pgrep", binaryName)
		return err == nil
	}

	return false
}

// GetStatus returns the current ServiceStatus for the launchd service identified by label.
func (m *LaunchdManager) GetStatus(label string) ServiceStatus {
	return ServiceStatus{
		Label:   label,
		Running: m.IsRunning(label),
	}
}
