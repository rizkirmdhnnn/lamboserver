// Package nginx manages Nginx installation and LaunchDaemon lifecycle for local
// reverse-proxy and static file serving.
package nginx

import (
	"fmt"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// ServiceLabel is the launchd daemon label used for the Nginx service.
const ServiceLabel = "com.lamboserver.nginx"

// ServiceStatus reports whether the Nginx binary is present and the daemon is running.
type ServiceStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
}

// Manager handles Nginx binary discovery, configuration generation, and LaunchDaemon
// lifecycle. Nginx runs as a root LaunchDaemon so it can bind to port 80 and 443.
type Manager struct {
	paths   *system.Paths
	launchd LaunchdService
	binary  *system.BinaryLocator
	fs      FileSystem
	helper  HelperRunner
}

// NewManager creates a Manager with injected dependencies. paths provides directory
// locations; launchd manages the system service; helper is used for privileged operations
// such as nginx reload that require root access.
func NewManager(paths *system.Paths, launchd LaunchdService, fs FileSystem, helper HelperRunner) *Manager {
	return &Manager{
		paths:   paths,
		launchd: launchd,
		fs:      fs,
		helper:  helper,
		binary: &system.BinaryLocator{
			Name:      "nginx",
			LocalPath: paths.NginxBin(),
		},
	}
}

// IsInstalled reports whether the nginx binary is present at the expected path or on system PATH.
func (m *Manager) IsInstalled() bool { return m.binary.IsInstalled() }

// IsDaemonInstalled checks if the LaunchDaemon plist is already installed
func (m *Manager) IsDaemonInstalled() bool {
	_, err := m.fs.Stat("/Library/LaunchDaemons/" + ServiceLabel + ".plist")
	return err == nil
}

// Start installs nginx as a root LaunchDaemon and starts it. It generates the master
// nginx.conf, ensures log files exist with user ownership, then calls launchd.Install.
// If nginx is not yet installed, it attempts to install the binary first.
func (m *Manager) Start() error {
	if !m.IsInstalled() {
		if err := m.binary.Install(); err != nil {
			return fmt.Errorf("nginx is not installed: %w", err)
		}
	}

	if err := m.EnsureConfig(); err != nil {
		return fmt.Errorf("failed to generate nginx config: %w", err)
	}

	// Ensure log files exist with user ownership before daemon starts as root
	ensureLogFile(m.paths.LogsDir() + "/nginx-stdout.log")
	ensureLogFile(m.paths.LogsDir() + "/nginx-stderr.log")
	ensureLogFile(m.paths.NginxErrorLog())
	ensureLogFile(m.paths.NginxAccessLog())

	// Install as LaunchDaemon (asks admin password once, then auto-starts on boot)
	return m.launchd.Install(system.ServiceConfig{
		Label:      ServiceLabel,
		Program:    m.binary.Find(),
		Args:       []string{"-c", m.paths.NginxConf(), "-g", "daemon off;"},
		RunAtLoad:  true,
		KeepAlive:  true,
		StdoutPath: m.paths.LogsDir() + "/nginx-stdout.log",
		StderrPath: m.paths.LogsDir() + "/nginx-stderr.log",
		Type:       system.ServiceDaemon,
	})
}

// Stop uninstalls the Nginx LaunchDaemon, which terminates the nginx process.
func (m *Manager) Stop() error {
	return m.launchd.Uninstall(system.ServiceConfig{
		Label: ServiceLabel,
		Type:  system.ServiceDaemon,
	})
}

// Reload sends a configuration reload signal to the running nginx master process via the
// privileged helper. This is needed because nginx runs as root and requires root access
// to send signals to its master process.
func (m *Manager) Reload() error {
	// Reload via helper (root) since nginx master runs as root LaunchDaemon
	_, err := m.helper.Run("reload-nginx")
	return err
}

// Status returns a snapshot of the current Nginx service state.
func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{
		Installed: m.IsInstalled(),
		Running:   m.launchd.IsRunning(ServiceLabel),
	}
}

// EnsureConfig regenerates the master nginx.conf, fastcgi_params, mime.types, and the
// default site config. It is called before Start and whenever site configs change.
func (m *Manager) EnsureConfig() error {
	return m.generateMasterConfig()
}

// generateMasterConfig delegates to the standalone config generation function.
// os.* calls in config.go remain as-is (standalone helpers, not Manager methods).
// They are candidates for a future refactor if config.go is converted to Manager methods.
func (m *Manager) generateMasterConfig() error {
	username := os.Getenv("USER")
	if username == "" {
		username = "nobody"
	}

	conf := fmt.Sprintf(`user %s staff;
worker_processes auto;
error_log %s;
pid %s/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include %s;
    default_type application/octet-stream;

    access_log %s;

    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 512M;

    include %s/*.conf;
}
`,
		username,
		m.paths.NginxErrorLog(),
		m.paths.NginxDir(),
		m.paths.NginxMimeTypes(),
		m.paths.NginxAccessLog(),
		m.paths.NginxSitesDir(),
	)

	if err := m.fs.WriteFile(m.paths.NginxConf(), []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write nginx.conf: %w", err)
	}

	if err := m.writeFastCGIParams(); err != nil {
		return err
	}

	if _, err := m.fs.Stat(m.paths.NginxMimeTypes()); os.IsNotExist(err) {
		return m.writeMimeTypes()
	}

	if err := m.writeDefaultSite(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) writeFastCGIParams() error {
	params := `fastcgi_param  QUERY_STRING       $query_string;
fastcgi_param  REQUEST_METHOD     $request_method;
fastcgi_param  CONTENT_TYPE       $content_type;
fastcgi_param  CONTENT_LENGTH     $content_length;

fastcgi_param  SCRIPT_NAME        $fastcgi_script_name;
fastcgi_param  REQUEST_URI        $request_uri;
fastcgi_param  DOCUMENT_URI       $document_uri;
fastcgi_param  DOCUMENT_ROOT      $document_root;
fastcgi_param  SERVER_PROTOCOL    $server_protocol;
fastcgi_param  REQUEST_SCHEME     $scheme;

fastcgi_param  GATEWAY_INTERFACE  CGI/1.1;
fastcgi_param  SERVER_SOFTWARE    LamboServer;

fastcgi_param  REMOTE_ADDR        $remote_addr;
fastcgi_param  REMOTE_PORT        $remote_port;
fastcgi_param  SERVER_ADDR        $server_addr;
fastcgi_param  SERVER_PORT        $server_port;
fastcgi_param  SERVER_NAME        $server_name;

fastcgi_param  REDIRECT_STATUS    200;
`
	return m.fs.WriteFile(m.paths.NginxFastCGIParams(), []byte(params), 0644)
}

func (m *Manager) writeMimeTypes() error {
	mimeTypes := `types {
    text/html                             html htm shtml;
    text/css                              css;
    text/xml                              xml;
    application/javascript                js;
    application/json                      json;
    image/png                             png;
    image/jpeg                            jpeg jpg;
    image/gif                             gif;
    image/svg+xml                         svg svgz;
    image/webp                            webp;
    application/font-woff                 woff;
    application/font-woff2                woff2;
    application/pdf                       pdf;
    application/zip                       zip;
    text/plain                            txt;
    application/x-httpd-php               php;
}
`
	return m.fs.WriteFile(m.paths.NginxMimeTypes(), []byte(mimeTypes), 0644)
}

func (m *Manager) writeDefaultSite() error {
	// Write default HTML page
	htmlPath := m.paths.DefaultSiteDir() + "/index.html"
	if _, err := m.fs.Stat(htmlPath); err == nil {
		return m.writeDefaultSiteConf()
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>LamboServer</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #18181b;
    color: #fafafa;
  }
  .container {
    text-align: center;
    padding: 2rem;
  }
  .logo {
    font-size: 3rem;
    font-weight: 800;
    color: #f97316;
    letter-spacing: -0.03em;
    margin-bottom: 0.5rem;
  }
  .subtitle {
    font-size: 1.1rem;
    color: #a1a1aa;
    margin-bottom: 2rem;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: #27272a;
    border: 1px solid #3f3f46;
    border-radius: 999px;
    padding: 8px 20px;
    font-size: 0.875rem;
    color: #a1a1aa;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #22c55e;
    box-shadow: 0 0 6px #22c55e;
  }
  .info {
    margin-top: 2rem;
    font-size: 0.8rem;
    color: #71717a;
  }
  .info code {
    background: #27272a;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
  }
</style>
</head>
<body>
<div class="container">
  <div class="logo">LamboServer</div>
  <div class="subtitle">Your local development environment is ready.</div>
  <div class="status">
    <span class="dot"></span>
    Nginx is running
  </div>
  <div class="info">
    Link a project in the app to get started, or visit <code>yourproject.test</code>
  </div>
</div>
</body>
</html>`

	if err := m.fs.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write default site HTML: %w", err)
	}

	return m.writeDefaultSiteConf()
}

func (m *Manager) writeDefaultSiteConf() error {
	conf := fmt.Sprintf(`server {
    listen 127.0.0.1:80 default_server;
    server_name _;
    root %s;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
`, m.paths.DefaultSiteDir())

	confPath := m.paths.NginxSitesDir() + "/00-default.conf"
	return m.fs.WriteFile(confPath, []byte(conf), 0644)
}
