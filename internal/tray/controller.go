package tray

import "context"

// AppController is the narrow interface the tray needs from the application.
// Follows consumer-defined interface convention (D-02).
type AppController interface {
	// Context returns the Wails runtime context for Show/Hide/Quit calls.
	Context() context.Context
	// GetAllStatuses returns status strings keyed by service registry name.
	GetAllStatuses() map[string]string
	// StartService starts the named service.
	StartService(name string) error
	// StopService stops the named service.
	StopService(name string) error
	// RestartService restarts the named service.
	RestartService(name string) error

	// Phase 3: Quick Access Links
	// GetTraySites returns sorted (newest-first), capped site info for the tray menu (D-04, D-05).
	// Named GetTraySites (not GetSites) because App already has GetSites() []sites.Site for Wails IPC.
	GetTraySites() []SiteInfo
	// OpenSiteInBrowser opens the site URL in the default browser (D-03).
	OpenSiteInBrowser(domain string)
	// GetWebAdminItems returns install status and URL for phpMyAdmin and pgweb (D-07, D-08).
	GetWebAdminItems() []WebAdminItem
	// OpenWebAdminInBrowser opens the named web admin tool in the default browser (D-08).
	// Named differently from App.OpenWebAdmin (which returns error for Wails IPC).
	OpenWebAdminInBrowser(name string)
	// GetTotalSiteCount returns total site count for overflow computation (D-05).
	GetTotalSiteCount() int
}

// SiteInfo is the tray-facing site descriptor.
// Domain-only label keeps menu items compact (D-06 discretion).
type SiteInfo struct {
	Domain string
	URL    string // Pre-computed: "https://domain" or "http://domain"
	Label  string // Display string, e.g. "myapp.test"
}

// WebAdminItem is the tray-facing descriptor for a web admin tool.
type WebAdminItem struct {
	Name      string // Registry key: "phpmyadmin" or "pgweb"
	Label     string // Display: "Open phpMyAdmin" / "Open pgweb"
	Installed bool
	URL       string
}

// Controller manages the macOS system tray icon and menu.
type Controller struct {
	app      AppController
	iconData []byte
	version  string
}

// New creates a tray controller. Call Start() to display the icon.
func New(app AppController, iconData []byte, version string) *Controller {
	return &Controller{
		app:      app,
		iconData: iconData,
		version:  version,
	}
}
