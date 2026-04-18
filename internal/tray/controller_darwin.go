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

//export onServiceAction
func onServiceAction(cName *C.char, cAction *C.char) {
	if instance == nil || instance.app == nil {
		return
	}
	name := C.GoString(cName)
	action := C.GoString(cAction)
	// Fire-and-forget per D-06: no toast, no spinner.
	go func() {
		switch action {
		case "start":
			instance.app.StartService(name)
		case "stop":
			instance.app.StopService(name)
		case "restart":
			instance.app.RestartService(name)
		}
	}()
}

//export onRefreshStatuses
func onRefreshStatuses() {
	if instance == nil || instance.app == nil {
		return
	}
	statuses := instance.app.GetAllStatuses()
	// Fixed order per D-02: dnsmasq, nginx, php, mysql, postgresql
	serviceOrder := []string{"dnsmasq", "nginx", "php", "mysql", "postgresql"}
	for i, name := range serviceOrder {
		status, ok := statuses[name]
		if !ok {
			status = "not_installed"
		}
		cStatus := C.CString(status)
		C.UpdateServiceStatus(C.int(i), cStatus)
		C.free(unsafe.Pointer(cStatus))
	}
}
