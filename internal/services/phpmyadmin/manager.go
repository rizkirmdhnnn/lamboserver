package phpmyadmin

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// version is the pinned phpMyAdmin version downloaded from files.phpmyadmin.net.
const version = "5.2.2"

// downloadURL is the template for the phpMyAdmin zip archive URL.
const downloadURL = "https://files.phpmyadmin.net/phpMyAdmin/%s/phpMyAdmin-%s-all-languages.zip"

// domain is the local development hostname for phpMyAdmin.
const domain = "phpmyadmin.test"

// osFileSystem is the production implementation of FileSystem using the os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (os.FileInfo, error)              { return os.Stat(name) }
func (osFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osFileSystem) RemoveAll(path string) error                   { return os.RemoveAll(path) }
func (osFileSystem) Remove(name string) error                      { return os.Remove(name) }
func (osFileSystem) Create(name string) (*os.File, error)          { return os.Create(name) }

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
	return system.RunCommand(name, args...)
}

// Manager handles phpMyAdmin download, configuration generation, Nginx vhost
// provisioning, and lifecycle management.
type Manager struct {
	paths    *system.Paths
	certMgr  CertManager
	nginxRld NginxReloader
	fs       FileSystem
	cmd      CommandRunner
}

// NewManager creates a Manager with injected dependencies. paths provides directory
// locations; certMgr generates and removes SSL certificates; nginxRld reloads Nginx
// after vhost changes; fs is the filesystem abstraction; cmd runs shell commands.
func NewManager(paths *system.Paths, certMgr CertManager, nginxRld NginxReloader, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{
		paths:    paths,
		certMgr:  certMgr,
		nginxRld: nginxRld,
		fs:       fs,
		cmd:      cmd,
	}
}

// Install downloads phpMyAdmin 5.2.2, generates config.inc.php, provisions the Nginx
// vhost with SSL and security deny rules, and reloads Nginx.
func (m *Manager) Install() error {
	destDir := filepath.Clean(m.paths.PhpMyAdminDir())
	zipPath := destDir + ".zip"
	extractDir := destDir + "-extract"

	url := fmt.Sprintf(downloadURL, version, version)

	// Step 1: Download the zip archive.
	_, err := m.cmd.Run("sh", "-c",
		fmt.Sprintf("curl -sL '%s' -o '%s'", url, zipPath))
	if err != nil {
		return fmt.Errorf("failed to download phpMyAdmin: %w", err)
	}

	// Step 2: Extract the archive to a staging directory.
	_, err = m.cmd.Run("sh", "-c",
		fmt.Sprintf("unzip -q '%s' -d '%s'", zipPath, extractDir))
	if err != nil {
		return fmt.Errorf("failed to extract phpMyAdmin: %w", err)
	}

	// Step 3: Move extracted contents into the final directory and remove staging dir.
	_, err = m.cmd.Run("sh", "-c",
		fmt.Sprintf("mv '%s'/phpMyAdmin-*/* '%s'/ && rm -rf '%s'", extractDir, destDir, extractDir))
	if err != nil {
		return fmt.Errorf("failed to move phpMyAdmin files: %w", err)
	}

	// Step 4: Clean up zip (best-effort).
	m.cmd.Run("rm", "-f", zipPath) //nolint:errcheck

	// Step 5: Write config.inc.php.
	if err := m.writeConfig(); err != nil {
		return err
	}

	// Step 6: Generate SSL certificate.
	certPath, keyPath, err := m.certMgr.GenerateCert(domain)
	if err != nil {
		return fmt.Errorf("failed to generate SSL certificate: %w", err)
	}

	// Step 7: Write Nginx vhost.
	if err := m.writeVhost(certPath, keyPath); err != nil {
		return err
	}

	// Step 8: Reload Nginx.
	if err := m.nginxRld.Reload(); err != nil {
		return fmt.Errorf("failed to reload Nginx: %w", err)
	}

	return nil
}

// writeConfig generates and writes config.inc.php with a secure blowfish_secret,
// auto-login auth type, root user, and MySQL socket path.
func (m *Manager) writeConfig() error {
	secret := generateBlowfishSecret()

	conf := fmt.Sprintf(`<?php
$cfg['blowfish_secret'] = '%s';

$i = 0;
$i++;
$cfg['Servers'][$i]['auth_type'] = 'config';
$cfg['Servers'][$i]['user'] = 'root';
$cfg['Servers'][$i]['password'] = '';
$cfg['Servers'][$i]['socket'] = '%s';
$cfg['Servers'][$i]['connect_type'] = 'socket';
$cfg['Servers'][$i]['compress'] = false;
$cfg['Servers'][$i]['AllowNoPassword'] = true;

$cfg['UploadDir'] = '';
$cfg['SaveDir'] = '';
`, secret, m.paths.MySQLSocket())

	if err := m.fs.WriteFile(m.paths.PhpMyAdminConf(), []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write config.inc.php: %w", err)
	}
	return nil
}

// generateBlowfishSecret generates a cryptographically random 32-character string
// using crypto/rand. Each byte is mapped to the printable alphanumeric character set.
func generateBlowfishSecret() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	rand.Read(b) //nolint:errcheck
	secret := make([]byte, 32)
	for i := range b {
		secret[i] = chars[b[i]%byte(len(chars))]
	}
	return string(secret)
}

// vhostData holds the template variables for the phpMyAdmin Nginx vhost.
type vhostData struct {
	DocumentRoot  string
	CertPath      string
	KeyPath       string
	FpmSocket     string
	FastCGIParams string
}

const vhostTemplate = `server {
    listen 127.0.0.1:80;
    listen 127.0.0.1:443 ssl;
    server_name phpmyadmin.test;
    root {{.DocumentRoot}};

    ssl_certificate {{.CertPath}};
    ssl_certificate_key {{.KeyPath}};

    index index.php;
    client_max_body_size 512M;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:{{.FpmSocket}};
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include {{.FastCGIParams}};
    }

    location ~ ^/(libraries|setup|doc|locale)/ {
        deny all;
        return 404;
    }

    location ~ /\.ht {
        deny all;
    }
}
`

// writeVhost renders and writes the Nginx vhost configuration for phpMyAdmin.
func (m *Manager) writeVhost(certPath, keyPath string) error {
	tmpl, err := template.New("vhost").Parse(vhostTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse vhost template: %w", err)
	}

	data := vhostData{
		DocumentRoot:  m.paths.PhpMyAdminDir(),
		CertPath:      certPath,
		KeyPath:       keyPath,
		FpmSocket:     filepath.Join(m.paths.Home, "php-fpm.sock"),
		FastCGIParams: m.paths.NginxFastCGIParams(),
	}

	vhostPath := filepath.Join(m.paths.NginxSitesDir(), "phpmyadmin.conf")
	f, err := m.fs.Create(vhostPath)
	if err != nil {
		return fmt.Errorf("failed to create vhost file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render vhost template: %w", err)
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write vhost content: %w", err)
	}

	return nil
}

// Uninstall removes all phpMyAdmin files, the Nginx vhost, the SSL certificate,
// and reloads Nginx.
func (m *Manager) Uninstall() error {
	if err := m.fs.RemoveAll(m.paths.PhpMyAdminDir()); err != nil {
		return fmt.Errorf("failed to remove phpMyAdmin directory: %w", err)
	}

	vhostPath := filepath.Join(m.paths.NginxSitesDir(), "phpmyadmin.conf")
	if err := m.fs.Remove(vhostPath); err != nil {
		return fmt.Errorf("failed to remove vhost config: %w", err)
	}

	if err := m.certMgr.RemoveCert(domain); err != nil {
		return fmt.Errorf("failed to remove SSL certificate: %w", err)
	}

	if err := m.nginxRld.Reload(); err != nil {
		return fmt.Errorf("failed to reload Nginx: %w", err)
	}

	return nil
}

// IsInstalled reports whether phpMyAdmin is installed by checking for the sentinel
// index.php file in the phpMyAdmin directory.
func (m *Manager) IsInstalled() bool {
	_, err := m.fs.Stat(filepath.Join(m.paths.PhpMyAdminDir(), "index.php"))
	return err == nil
}

// Status returns the current installation state of phpMyAdmin.
func (m *Manager) Status() ServiceStatus {
	return ServiceStatus{Installed: m.IsInstalled()}
}
