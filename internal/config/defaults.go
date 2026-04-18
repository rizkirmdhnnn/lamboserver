package config

// Default port assignments for all managed services.
// These are used when initializing a fresh configuration and as the
// canonical reference for port allocation across the application.
const (
	// DefaultNginxPort is the default HTTP port for Nginx.
	DefaultNginxPort = 80

	// DefaultNginxSSLPort is the default HTTPS port for Nginx.
	DefaultNginxSSLPort = 443

	// DefaultDNSPort is the default port for the DNSMasq resolver.
	DefaultDNSPort = 53

	// DefaultPHPFPMBasePort is the base port for PHP-FPM pool sockets.
	// Each PHP version gets its own port: 8.1 → 9081, 8.2 → 9082, 8.3 → 9083.
	DefaultPHPFPMBasePort = 9081

	// DefaultMySQLPort is the default port for MySQL.
	DefaultMySQLPort = 3306

	// DefaultPostgresPort is the default port for PostgreSQL.
	DefaultPostgresPort = 5432

	// DefaultPhpMyAdminPort is the port where phpMyAdmin is served via Nginx.
	DefaultPhpMyAdminPort = 8088

	// DefaultPgwebPort is the default HTTP port for the pgweb web admin tool.
	DefaultPgwebPort = 8081
)

// PHPFPMPort returns the FPM port for a given PHP minor version number.
// For example, minor version 1 (PHP 8.1) returns 9081, minor 3 (PHP 8.3) returns 9083.
func PHPFPMPort(minor int) int {
	return DefaultPHPFPMBasePort + minor - 1
}
