package main

import (
	"fmt"

	"github.com/rizkirmdhnnn/lamboserver/internal/dns"
	"github.com/rizkirmdhnnn/lamboserver/internal/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/site"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Nginx Methods ---

func (a *App) StartNginx() error {
	a.Debug.Info("StartNginx called")
	err := a.Nginx.Start()
	a.Debug.Action("StartNginx", err)
	return err
}

func (a *App) StopNginx() error {
	a.Debug.Info("StopNginx called")
	err := a.Nginx.Stop()
	a.Debug.Action("StopNginx", err)
	return err
}

func (a *App) ReloadNginx() error {
	a.Debug.Info("ReloadNginx called")
	err := a.Nginx.Reload()
	a.Debug.Action("ReloadNginx", err)
	return err
}

func (a *App) GetNginxStatus() nginx.ServiceStatus {
	status := a.Nginx.Status()
	a.Debug.Info("GetNginxStatus: installed=%v, running=%v", status.Installed, status.Running)
	return status
}

// --- DNS Methods ---

func (a *App) StartDns() error {
	a.Debug.Info("StartDns called")
	err := a.Dns.Start()
	a.Debug.Action("StartDns", err)
	return err
}

func (a *App) StopDns() error {
	a.Debug.Info("StopDns called")
	err := a.Dns.Stop()
	a.Debug.Action("StopDns", err)
	return err
}

func (a *App) GetDnsStatus() dns.ServiceStatus {
	status := a.Dns.Status()
	a.Debug.Info("GetDnsStatus: installed=%v, running=%v, resolver=%v",
		status.Installed, status.Running, status.Resolver)
	return status
}

// --- SSL Methods ---

func (a *App) IsCAInstalled() bool {
	installed := a.Certs.IsCAInstalled()
	a.Debug.Info("IsCAInstalled: %v", installed)
	return installed
}

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

// --- Dialog Methods ---

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
}

// --- Site Methods ---

func (a *App) LinkSite(path, domain string) error {
	a.Debug.Info("LinkSite called: path=%s, domain=%s", path, domain)
	if err := a.Sites.Link(path, domain); err != nil {
		a.Debug.Action(fmt.Sprintf("LinkSite(%s, %s)", path, domain), err)
		return err
	}
	if a.Nginx.Status().Running {
		err := a.Nginx.Reload()
		a.Debug.Action(fmt.Sprintf("LinkSite(%s, %s):reload", path, domain), err)
		return err
	}
	a.Debug.Action(fmt.Sprintf("LinkSite(%s, %s)", path, domain), nil)
	return nil
}

func (a *App) UnlinkSite(domain string) error {
	a.Debug.Info("UnlinkSite called: domain=%s", domain)
	if err := a.Sites.Unlink(domain); err != nil {
		a.Debug.Action(fmt.Sprintf("UnlinkSite(%s)", domain), err)
		return err
	}
	if a.Nginx.Status().Running {
		err := a.Nginx.Reload()
		a.Debug.Action(fmt.Sprintf("UnlinkSite(%s):reload", domain), err)
		return err
	}
	a.Debug.Action(fmt.Sprintf("UnlinkSite(%s)", domain), nil)
	return nil
}

func (a *App) GetSites() []site.Site {
	siteList := a.Sites.List()
	a.Debug.Info("GetSites: %d sites", len(siteList))
	return siteList
}
