package nginx

import (
	"fmt"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

func (m *Manager) generateMasterConfig() error {
	// Ensure default site exists
	if err := writeDefaultSite(m.paths); err != nil {
		return err
	}

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

	if err := os.WriteFile(m.paths.NginxConf(), []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write nginx.conf: %w", err)
	}

	if err := writeFastCGIParams(m.paths); err != nil {
		return err
	}

	if _, err := os.Stat(m.paths.NginxMimeTypes()); os.IsNotExist(err) {
		return writeMimeTypes(m.paths)
	}
	return nil
}

func writeFastCGIParams(paths *system.Paths) error {
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
	return os.WriteFile(paths.NginxFastCGIParams(), []byte(params), 0644)
}

func writeMimeTypes(paths *system.Paths) error {
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
	return os.WriteFile(paths.NginxMimeTypes(), []byte(mimeTypes), 0644)
}

func writeDefaultSite(paths *system.Paths) error {
	// Write default HTML page
	htmlPath := paths.DefaultSiteDir() + "/index.html"
	if _, err := os.Stat(htmlPath); err == nil {
		return writeDefaultSiteConf(paths)
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

	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write default site HTML: %w", err)
	}

	return writeDefaultSiteConf(paths)
}

func writeDefaultSiteConf(paths *system.Paths) error {
	conf := fmt.Sprintf(`server {
    listen 127.0.0.1:80 default_server;
    server_name _;
    root %s;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
`, paths.DefaultSiteDir())

	confPath := paths.NginxSitesDir() + "/00-default.conf"
	return os.WriteFile(confPath, []byte(conf), 0644)
}
