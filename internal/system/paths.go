package system

import (
	"os"
	"path/filepath"
)

const (
	AppName    = "lamboserver"
	AppDirName = ".lamboserver"
	TestTLD    = "test"
)

type Paths struct {
	Home string
}

func NewPaths() *Paths {
	home, _ := os.UserHomeDir()
	return &Paths{
		Home: filepath.Join(home, AppDirName),
	}
}

// Root directories
func (p *Paths) BinDir() string      { return filepath.Join(p.Home, "bin") }
func (p *Paths) PhpDir() string      { return filepath.Join(p.Home, "php") }
func (p *Paths) NodeDir() string     { return filepath.Join(p.Home, "nodejs") }
func (p *Paths) NginxDir() string    { return filepath.Join(p.Home, "nginx") }
func (p *Paths) CertsDir() string    { return filepath.Join(p.Home, "certs") }
func (p *Paths) LogsDir() string     { return filepath.Join(p.Home, "logs") }
func (p *Paths) ServicesDir() string  { return filepath.Join(p.Home, "services") }
func (p *Paths) DnsmasqDir() string   { return filepath.Join(p.Home, "dnsmasq") }

// Config
func (p *Paths) ConfigFile() string { return filepath.Join(p.Home, "config.json") }

// Nginx
func (p *Paths) NginxBin() string       { return filepath.Join(p.BinDir(), "nginx") }
func (p *Paths) NginxConf() string      { return filepath.Join(p.NginxDir(), "nginx.conf") }
func (p *Paths) NginxSitesDir() string  { return filepath.Join(p.NginxDir(), "sites") }
func (p *Paths) NginxLogsDir() string   { return filepath.Join(p.NginxDir(), "logs") }
func (p *Paths) NginxAccessLog() string { return filepath.Join(p.NginxLogsDir(), "access.log") }
func (p *Paths) NginxErrorLog() string  { return filepath.Join(p.NginxLogsDir(), "error.log") }
func (p *Paths) NginxMimeTypes() string { return filepath.Join(p.NginxDir(), "mime.types") }
func (p *Paths) NginxFastCGIParams() string { return filepath.Join(p.NginxDir(), "fastcgi_params") }
func (p *Paths) DefaultSiteDir() string     { return filepath.Join(p.Home, "default-site") }

// Dnsmasq
func (p *Paths) DnsmasqBin() string  { return filepath.Join(p.BinDir(), "dnsmasq") }
func (p *Paths) DnsmasqConf() string { return filepath.Join(p.DnsmasqDir(), "dnsmasq.conf") }
func (p *Paths) ResolverDir() string { return "/etc/resolver" }
func (p *Paths) ResolverFile() string {
	return filepath.Join(p.ResolverDir(), TestTLD)
}

// Certs
func (p *Paths) CACert() string    { return filepath.Join(p.CertsDir(), "ca.pem") }
func (p *Paths) CAKey() string     { return filepath.Join(p.CertsDir(), "ca-key.pem") }
func (p *Paths) SiteCert(domain string) string {
	return filepath.Join(p.CertsDir(), domain+".pem")
}
func (p *Paths) SiteKey(domain string) string {
	return filepath.Join(p.CertsDir(), domain+"-key.pem")
}

// PHP
func (p *Paths) PhpVersionDir(version string) string {
	return filepath.Join(p.PhpDir(), version)
}
func (p *Paths) PhpCurrentLink() string {
	return filepath.Join(p.PhpDir(), "current")
}
func (p *Paths) PhpFpmSocket(version string) string {
	return filepath.Join(p.Home, "php", version, "var", "run", "php-fpm.sock")
}

// Node.js
func (p *Paths) NodeVersionDir(version string) string {
	return filepath.Join(p.NodeDir(), version)
}
func (p *Paths) NodeCurrentLink() string {
	return filepath.Join(p.NodeDir(), "current")
}

// Launchd
func (p *Paths) LaunchDaemonsDir() string {
	return "/Library/LaunchDaemons"
}
func (p *Paths) LaunchAgentsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents")
}

// EnsureDirectories creates all required directories
func (p *Paths) EnsureDirectories() error {
	dirs := []string{
		p.Home,
		p.BinDir(),
		p.PhpDir(),
		p.NodeDir(),
		p.NginxDir(),
		p.NginxSitesDir(),
		p.NginxLogsDir(),
		p.CertsDir(),
		p.LogsDir(),
		p.ServicesDir(),
		p.DnsmasqDir(),
		p.DefaultSiteDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
