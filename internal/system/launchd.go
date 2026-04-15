package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ServiceType int

const (
	ServiceDaemon ServiceType = iota // /Library/LaunchDaemons (root)
	ServiceAgent                     // ~/Library/LaunchAgents (user)
)

type ServiceConfig struct {
	Label                string
	Program              string
	Args                 []string
	RunAtLoad            bool
	KeepAlive            bool
	WorkingDir           string
	StdoutPath           string
	StderrPath           string
	Type                 ServiceType
	EnvironmentVariables map[string]string
}

// IsDaemon returns true if this service runs as a system daemon (root).
func (c ServiceConfig) IsDaemon() bool {
	return c.Type == ServiceDaemon
}

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
	<string>root</string>
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

type LaunchdManager struct {
	paths  *Paths
	helper *Helper
}

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

func (m *LaunchdManager) Uninstall(cfg ServiceConfig) error {
	if cfg.Type == ServiceDaemon {
		_, err := m.helper.Run("uninstall-daemon", cfg.Label)
		return err
	}

	plistPath := m.plistPath(cfg)
	RunCommand("launchctl", "unload", plistPath)
	return os.Remove(plistPath)
}

func (m *LaunchdManager) Start(label string) error {
	_, err := RunCommand("launchctl", "start", label)
	return err
}

func (m *LaunchdManager) Stop(label string) error {
	_, err := RunCommand("launchctl", "stop", label)
	return err
}

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

func (m *LaunchdManager) GetStatus(label string) ServiceStatus {
	return ServiceStatus{
		Label:   label,
		Running: m.IsRunning(label),
	}
}
