package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/pkg/browser"
	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/services"
	"github.com/rizkirmdhnnn/lamboserver/internal/tray"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/dnsmasq"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/mysql"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/nodejs"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/pgweb"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/phpmyadmin"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/postgres"
	"github.com/rizkirmdhnnn/lamboserver/internal/sites"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/rizkirmdhnnn/lamboserver/pkg/logger"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the composition root that holds all service managers and coordinates
// the application lifecycle. Public methods on App are automatically exposed to
// the React frontend via Wails IPC bindings.
//
// app.go is the glue layer between Go and the frontend. Business logic lives in
// internal/ packages — methods here only validate input, delegate to a manager,
// and return results.
type App struct {
	ctx context.Context

	Paths      *system.Paths
	Config     *config.Store
	Launchd    *system.LaunchdManager
	Manager    *services.Manager
	Php        *php.Manager
	PhpFpm     *php.FpmManager
	Nginx      *nginx.Manager
	Dns        *dnsmasq.Manager
	Certs      *cert.Manager
	Sites      *sites.Manager
	Node       *nodejs.Manager
	MySQL      *mysql.Manager
	PostgreSQL *postgres.Manager
	PhpMyAdmin *phpmyadmin.Manager
	Pgweb      *pgweb.Manager
	Logs       *logger.Reader
	Debug      *logger.Logger
	Shell      *system.Integration
	Tray       *tray.Controller
}

// NewApp creates a new App instance, wiring all service managers with their
// real filesystem, command runner, and admin runner implementations. This is
// the composition root — all dependency injection happens here.
func NewApp() *App {
	paths := system.NewPaths()
	store := config.NewStore(paths.ConfigFile())
	launchd := system.NewLaunchdManager(paths)
	fs := system.RealFS{}
	cmd := system.RealCmdRunner{}
	admin := system.RealAdminRunner{}
	certMgr := cert.NewManager(paths, fs, admin)

	nginxMgr := nginx.NewManager(paths, launchd, fs, launchd.Helper())
	phpMgr := php.NewManager(paths, store, fs, cmd)
	fpmMgr := php.NewFpmManager(paths, store, launchd, fs)
	nodeMgr := nodejs.NewManager(paths, store, fs, cmd)
	mysqlMgr := mysql.NewManager(paths, launchd, fs, cmd, launchd.Helper(), admin)
	postgresMgr := postgres.NewManager(paths, fs, cmd, launchd.Helper(), admin)
	phpMyAdminMgr := phpmyadmin.NewManager(paths, certMgr, nginxMgr, fs, cmd)
	pgwebMgr := pgweb.NewManager(paths, pgweb.OsFileSystem{}, pgweb.SystemCommandRunner{})
	dnsMgr := dnsmasq.NewManager(paths, launchd, fs, admin)

	// Build the unified service manager and register all services.
	mgr := services.NewManager()
	mgr.Register("nginx", nginx.NewServiceAdapter(nginxMgr))
	mgr.Register("dnsmasq", dnsmasq.NewServiceAdapter(dnsMgr))
	mgr.Register("php", php.NewServiceAdapter(phpMgr, fpmMgr))
	mgr.Register("nodejs", nodejs.NewServiceAdapter(nodeMgr))
	mgr.Register("mysql", mysql.NewServiceAdapter(mysqlMgr))
	mgr.Register("postgresql", postgres.NewServiceAdapter(postgresMgr))
	mgr.RegisterWebAdmin("phpmyadmin", phpmyadmin.NewServiceAdapter(phpMyAdminMgr))

	return &App{
		Paths:      paths,
		Config:     store,
		Launchd:    launchd,
		Manager:    mgr,
		Php:        phpMgr,
		PhpFpm:     fpmMgr,
		Nginx:      nginxMgr,
		Dns:        dnsMgr,
		Certs:      certMgr,
		Sites:      sites.NewManager(paths, store, certMgr, fs),
		Node:       nodeMgr,
		MySQL:      mysqlMgr,
		PostgreSQL: postgresMgr,
		PhpMyAdmin: phpMyAdminMgr,
		Pgweb:      pgwebMgr,
		Logs:       logger.NewReader(paths),
		Debug:      logger.NewLogger(paths.LogsDir()),
		Shell:      system.NewIntegration(paths),
	}
}

// ── Lifecycle ─────────────────────────────────────────────────────────

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Paths.EnsureDirectories()

	if a.Config.Get().DebugMode {
		a.Debug.Enable()
	}
	a.Debug.Info("LamboServer started")

	if !a.Launchd.Helper().IsInstalled() {
		a.Debug.Info("First-time setup: installing privileged helper")
		if err := a.Launchd.Helper().Install(); err != nil {
			a.Debug.Error("Failed to install helper: %v", err)
		} else {
			a.Debug.Info("Privileged helper installed - no more password prompts")
		}
	} else {
		a.Launchd.Helper().UpdateScript()
	}

	if !a.Certs.IsCAInstalled() {
		a.Debug.Info("Setting up local CA for SSL...")
		if err := a.Certs.SetupCA(); err != nil {
			a.Debug.Error("Failed to setup CA: %v", err)
		} else {
			if err := a.Certs.TrustCA(); err != nil {
				a.Debug.Error("Failed to trust CA: %v", err)
			} else {
				a.Debug.Info("Local CA installed and trusted")
			}
		}
	}

	if !a.Shell.IsInstalled() {
		a.Debug.Info("Shell integration not found, installing PATH to shell rc files")
		if err := a.Shell.Install(); err != nil {
			a.Debug.Error("Failed to install shell integration: %v", err)
		}
	}

	a.Tray = tray.New(a, tray.Icon, "1.0.0")
	a.Tray.Start()
	a.Debug.Info("System tray initialized")

	a.restoreSymlinks()
	a.cleanupStaleAgents()
	a.ensureServicesRunning()
}

// Context returns the Wails runtime context for tray callbacks.
func (a *App) Context() context.Context {
	return a.ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.Debug.Info("LamboServer shutting down, stopping services...")
	if a.Tray != nil {
		a.Tray.Destroy()
	}
	a.MySQL.Stop()
	a.Pgweb.Stop()       // D-08: stop pgweb before PostgreSQL (pgweb depends on PG)
	a.PostgreSQL.Stop()
	a.PhpFpm.Stop()
	a.Nginx.Stop()
	a.Dns.Stop()
	system.RunCommand("pkill", "-f", "php-fpm.*lamboserver")
	system.RunCommand("pkill", "-f", "nginx.*lamboserver")
	system.RunCommand("pkill", "-f", "dnsmasq.*lamboserver")
	a.Debug.Info("LamboServer shutdown complete")
}

func (a *App) cleanupStaleAgents() {
	daemonLabels := []string{
		nginx.ServiceLabel,
		dnsmasq.ServiceLabel,
		mysql.ServiceLabel,
	}
	for _, label := range daemonLabels {
		agentPlist := a.Paths.LaunchAgentsDir() + "/" + label + ".plist"
		if _, err := os.Stat(agentPlist); err == nil {
			a.Debug.Info("Removing stale user agent: %s", label)
			system.RunCommand("launchctl", "bootout", fmt.Sprintf("gui/%d/%s", os.Getuid(), label))
			os.Remove(agentPlist)
		}
	}
}

func (a *App) ensureServicesRunning() {
	if !a.Nginx.Status().Running {
		a.Debug.Info("Nginx not running, starting...")
		if err := a.Nginx.Start(); err != nil {
			a.Debug.Error("Failed to start nginx: %v", err)
		}
	}
	if !a.Dns.Status().Running {
		a.Debug.Info("DNS not running, starting...")
		if err := a.Dns.Start(); err != nil {
			a.Debug.Error("Failed to start dnsmasq: %v", err)
		}
	}
	if a.Config.Get().MySQLEnabled && a.MySQL.IsInstalled() && !a.MySQL.Status().Running {
		a.Debug.Info("MySQL not running, starting...")
		if err := a.MySQL.Start(); err != nil {
			a.Debug.Error("Failed to start MySQL: %v", err)
		}
	}
	if a.Config.Get().PostgreSQLEnabled && a.PostgreSQL.IsInstalled() && !a.PostgreSQL.Status().Running {
		a.Debug.Info("PostgreSQL not running, starting...")
		if err := a.PostgreSQL.Start(); err != nil {
			a.Debug.Error("Failed to start PostgreSQL: %v", err)
		}
	}
	if a.Config.Get().ActivePhpVersion != "" && !a.PhpFpm.Status().Running {
		a.Debug.Info("PHP-FPM not running, starting...")
		if err := a.PhpFpm.Start(); err != nil {
			a.Debug.Error("Failed to start php-fpm: %v", err)
		}
	}
}

func (a *App) restoreSymlinks() {
	cfg := a.Config.Get()
	if cfg.ActivePhpVersion != "" {
		versions, err := a.Php.ListInstalled()
		if err == nil {
			for _, v := range versions {
				if v.Version == cfg.ActivePhpVersion {
					a.Shell.LinkPhpVersion(v.Binary, v.FpmBin, v.Path)
					a.Debug.Info("Restored PHP symlinks for %s", v.Version)
					break
				}
			}
		}
	}
	if cfg.ActiveNodeVersion != "" {
		nodeDir := a.Paths.NodeVersionDir(cfg.ActiveNodeVersion)
		a.Shell.LinkNodeBinaries(nodeDir)
		a.Debug.Info("Restored Node symlinks for %s", cfg.ActiveNodeVersion)
	}
}

// ── Generic Service Methods ───────────────────────────────────────────

// StartService starts the named service.
func (a *App) StartService(name string) error {
	a.Debug.Info("StartService called: %s", name)
	svc, err := a.Manager.Get(name)
	if err != nil {
		return err
	}
	err = svc.Start()
	a.Debug.Action(fmt.Sprintf("StartService(%s)", name), err)
	return err
}

// StopService stops the named service.
func (a *App) StopService(name string) error {
	a.Debug.Info("StopService called: %s", name)
	svc, err := a.Manager.Get(name)
	if err != nil {
		return err
	}
	err = svc.Stop()
	// D-08: pgweb must stop when PostgreSQL stops (pgweb depends on PG).
	if err == nil && name == "postgresql" {
		_ = a.Pgweb.Stop()
	}
	a.Debug.Action(fmt.Sprintf("StopService(%s)", name), err)
	return err
}

// RestartService restarts the named service.
func (a *App) RestartService(name string) error {
	a.Debug.Info("RestartService called: %s", name)
	svc, err := a.Manager.Get(name)
	if err != nil {
		return err
	}
	err = svc.Restart()
	a.Debug.Action(fmt.Sprintf("RestartService(%s)", name), err)
	return err
}

// GetAllStatuses returns the status of every registered service.
func (a *App) GetAllStatuses() map[string]string {
	result := make(map[string]string)
	for name, svc := range a.Manager.All() {
		result[name] = string(svc.Status())
	}
	for name, wa := range a.Manager.AllWebAdmin() {
		if wa.IsInstalled() {
			result[name] = "installed"
		} else {
			result[name] = "not_installed"
		}
	}
	return result
}

// GetServiceLogs returns the last n log lines for the named service.
func (a *App) GetServiceLogs(name string, lines int) []string {
	svc, err := a.Manager.Get(name)
	if err != nil {
		return nil
	}
	logs, err := svc.Logs()
	if err != nil {
		return nil
	}
	if len(logs) > lines {
		logs = logs[len(logs)-lines:]
	}
	return logs
}

// ── Version Management ────────────────────────────────────────────────

// GetPhpVersions returns all installed PHP versions with full metadata.
func (a *App) GetPhpVersions() ([]php.PhpVersion, error) {
	return a.Php.ListInstalled()
}

// GetAvailablePhpVersions returns PHP versions available for download from herdphp.com.
func (a *App) GetAvailablePhpVersions() ([]php.AvailablePhpVersion, error) {
	return a.Php.ListAvailable()
}

// GetNodeVersions returns all installed Node.js versions with full metadata.
func (a *App) GetNodeVersions() ([]nodejs.NodeVersion, error) {
	return a.Node.ListInstalled()
}

// GetAvailableNodeVersions returns Node.js LTS versions available for download.
func (a *App) GetAvailableNodeVersions() ([]string, error) {
	return a.Node.ListAvailable()
}

// GetAvailableVersions returns versions available for download.
func (a *App) GetAvailableVersions(service string) []string {
	switch service {
	case "php":
		versions, err := a.Php.ListAvailable()
		if err != nil {
			return nil
		}
		var result []string
		for _, v := range versions {
			result = append(result, v.Version)
		}
		return result
	case "nodejs":
		versions, err := a.Node.ListAvailable()
		if err != nil {
			return nil
		}
		return versions
	default:
		return nil
	}
}

// GetInstalledVersions returns the installed versions for a versioned service.
func (a *App) GetInstalledVersions(service string) []string {
	svc, err := a.Manager.GetVersioned(service)
	if err != nil {
		return nil
	}
	return svc.InstalledVersions()
}

// InstallVersion installs a specific version for a service. Works with both
// versioned services (php, nodejs) and regular services (mysql, postgresql).
func (a *App) InstallVersion(service, version string) error {
	a.Debug.Info("InstallVersion called: %s %s", service, version)
	// Try versioned service first.
	if svc, err := a.Manager.GetVersioned(service); err == nil {
		err = svc.Install(version)
		a.Debug.Action(fmt.Sprintf("InstallVersion(%s, %s)", service, version), err)
		return err
	}
	// Fall back to regular service.
	svc, err := a.Manager.Get(service)
	if err != nil {
		return err
	}
	err = svc.Install(version)
	a.Debug.Action(fmt.Sprintf("InstallVersion(%s, %s)", service, version), err)
	return err
}

// SwitchVersion switches the active version for a versioned service.
func (a *App) SwitchVersion(service, version string) error {
	a.Debug.Info("SwitchVersion called: %s -> %s", service, version)
	svc, err := a.Manager.GetVersioned(service)
	if err != nil {
		return err
	}
	err = svc.SwitchVersion(version)
	a.Debug.Action(fmt.Sprintf("SwitchVersion(%s, %s)", service, version), err)
	return err
}

// UninstallVersion removes an installed version for a versioned service.
func (a *App) UninstallVersion(service, version string) error {
	a.Debug.Info("UninstallVersion called: %s %s", service, version)
	var err error
	switch service {
	case "php":
		err = a.Php.Uninstall(version)
	case "nodejs":
		err = a.Node.Uninstall(version)
	default:
		err = fmt.Errorf("uninstall not supported for service %q", service)
	}
	a.Debug.Action(fmt.Sprintf("UninstallVersion(%s, %s)", service, version), err)
	return err
}

// ── Web Admin ─────────────────────────────────────────────────────────

// GetWebAdminURL returns the URL for the named web admin service.
func (a *App) GetWebAdminURL(name string) string {
	if svc, err := a.Manager.GetWebAdmin(name); err == nil {
		return svc.URL()
	}
	return ""
}

// OpenWebAdmin opens the named web admin service in the default browser.
func (a *App) OpenWebAdmin(name string) error {
	a.Debug.Info("OpenWebAdmin called: %s", name)

	// Check WebAdminService registry first.
	if svc, err := a.Manager.GetWebAdmin(name); err == nil {
		url := svc.URL()
		if url == "" {
			return fmt.Errorf("web admin %q has no URL", name)
		}
		return browser.OpenURL(url)
	}

	if name == "pgweb" {
		a.Debug.Info("OpenWebAdmin: pgweb via direct manager")
		return browser.OpenURL(a.Pgweb.URL())
	}

	return fmt.Errorf("web admin %q not found", name)
}

// InstallWebAdmin installs the named web admin service.
func (a *App) InstallWebAdmin(name, version string) error {
	a.Debug.Info("InstallWebAdmin called: %s %s", name, version)
	svc, err := a.Manager.GetWebAdmin(name)
	if err != nil {
		return err
	}
	err = svc.Install(version)
	a.Debug.Action(fmt.Sprintf("InstallWebAdmin(%s, %s)", name, version), err)
	return err
}

// UninstallWebAdmin uninstalls the named web admin service.
func (a *App) UninstallWebAdmin(name string) error {
	a.Debug.Info("UninstallWebAdmin called: %s", name)
	var err error
	switch name {
	case "phpmyadmin":
		err = a.PhpMyAdmin.Uninstall()
	default:
		err = fmt.Errorf("uninstall not supported for web admin %q", name)
	}
	a.Debug.Action(fmt.Sprintf("UninstallWebAdmin(%s)", name), err)
	return err
}

// ── pgweb ─────────────────────────────────────────────────────────────

// PgwebStatus is the frontend-facing status snapshot for pgweb.
type PgwebStatus struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Port      int  `json:"port"`
}

// GetPgwebStatus returns the current pgweb status for the frontend.
func (a *App) GetPgwebStatus() PgwebStatus {
	s := a.Pgweb.Status()
	return PgwebStatus{Installed: s.Installed, Running: s.Running, Port: s.Port}
}

// InstallPgweb downloads the pgweb binary from GitHub releases.
func (a *App) InstallPgweb() error {
	a.Debug.Info("InstallPgweb called")
	err := a.Pgweb.Install()
	a.Debug.Action("InstallPgweb", err)
	return err
}

// StartPgweb starts the pgweb HTTP daemon on port 8081.
func (a *App) StartPgweb() error {
	a.Debug.Info("StartPgweb called")
	err := a.Pgweb.Start()
	a.Debug.Action("StartPgweb", err)
	return err
}

// StopPgweb stops the running pgweb process.
func (a *App) StopPgweb() error {
	a.Debug.Info("StopPgweb called")
	err := a.Pgweb.Stop()
	a.Debug.Action("StopPgweb", err)
	return err
}

// OpenPgweb opens pgweb in the default browser.
func (a *App) OpenPgweb() error {
	a.Debug.Info("OpenPgweb called")
	return browser.OpenURL("http://127.0.0.1:8081")
}

// ── Database CRUD ─────────────────────────────────────────────────────

// InitService runs first-time initialization for services that need it (MySQL, PostgreSQL).
func (a *App) InitService(name string) error {
	a.Debug.Info("InitService called: %s", name)
	var err error
	switch name {
	case "mysql":
		err = a.MySQL.InitDataDir()
	case "postgresql":
		err = a.PostgreSQL.InitDataDir()
	default:
		err = fmt.Errorf("init not supported for service %q", name)
	}
	a.Debug.Action(fmt.Sprintf("InitService(%s)", name), err)
	return err
}

// DatabaseEntry represents a single database with optional size info.
type DatabaseEntry struct {
	Name string `json:"name"`
	Size string `json:"size,omitempty"`
}

// ListServiceDatabases returns databases for the named database service.
func (a *App) ListServiceDatabases(service string) ([]DatabaseEntry, error) {
	a.Debug.Info("ListServiceDatabases called: %s", service)
	switch service {
	case "mysql":
		names, err := a.MySQL.ListDatabases()
		if err != nil {
			return nil, err
		}
		entries := make([]DatabaseEntry, len(names))
		for i, n := range names {
			entries[i] = DatabaseEntry{Name: n}
		}
		return entries, nil
	case "postgresql":
		dbs, err := a.PostgreSQL.ListDatabases()
		if err != nil {
			return nil, err
		}
		entries := make([]DatabaseEntry, len(dbs))
		for i, d := range dbs {
			entries[i] = DatabaseEntry{Name: d.Name, Size: d.Size}
		}
		return entries, nil
	default:
		return nil, fmt.Errorf("database listing not supported for service %q", service)
	}
}

// CreateServiceDatabase creates a database on the named database service.
func (a *App) CreateServiceDatabase(service, name string) error {
	a.Debug.Info("CreateServiceDatabase called: %s %s", service, name)
	var err error
	switch service {
	case "mysql":
		err = a.MySQL.CreateDatabase(name)
	case "postgresql":
		err = a.PostgreSQL.CreateDatabase(name)
	default:
		err = fmt.Errorf("database creation not supported for service %q", service)
	}
	a.Debug.Action(fmt.Sprintf("CreateServiceDatabase(%s, %s)", service, name), err)
	return err
}

// DropServiceDatabase drops a database on the named database service.
func (a *App) DropServiceDatabase(service, name string) error {
	a.Debug.Info("DropServiceDatabase called: %s %s", service, name)
	var err error
	switch service {
	case "mysql":
		err = a.MySQL.DropDatabase(name)
	case "postgresql":
		err = a.PostgreSQL.DropDatabase(name)
	default:
		err = fmt.Errorf("database drop not supported for service %q", service)
	}
	a.Debug.Action(fmt.Sprintf("DropServiceDatabase(%s, %s)", service, name), err)
	return err
}

// ── Logs ──────────────────────────────────────────────────────────────

// GetLogFiles returns metadata for all .log files found in the app log directories.
func (a *App) GetLogFiles() []logger.LogFile {
	return a.Logs.ListLogFiles()
}

// ReadLog returns the last n lines from the specified log file.
func (a *App) ReadLog(filePath string, lines int) ([]string, error) {
	return a.Logs.ReadLastN(filePath, lines)
}

// ── Dashboard ─────────────────────────────────────────────────────────

// DashboardStatus aggregates all service states for the frontend dashboard.
type DashboardStatus struct {
	NginxStatus      nginx.ServiceStatus    `json:"nginx_status"`
	DnsStatus        dnsmasq.ServiceStatus  `json:"dns_status"`
	FpmStatus        php.FpmStatus          `json:"fpm_status"`
	MySQLStatus      mysql.ServiceStatus    `json:"mysql_status"`
	PostgreSQLStatus postgres.ServiceStatus `json:"postgresql_status"`
	PhpVersions      int                    `json:"php_versions"`
	NodeVersions     int                    `json:"node_versions"`
	SitesCount       int                    `json:"sites_count"`
	ActivePhp        string                 `json:"active_php"`
	ActiveNode       string                 `json:"active_node"`
	FirstRun         bool                   `json:"first_run"`
	DebugMode        bool                   `json:"debug_mode"`
	ShellIntegrated  bool                   `json:"shell_integrated"`
}

// GetDashboardStatus returns an aggregated snapshot of all services and config.
func (a *App) GetDashboardStatus() DashboardStatus {
	cfg := a.Config.Get()
	phpVersions, _ := a.Php.ListInstalled()
	nodeVersions, _ := a.Node.ListInstalled()
	return DashboardStatus{
		NginxStatus:      a.Nginx.Status(),
		DnsStatus:        a.Dns.Status(),
		FpmStatus:        a.PhpFpm.Status(),
		MySQLStatus:      a.MySQL.Status(),
		PostgreSQLStatus: a.PostgreSQL.Status(),
		PhpVersions:      len(phpVersions),
		NodeVersions:     len(nodeVersions),
		SitesCount:       len(cfg.Sites),
		ActivePhp:        cfg.ActivePhpVersion,
		ActiveNode:       cfg.ActiveNodeVersion,
		FirstRun:         !cfg.FirstRunComplete,
		DebugMode:        cfg.DebugMode,
		ShellIntegrated:  a.Shell.IsInstalled(),
	}
}

// ── Sites ─────────────────────────────────────────────────────────────

// GetSites returns all configured site virtual hosts.
func (a *App) GetSites() []sites.Site {
	return a.Sites.List()
}

// GetTraySites returns sorted (newest-first), capped site info for the tray menu.
// Satisfies tray.AppController.GetTraySites (D-04: newest first, D-05: cap at 15).
func (a *App) GetTraySites() []tray.SiteInfo {
	raw := a.Sites.List()
	// D-04: Sort newest first. RFC3339 is lexicographically sortable.
	sort.Slice(raw, func(i, j int) bool {
		return raw[i].CreatedAt > raw[j].CreatedAt
	})
	const maxSites = 15
	if len(raw) > maxSites {
		raw = raw[:maxSites]
	}
	items := make([]tray.SiteInfo, len(raw))
	for i, s := range raw {
		scheme := "https"
		if !s.SSLEnabled {
			scheme = "http"
		}
		items[i] = tray.SiteInfo{
			Domain: s.Domain,
			URL:    scheme + "://" + s.Domain,
			Label:  s.Domain, // D-06 discretion: domain-only label
		}
	}
	return items
}

// OpenSiteInBrowser opens the given site domain in the default browser.
// Satisfies tray.AppController.OpenSiteInBrowser (D-03: https if SSL enabled).
func (a *App) OpenSiteInBrowser(domain string) {
	a.Debug.Info("OpenSiteInBrowser called: %s", domain)
	for _, s := range a.Sites.List() {
		if s.Domain == domain {
			scheme := "https"
			if !s.SSLEnabled {
				scheme = "http"
			}
			_ = browser.OpenURL(scheme + "://" + domain)
			return
		}
	}
	// Domain not found in config -- best-effort https.
	_ = browser.OpenURL("https://" + domain)
}

// GetWebAdminItems returns install status for phpMyAdmin (via registry) and pgweb
// (via direct Manager field -- pgweb is NOT registered as a WebAdminService).
// Satisfies tray.AppController.GetWebAdminItems (D-07, D-08).
func (a *App) GetWebAdminItems() []tray.WebAdminItem {
	var items []tray.WebAdminItem
	// phpMyAdmin via WebAdmin registry
	if pma, err := a.Manager.GetWebAdmin("phpmyadmin"); err == nil {
		items = append(items, tray.WebAdminItem{
			Name:      "phpmyadmin",
			Label:     "Open phpMyAdmin",
			Installed: pma.IsInstalled(),
			URL:       pma.URL(),
		})
	}
	// pgweb via direct Manager field (NOT in WebAdmin registry -- RESEARCH pitfall 1)
	items = append(items, tray.WebAdminItem{
		Name:      "pgweb",
		Label:     "Open pgweb",
		Installed: a.Pgweb.IsInstalled(),
		URL:       a.Pgweb.URL(),
	})
	return items
}

// GetTotalSiteCount returns the total number of configured sites.
// Used by tray to compute overflow count for the "(+N more)" label.
func (a *App) GetTotalSiteCount() int {
	return len(a.Sites.List())
}

// OpenWebAdminInBrowser opens the named web admin tool in the default browser.
// Fire-and-forget wrapper for tray.AppController (no error return).
func (a *App) OpenWebAdminInBrowser(name string) {
	a.Debug.Info("OpenWebAdminInBrowser called: %s", name)
	_ = a.OpenWebAdmin(name)
}

// CreateSite creates an Nginx virtual host and SSL certificate for the given
// domain and path, then reloads Nginx.
func (a *App) CreateSite(domain, path, phpVersion string) error {
	a.Debug.Info("CreateSite called: domain=%s, path=%s", domain, path)
	if err := a.Sites.Link(path, domain); err != nil {
		a.Debug.Action(fmt.Sprintf("CreateSite(%s)", domain), err)
		return err
	}
	if a.Nginx.Status().Running {
		if err := a.Nginx.Reload(); err != nil {
			return err
		}
	}
	a.Debug.Action(fmt.Sprintf("CreateSite(%s)", domain), nil)
	return nil
}

// DeleteSite removes the Nginx virtual host and SSL certificate for the given domain.
func (a *App) DeleteSite(domain string) error {
	a.Debug.Info("DeleteSite called: domain=%s", domain)
	if err := a.Sites.Unlink(domain); err != nil {
		a.Debug.Action(fmt.Sprintf("DeleteSite(%s)", domain), err)
		return err
	}
	if a.Nginx.Status().Running {
		if err := a.Nginx.Reload(); err != nil {
			return err
		}
	}
	a.Debug.Action(fmt.Sprintf("DeleteSite(%s)", domain), nil)
	return nil
}

// ── Config ────────────────────────────────────────────────────────────

// GetConfig returns the current application configuration.
func (a *App) GetConfig() config.AppConfig {
	return a.Config.Get()
}

// SaveConfig persists the given configuration.
func (a *App) SaveConfig(cfg config.AppConfig) error {
	return a.Config.Save()
}

// ── Certificate Authority ─────────────────────────────────────────────

// IsCAInstalled reports whether the local CA certificate files are present.
func (a *App) IsCAInstalled() bool {
	return a.Certs.IsCAInstalled()
}

// SetupCA generates the local CA and trusts it in the system keychain.
func (a *App) SetupCA() error {
	a.Debug.Info("SetupCA called")
	if err := a.Certs.SetupCA(); err != nil {
		a.Debug.Action("SetupCA:generate", err)
		return err
	}
	err := a.Certs.TrustCA()
	a.Debug.Action("SetupCA:trust", err)
	return err
}

// ── Debug & Shell Integration ─────────────────────────────────────────

// EnableDebug activates file-based debug logging.
func (a *App) EnableDebug() error {
	if err := a.Debug.Enable(); err != nil {
		return err
	}
	return a.Config.SetDebugMode(true)
}

// DisableDebug deactivates file-based debug logging.
func (a *App) DisableDebug() {
	a.Debug.Disable()
	a.Config.SetDebugMode(false)
}

// IsDebugEnabled reports whether debug logging is currently active.
func (a *App) IsDebugEnabled() bool {
	return a.Debug.IsEnabled()
}

// GetDebugLogPath returns the absolute path to the debug log file.
func (a *App) GetDebugLogPath() string {
	return a.Debug.LogPath()
}

// ClearDebugLog truncates the debug log file to zero bytes.
func (a *App) ClearDebugLog() error {
	return a.Debug.ClearLog()
}

// IsShellIntegrated reports whether PATH injection is present in shell rc files.
func (a *App) IsShellIntegrated() bool {
	return a.Shell.IsInstalled()
}

// InstallShellIntegration injects the bin PATH export into shell rc files.
func (a *App) InstallShellIntegration() error {
	a.Debug.Info("InstallShellIntegration called")
	err := a.Shell.Install()
	a.Debug.Action("InstallShellIntegration", err)
	return err
}

// UninstallShellIntegration removes the PATH injection from shell rc files.
func (a *App) UninstallShellIntegration() error {
	a.Debug.Info("UninstallShellIntegration called")
	err := a.Shell.Uninstall()
	a.Debug.Action("UninstallShellIntegration", err)
	return err
}

// ── Dialog ────────────────────────────────────────────────────────────

// SelectDirectory opens a native directory picker dialog.
func (a *App) SelectDirectory() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
}
