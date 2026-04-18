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
