package system_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

func TestPaths_NewPaths_UsesHomeDir(t *testing.T) {
	p := system.NewPaths()
	assert.NotEmpty(t, p.Home)
	assert.True(t, strings.HasSuffix(p.Home, ".lamboserver"), "Home should end with .lamboserver, got: %s", p.Home)
}

func TestPaths_RootDirectories(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"BinDir", p.BinDir(), root + "/bin"},
		{"PhpDir", p.PhpDir(), root + "/php"},
		{"NodeDir", p.NodeDir(), root + "/nodejs"},
		{"NginxDir", p.NginxDir(), root + "/nginx"},
		{"CertsDir", p.CertsDir(), root + "/certs"},
		{"LogsDir", p.LogsDir(), root + "/logs"},
		{"ServicesDir", p.ServicesDir(), root + "/services"},
		{"DnsmasqDir", p.DnsmasqDir(), root + "/dnsmasq"},
		{"ConfigFile", p.ConfigFile(), root + "/config.json"},
		{"DefaultSiteDir", p.DefaultSiteDir(), root + "/default-site"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

func TestPaths_NginxPaths(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"NginxBin", p.NginxBin(), root + "/bin/nginx"},
		{"NginxConf", p.NginxConf(), root + "/nginx/nginx.conf"},
		{"NginxSitesDir", p.NginxSitesDir(), root + "/nginx/sites"},
		{"NginxLogsDir", p.NginxLogsDir(), root + "/nginx/logs"},
		{"NginxAccessLog", p.NginxAccessLog(), root + "/nginx/logs/access.log"},
		{"NginxErrorLog", p.NginxErrorLog(), root + "/nginx/logs/error.log"},
		{"NginxMimeTypes", p.NginxMimeTypes(), root + "/nginx/mime.types"},
		{"NginxFastCGIParams", p.NginxFastCGIParams(), root + "/nginx/fastcgi_params"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

func TestPaths_DnsmasqPaths(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	assert.Equal(t, root+"/bin/dnsmasq", p.DnsmasqBin())
	assert.Equal(t, root+"/dnsmasq/dnsmasq.conf", p.DnsmasqConf())
}

func TestPaths_CertPaths(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	assert.Equal(t, root+"/certs/ca.pem", p.CACert())
	assert.Equal(t, root+"/certs/ca-key.pem", p.CAKey())
	assert.Equal(t, root+"/certs/foo.test.pem", p.SiteCert("foo.test"))
	assert.Equal(t, root+"/certs/foo.test-key.pem", p.SiteKey("foo.test"))
}

func TestPaths_PhpPaths(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	assert.Equal(t, root+"/php/8.3", p.PhpVersionDir("8.3"))
	assert.Equal(t, root+"/php/current", p.PhpCurrentLink())
	assert.Equal(t, root+"/php/8.3/var/run/php-fpm.sock", p.PhpFpmSocket("8.3"))
}

func TestPaths_NodePaths(t *testing.T) {
	root := "/tmp/testroot"
	p := newTestPaths(root)

	assert.Equal(t, root+"/nodejs/20.0.0", p.NodeVersionDir("20.0.0"))
	assert.Equal(t, root+"/nodejs/current", p.NodeCurrentLink())
}

func TestPaths_SystemPaths_AreAbsolute(t *testing.T) {
	p := newTestPaths("/tmp/testroot")

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"ResolverDir", p.ResolverDir(), "/etc/resolver"},
		{"ResolverFile", p.ResolverFile(), "/etc/resolver/test"},
		{"LaunchDaemonsDir", p.LaunchDaemonsDir(), "/Library/LaunchDaemons"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got)
			assert.True(t, strings.HasPrefix(tt.got, "/"), "path should be absolute: %s", tt.got)
		})
	}
}

func TestPaths_LaunchAgentsDir(t *testing.T) {
	p := newTestPaths("/tmp/testroot")
	dir := p.LaunchAgentsDir()
	assert.Contains(t, dir, "Library/LaunchAgents")
}

func TestPaths_EnsureDirectories_CreatesAll(t *testing.T) {
	root := t.TempDir()
	p := newTestPaths(root)

	require.NoError(t, p.EnsureDirectories())

	expectedDirs := []struct {
		name string
		path string
	}{
		{"Home", p.Home},
		{"BinDir", p.BinDir()},
		{"PhpDir", p.PhpDir()},
		{"NodeDir", p.NodeDir()},
		{"NginxDir", p.NginxDir()},
		{"NginxSitesDir", p.NginxSitesDir()},
		{"NginxLogsDir", p.NginxLogsDir()},
		{"CertsDir", p.CertsDir()},
		{"LogsDir", p.LogsDir()},
		{"ServicesDir", p.ServicesDir()},
		{"DnsmasqDir", p.DnsmasqDir()},
		{"DefaultSiteDir", p.DefaultSiteDir()},
		{"MySQLDir", p.MySQLDir()},
		{"PhpMyAdminDir", p.PhpMyAdminDir()},
		{"PostgreSQLDir", p.PostgreSQLDir()},
		{"PostgreSQLSocketDir", p.PostgreSQLSocketDir()},
	}

	for _, d := range expectedDirs {
		t.Run(d.name, func(t *testing.T) {
			info, err := os.Stat(d.path)
			require.NoError(t, err, "directory should exist: %s", d.path)
			assert.True(t, info.IsDir(), "should be a directory: %s", d.path)
		})
	}

	// MySQLDataDir must NOT be created by EnsureDirectories (needs _mysql ownership)
	_, err := os.Stat(p.MySQLDataDir())
	assert.True(t, os.IsNotExist(err), "MySQLDataDir should NOT be created by EnsureDirectories")

	// PostgreSQLDataDir must NOT be created by EnsureDirectories (requires 0700 from initdb)
	_, errPg := os.Stat(p.PostgreSQLDataDir())
	assert.True(t, os.IsNotExist(errPg), "PostgreSQLDataDir should NOT be created by EnsureDirectories")
}

func TestPaths_MySQLPaths(t *testing.T) {
	p := &system.Paths{Home: "/test/home"}
	assert.Equal(t, "/test/home/mysql", p.MySQLDir())
	assert.Equal(t, "/test/home/mysql/data", p.MySQLDataDir())
	assert.Equal(t, "/test/home/mysql/my.cnf", p.MySQLConf())
	assert.Equal(t, "/test/home/mysql/mysql.sock", p.MySQLSocket())
	assert.Equal(t, "/test/home/mysql/bin", p.MySQLBinDir())
}

func TestPaths_PhpMyAdminPaths(t *testing.T) {
	p := &system.Paths{Home: "/test/home"}
	assert.Equal(t, "/test/home/phpmyadmin", p.PhpMyAdminDir())
	assert.Equal(t, "/test/home/phpmyadmin/config.inc.php", p.PhpMyAdminConf())
}

func TestPaths_PostgreSQLPaths(t *testing.T) {
	p := &system.Paths{Home: "/test/home"}
	assert.Equal(t, "/test/home/postgresql", p.PostgreSQLDir())
	assert.Equal(t, "/test/home/postgresql/data", p.PostgreSQLDataDir())
	assert.Equal(t, "/test/home/postgresql/bin", p.PostgreSQLBinDir())
	assert.Equal(t, "/test/home/postgresql/run", p.PostgreSQLSocketDir())
	assert.Equal(t, "/test/home/postgresql/data/log/postgresql.log", p.PostgreSQLLogFile())
	assert.Equal(t, "/test/home/postgresql/data/postgresql.conf", p.PostgreSQLConfFile())
	assert.Equal(t, "/test/home/postgresql/data/pg_hba.conf", p.PostgreSQLHbaFile())
}

func TestPaths_EnsureDirectories_Idempotent(t *testing.T) {
	root := t.TempDir()
	p := newTestPaths(root)

	require.NoError(t, p.EnsureDirectories(), "first call should succeed")
	require.NoError(t, p.EnsureDirectories(), "second call should succeed (idempotent)")
}
