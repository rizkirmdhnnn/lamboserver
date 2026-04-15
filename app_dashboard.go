package main

import (
	"github.com/rizkirmdhnnn/lamboserver/internal/dns"
	"github.com/rizkirmdhnnn/lamboserver/internal/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/php"
)

type DashboardStatus struct {
	NginxStatus     nginx.ServiceStatus `json:"nginx_status"`
	DnsStatus       dns.ServiceStatus   `json:"dns_status"`
	FpmStatus       php.FpmStatus       `json:"fpm_status"`
	PhpVersions     int                 `json:"php_versions"`
	NodeVersions    int                 `json:"node_versions"`
	SitesCount      int                 `json:"sites_count"`
	ActivePhp       string              `json:"active_php"`
	ActiveNode      string              `json:"active_node"`
	FirstRun        bool                `json:"first_run"`
	DebugMode       bool                `json:"debug_mode"`
	ShellIntegrated bool                `json:"shell_integrated"`
}

func (a *App) GetDashboardStatus() DashboardStatus {
	a.Debug.Info("GetDashboardStatus called")
	cfg := a.Config.Get()

	phpVersions, _ := a.Php.ListInstalled()
	nodeVersions, _ := a.Node.ListInstalled()

	return DashboardStatus{
		NginxStatus:     a.Nginx.Status(),
		DnsStatus:       a.Dns.Status(),
		FpmStatus:       a.PhpFpm.Status(),
		PhpVersions:     len(phpVersions),
		NodeVersions:    len(nodeVersions),
		SitesCount:      len(cfg.Sites),
		ActivePhp:       cfg.ActivePhpVersion,
		ActiveNode:      cfg.ActiveNodeVersion,
		FirstRun:        !cfg.FirstRunComplete,
		DebugMode:       cfg.DebugMode,
		ShellIntegrated: a.Shell.IsInstalled(),
	}
}
