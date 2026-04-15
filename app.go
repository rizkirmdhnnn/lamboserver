package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/debug"
	"github.com/rizkirmdhnnn/lamboserver/internal/dns"
	applog "github.com/rizkirmdhnnn/lamboserver/internal/log"
	"github.com/rizkirmdhnnn/lamboserver/internal/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/node"
	"github.com/rizkirmdhnnn/lamboserver/internal/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/site"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// App struct holds all service managers
type App struct {
	ctx context.Context

	Paths   *system.Paths
	Config  *config.Store
	Launchd *system.LaunchdManager
	Php     *php.Manager
	PhpFpm  *php.FpmManager
	Nginx   *nginx.Manager
	Dns     *dns.Manager
	Certs   *cert.Manager
	Sites   *site.Manager
	Node    *node.Manager
	Logs    *applog.Reader
	Debug   *debug.Logger
	Shell   *system.Integration
}

// NewApp creates a new App with all managers initialized
func NewApp() *App {
	paths := system.NewPaths()
	store := config.NewStore(paths.ConfigFile())
	launchd := system.NewLaunchdManager(paths)
	certMgr := cert.NewManager(paths)

	return &App{
		Paths:   paths,
		Config:  store,
		Launchd: launchd,
		Php:     php.NewManager(paths, store),
		PhpFpm:  php.NewFpmManager(paths, store, launchd),
		Nginx:   nginx.NewManager(paths, launchd),
		Dns:     dns.NewManager(paths, launchd),
		Certs:   certMgr,
		Sites:   site.NewManager(paths, store, certMgr),
		Node:    node.NewManager(paths, store),
		Logs:    applog.NewReader(paths),
		Debug:   debug.NewLogger(paths.LogsDir()),
		Shell:   system.NewIntegration(paths),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Paths.EnsureDirectories()

	if a.Config.Get().DebugMode {
		a.Debug.Enable()
	}
	a.Debug.Info("LamboServer started")

	// Install privileged helper (asks admin password ONCE, ever)
	if !a.Launchd.Helper().IsInstalled() {
		a.Debug.Info("First-time setup: installing privileged helper")
		if err := a.Launchd.Helper().Install(); err != nil {
			a.Debug.Error("Failed to install helper: %v", err)
		} else {
			a.Debug.Info("Privileged helper installed - no more password prompts")
		}
	} else {
		// Always update helper script to pick up code changes
		a.Launchd.Helper().UpdateScript()
	}

	// Setup CA if not exists (for SSL certs)
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

	a.restoreSymlinks()
	a.cleanupStaleAgents()
	a.ensureServicesRunning()
}

// cleanupStaleAgents removes leftover user-level LaunchAgents for services
// that should run as system-level LaunchDaemons.
func (a *App) cleanupStaleAgents() {
	daemonLabels := []string{
		nginx.ServiceLabel,
		dns.ServiceLabel,
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

// shutdown is called when the app is closing.
// Uses direct process kill for speed and reliability instead of launchd uninstall.
func (a *App) shutdown(ctx context.Context) {
	a.Debug.Info("LamboServer shutting down, stopping services...")

	// Kill processes directly - fast and doesn't need admin
	// Also uninstall daemons so KeepAlive doesn't restart them
	a.PhpFpm.Stop()
	a.Nginx.Stop()
	a.Dns.Stop()

	// Also kill by process name as fallback
	system.RunCommand("pkill", "-f", "php-fpm.*lamboserver")
	system.RunCommand("pkill", "-f", "nginx.*lamboserver")
	system.RunCommand("pkill", "-f", "dnsmasq.*lamboserver")

	a.Debug.Info("LamboServer shutdown complete")
}

// ensureServicesRunning starts all services if not already running
func (a *App) ensureServicesRunning() {
	if !a.Nginx.Status().Running {
		a.Debug.Info("Nginx not running, starting...")
		if err := a.Nginx.Start(); err != nil {
			a.Debug.Error("Failed to start nginx: %v", err)
		} else {
			a.Debug.Info("Nginx started")
		}
	}

	if !a.Dns.Status().Running {
		a.Debug.Info("DNS not running, starting...")
		if err := a.Dns.Start(); err != nil {
			a.Debug.Error("Failed to start dnsmasq: %v", err)
		} else {
			a.Debug.Info("Dnsmasq started")
		}
	}

	// Start PHP-FPM if a PHP version is active
	if a.Config.Get().ActivePhpVersion != "" && !a.PhpFpm.Status().Running {
		a.Debug.Info("PHP-FPM not running, starting...")
		if err := a.PhpFpm.Start(); err != nil {
			a.Debug.Error("Failed to start php-fpm: %v", err)
		} else {
			a.Debug.Info("PHP-FPM started")
		}
	}
}

// restoreSymlinks re-creates symlinks for the currently active PHP and Node versions
func (a *App) restoreSymlinks() {
	cfg := a.Config.Get()

	if cfg.ActivePhpVersion != "" {
		versions, err := a.Php.ListInstalled()
		if err == nil {
			for _, v := range versions {
				if v.Version == cfg.ActivePhpVersion {
					a.Shell.LinkPhpVersion(v.Binary, v.FpmBin, v.Path)
					a.Debug.Info("Restored PHP symlinks for %s from %s", v.Version, v.Binary)
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
