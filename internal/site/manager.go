package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

type Site struct {
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	PhpVersion string `json:"php_version"`
	SSLEnabled bool   `json:"ssl_enabled"`
	CreatedAt  string `json:"created_at"`
}

const siteConfTemplate = `server {
    listen 127.0.0.1:80;
    listen 127.0.0.1:443 ssl;
    server_name {{.Domain}};
    root {{.DocumentRoot}};

    ssl_certificate {{.CertPath}};
    ssl_certificate_key {{.KeyPath}};

    index index.php index.html index.htm;
    client_max_body_size 512M;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:{{.FpmSocket}};
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include {{.FastCGIParams}};
    }

    location ~ /\.ht {
        deny all;
    }
}
`

type siteTemplateData struct {
	Domain        string
	DocumentRoot  string
	CertPath      string
	KeyPath       string
	FpmSocket     string
	FastCGIParams string
}

type Manager struct {
	paths       *system.Paths
	store       *config.Store
	certManager *cert.Manager
}

func NewManager(paths *system.Paths, store *config.Store, certManager *cert.Manager) *Manager {
	return &Manager{
		paths:       paths,
		store:       store,
		certManager: certManager,
	}
}

func (m *Manager) Link(projectPath, domain string) error {
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", projectPath)
	}

	// Auto-detect document root (Laravel uses /public)
	documentRoot := projectPath
	publicDir := filepath.Join(projectPath, "public")
	if _, err := os.Stat(publicDir); err == nil {
		documentRoot = publicDir
	}

	if !strings.HasSuffix(domain, ".test") {
		domain = domain + ".test"
	}

	// Ensure CA exists
	if !m.certManager.IsCAInstalled() {
		if err := m.certManager.SetupCA(); err != nil {
			return fmt.Errorf("failed to setup CA: %w", err)
		}
		// Trust CA in keychain (uses helper, no extra password)
		if err := m.certManager.TrustCA(); err != nil {
			return fmt.Errorf("failed to trust CA: %w", err)
		}
	}

	// Generate SSL cert for this domain
	certPath, keyPath, err := m.certManager.GenerateCert(domain)
	if err != nil {
		return fmt.Errorf("failed to generate SSL cert: %w", err)
	}

	// FPM socket is at a fixed path managed by FpmManager
	fpmSocket := filepath.Join(m.paths.Home, "php-fpm.sock")

	data := siteTemplateData{
		Domain:        domain,
		DocumentRoot:  documentRoot,
		CertPath:      certPath,
		KeyPath:       keyPath,
		FpmSocket:     fpmSocket,
		FastCGIParams: m.paths.NginxFastCGIParams(),
	}

	tmpl, err := template.New("site").Parse(siteConfTemplate)
	if err != nil {
		return err
	}

	confPath := filepath.Join(m.paths.NginxSitesDir(), domain+".conf")
	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	return m.store.AddSite(domain, projectPath)
}

func (m *Manager) Unlink(domain string) error {
	if !strings.HasSuffix(domain, ".test") {
		domain = domain + ".test"
	}

	confPath := filepath.Join(m.paths.NginxSitesDir(), domain+".conf")
	os.Remove(confPath)

	m.certManager.RemoveCert(domain)

	return m.store.RemoveSite(domain)
}

func (m *Manager) List() []Site {
	siteConfigs := m.store.GetSites()
	sites := make([]Site, len(siteConfigs))
	for i, sc := range siteConfigs {
		sites[i] = Site{
			Domain:     sc.Domain,
			Path:       sc.Path,
			PhpVersion: sc.PhpVersion,
			SSLEnabled: sc.SSLEnabled,
			CreatedAt:  sc.CreatedAt,
		}
	}
	return sites
}
