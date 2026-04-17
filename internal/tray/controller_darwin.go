//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "tray_darwin.h"
*/
import "C"

import (
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// instance holds the singleton Controller for CGO callback routing.
var instance *Controller

// Start creates the NSStatusItem and displays the tray icon.
// Must be called after the Wails context is available (i.e., in startup()).
func (c *Controller) Start() {
	instance = c
	iconPtr := unsafe.Pointer(&c.iconData[0])
	cVersion := C.CString(c.version)
	defer C.free(unsafe.Pointer(cVersion))
	C.CreateTray(iconPtr, C.int(len(c.iconData)), cVersion)
}

// Destroy removes the NSStatusItem from the menu bar.
func (c *Controller) Destroy() {
	C.DestroyTray()
	instance = nil
}

//export onShowWindow
func onShowWindow() {
	if instance == nil || instance.app == nil {
		return
	}
	ctx := instance.app.Context()
	if ctx == nil {
		return
	}
	wailsRuntime.Show(ctx)
}

//export onQuit
func onQuit() {
	if instance == nil || instance.app == nil {
		return
	}
	ctx := instance.app.Context()
	if ctx == nil {
		return
	}
	wailsRuntime.Quit(ctx)
}
