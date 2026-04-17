package system

import (
	"os"
	"path/filepath"
)

// AppName is the internal identifier used in service labels and helper scripts.
const AppName = "lamboserver"

// AppDirName is the name of the application's home directory under the user's home.
const AppDirName = ".lamboserver"

// TestTLD is the local DNS top-level domain resolved by dnsmasq (e.g. myapp.test).
const TestTLD = "test"

// Paths is the single source of truth for all file and directory locations used
// by LamboServer. All paths are rooted under ~/.lamboserver/. If a location needs
// to change, only this file needs updating.
type Paths struct {
	Home   string // ~/.lamboserver/
	OSHome string // ~/  (the user's home directory)
}

// NewPaths creates a Paths instance rooted at ~/.lamboserver/.
func NewPaths() *Paths {
	home, _ := os.UserHomeDir()
	return &Paths{
		Home:   filepath.Join(home, AppDirName),
		OSHome: home,
	}
}

// BinDir returns the directory containing managed symlinks and helper binaries (~/.lamboserver/bin/).
func (p *Paths) BinDir() string { return filepath.Join(p.Home, "bin") }

// PhpDir returns the directory containing all installed PHP versions (~/.lamboserver/php/).
func (p *Paths) PhpDir() string { return filepath.Join(p.Home, "php") }

// NodeDir returns the directory containing all installed Node.js versions (~/.lamboserver/nodejs/).
func (p *Paths) NodeDir() string { return filepath.Join(p.Home, "nodejs") }

// NginxDir returns the Nginx configuration and data directory (~/.lamboserver/nginx/).
func (p *Paths) NginxDir() string { return filepath.Join(p.Home, "nginx") }

// CertsDir returns the directory storing CA and site SSL certificates (~/.lamboserver/certs/).
func (p *Paths) CertsDir() string { return filepath.Join(p.Home, "certs") }

// LogsDir returns the directory for application log files (~/.lamboserver/logs/).
func (p *Paths) LogsDir() string { return filepath.Join(p.Home, "logs") }

// ServicesDir returns the directory for service-related data (~/.lamboserver/services/).
func (p *Paths) ServicesDir() string { return filepath.Join(p.Home, "services") }

// DnsmasqDir returns the dnsmasq configuration directory (~/.lamboserver/dnsmasq/).
func (p *Paths) DnsmasqDir() string { return filepath.Join(p.Home, "dnsmasq") }

// ConfigFile returns the path to the application configuration JSON file (~/.lamboserver/config.json).
func (p *Paths) ConfigFile() string { return filepath.Join(p.Home, "config.json") }

// NginxBin returns the path to the managed Nginx binary (~/.lamboserver/bin/nginx).
func (p *Paths) NginxBin() string { return filepath.Join(p.BinDir(), "nginx") }

// NginxConf returns the path to the main Nginx configuration file.
func (p *Paths) NginxConf() string { return filepath.Join(p.NginxDir(), "nginx.conf") }

// NginxSitesDir returns the directory containing per-site Nginx configuration files.
func (p *Paths) NginxSitesDir() string { return filepath.Join(p.NginxDir(), "sites") }

// NginxLogsDir returns the directory containing Nginx access and error log files.
func (p *Paths) NginxLogsDir() string { return filepath.Join(p.NginxDir(), "logs") }

// NginxAccessLog returns the path to the Nginx access log file.
func (p *Paths) NginxAccessLog() string { return filepath.Join(p.NginxLogsDir(), "access.log") }

// NginxErrorLog returns the path to the Nginx error log file.
func (p *Paths) NginxErrorLog() string { return filepath.Join(p.NginxLogsDir(), "error.log") }

// NginxMimeTypes returns the path to the Nginx mime.types file.
func (p *Paths) NginxMimeTypes() string { return filepath.Join(p.NginxDir(), "mime.types") }

// NginxFastCGIParams returns the path to the Nginx fastcgi_params file used for PHP-FPM proxying.
func (p *Paths) NginxFastCGIParams() string { return filepath.Join(p.NginxDir(), "fastcgi_params") }

// DefaultSiteDir returns the directory served by the default Nginx virtual host.
func (p *Paths) DefaultSiteDir() string { return filepath.Join(p.Home, "default-site") }

// DnsmasqBin returns the path to the managed dnsmasq binary (~/.lamboserver/bin/dnsmasq).
func (p *Paths) DnsmasqBin() string { return filepath.Join(p.BinDir(), "dnsmasq") }

// DnsmasqConf returns the path to the dnsmasq configuration file.
func (p *Paths) DnsmasqConf() string { return filepath.Join(p.DnsmasqDir(), "dnsmasq.conf") }

// ResolverDir returns the macOS system resolver configuration directory (/etc/resolver).
func (p *Paths) ResolverDir() string { return "/etc/resolver" }

// ResolverFile returns the path to the resolver file for the .test TLD (/etc/resolver/test).
func (p *Paths) ResolverFile() string {
	return filepath.Join(p.ResolverDir(), TestTLD)
}

// CACert returns the path to the local CA certificate PEM file.
func (p *Paths) CACert() string { return filepath.Join(p.CertsDir(), "ca.pem") }

// CAKey returns the path to the local CA private key PEM file.
func (p *Paths) CAKey() string { return filepath.Join(p.CertsDir(), "ca-key.pem") }

// SiteCert returns the path to the SSL certificate PEM file for the given domain.
func (p *Paths) SiteCert(domain string) string {
	return filepath.Join(p.CertsDir(), domain+".pem")
}

// SiteKey returns the path to the SSL private key PEM file for the given domain.
func (p *Paths) SiteKey(domain string) string {
	return filepath.Join(p.CertsDir(), domain+"-key.pem")
}

// PhpVersionDir returns the installation directory for a specific PHP version.
func (p *Paths) PhpVersionDir(version string) string {
	return filepath.Join(p.PhpDir(), version)
}

// PhpCurrentLink returns the path to the "current" symlink pointing at the active PHP version.
func (p *Paths) PhpCurrentLink() string {
	return filepath.Join(p.PhpDir(), "current")
}

// PhpFpmSocket returns the Unix socket path used by PHP-FPM for the given PHP version.
func (p *Paths) PhpFpmSocket(version string) string {
	return filepath.Join(p.Home, "php", version, "var", "run", "php-fpm.sock")
}

// NodeVersionDir returns the installation directory for a specific Node.js version.
func (p *Paths) NodeVersionDir(version string) string {
	return filepath.Join(p.NodeDir(), version)
}

// NodeCurrentLink returns the path to the "current" symlink pointing at the active Node.js version.
func (p *Paths) NodeCurrentLink() string {
	return filepath.Join(p.NodeDir(), "current")
}

// MySQLDir returns the MySQL root directory (~/.lamboserver/mysql/).
func (p *Paths) MySQLDir() string { return filepath.Join(p.Home, "mysql") }

// MySQLDataDir returns the MySQL data directory (~/.lamboserver/mysql/data/).
// Note: This directory requires _mysql ownership and is created during MySQL install, not at app startup.
func (p *Paths) MySQLDataDir() string { return filepath.Join(p.MySQLDir(), "data") }

// MySQLConf returns the path to the MySQL configuration file (~/.lamboserver/mysql/my.cnf).
func (p *Paths) MySQLConf() string { return filepath.Join(p.MySQLDir(), "my.cnf") }

// MySQLSocket returns the path to the MySQL Unix socket (~/.lamboserver/mysql/mysql.sock).
func (p *Paths) MySQLSocket() string { return filepath.Join(p.MySQLDir(), "mysql.sock") }

// MySQLBinDir returns the directory containing MySQL binaries (~/.lamboserver/mysql/bin/).
func (p *Paths) MySQLBinDir() string { return filepath.Join(p.MySQLDir(), "bin") }

// PhpMyAdminDir returns the phpMyAdmin web files directory (~/.lamboserver/phpmyadmin/).
func (p *Paths) PhpMyAdminDir() string { return filepath.Join(p.Home, "phpmyadmin") }

// PhpMyAdminConf returns the path to the phpMyAdmin configuration file (~/.lamboserver/phpmyadmin/config.inc.php).
func (p *Paths) PhpMyAdminConf() string { return filepath.Join(p.PhpMyAdminDir(), "config.inc.php") }

// PostgreSQLDir returns the PostgreSQL root directory (~/.lamboserver/postgresql/).
func (p *Paths) PostgreSQLDir() string { return filepath.Join(p.Home, "postgresql") }

// PostgreSQLDataDir returns the PostgreSQL data directory (~/.lamboserver/postgresql/data/).
// Note: This directory requires 0700 permissions and is created by initdb during install, not at app startup.
func (p *Paths) PostgreSQLDataDir() string { return filepath.Join(p.PostgreSQLDir(), "data") }

// PostgreSQLBinDir returns the directory containing PostgreSQL binaries (~/.lamboserver/postgresql/bin/).
func (p *Paths) PostgreSQLBinDir() string { return filepath.Join(p.PostgreSQLDir(), "bin") }

// PostgreSQLSocketDir returns the Unix socket directory (~/.lamboserver/postgresql/run/).
func (p *Paths) PostgreSQLSocketDir() string { return filepath.Join(p.PostgreSQLDir(), "run") }

// PostgreSQLLogFile returns the path to the PostgreSQL log file (~/.lamboserver/postgresql/data/log/postgresql.log).
func (p *Paths) PostgreSQLLogFile() string {
	return filepath.Join(p.PostgreSQLDataDir(), "log", "postgresql.log")
}

// PostgreSQLConfFile returns the path to the PostgreSQL configuration file (~/.lamboserver/postgresql/data/postgresql.conf).
func (p *Paths) PostgreSQLConfFile() string {
	return filepath.Join(p.PostgreSQLDataDir(), "postgresql.conf")
}

// PostgreSQLHbaFile returns the path to the pg_hba.conf file (~/.lamboserver/postgresql/data/pg_hba.conf).
func (p *Paths) PostgreSQLHbaFile() string {
	return filepath.Join(p.PostgreSQLDataDir(), "pg_hba.conf")
}

// PgwebDir returns the pgweb binary directory (~/.lamboserver/pgweb/).
func (p *Paths) PgwebDir() string { return filepath.Join(p.Home, "pgweb") }

// LaunchDaemonsDir returns the system-level launchd plist directory (/Library/LaunchDaemons).
// Plists here run as root.
func (p *Paths) LaunchDaemonsDir() string {
	return "/Library/LaunchDaemons"
}

// LaunchAgentsDir returns the user-level launchd plist directory (~/Library/LaunchAgents).
func (p *Paths) LaunchAgentsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents")
}

// EnsureDirectories creates all required application directories if they do not exist.
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
		p.MySQLDir(),            // new: MySQL parent dir (data dir excluded — needs _mysql ownership)
		p.PhpMyAdminDir(),       // new: phpMyAdmin web files dir
		p.PostgreSQLDir(),       // new: PostgreSQL parent dir (data dir excluded — needs 0700 from initdb)
		p.PostgreSQLSocketDir(), // new: Unix socket directory
		p.PgwebDir(),            // pgweb binary dir
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
