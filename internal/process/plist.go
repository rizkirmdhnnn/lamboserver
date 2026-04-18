package process

import (
	"fmt"
	"strings"
	"text/template"
)

// PlistConfig holds the data needed to render a launchd plist XML file.
// This is a value type mirroring the fields from system.ServiceConfig
// that the plist template needs. Defined here (not importing system)
// to prevent circular dependencies per STRC-04.
type PlistConfig struct {
	Label                string
	Program              string
	Args                 []string
	RunAtLoad            bool
	KeepAlive            bool
	WorkingDir           string
	StdoutPath           string
	StderrPath           string
	IsDaemon             bool
	UserName             string // optional; defaults to "root" for daemons if empty
	EnvironmentVariables map[string]string
}

// GeneratePlist renders a PlistConfig into a launchd plist XML string.
func GeneratePlist(cfg PlistConfig) (string, error) {
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse plist template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute plist template: %w", err)
	}
	return buf.String(), nil
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
