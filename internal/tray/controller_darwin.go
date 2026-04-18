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

//export onOpenSite
func onOpenSite(cDomain *C.char) {
	if instance == nil || instance.app == nil {
		return
	}
	domain := C.GoString(cDomain)
	go func() {
		instance.app.OpenSiteInBrowser(domain)
	}()
}

//export onOpenWebAdmin
func onOpenWebAdmin(cName *C.char) {
	if instance == nil || instance.app == nil {
		return
	}
	name := C.GoString(cName)
	go func() {
		instance.app.OpenWebAdminInBrowser(name)
	}()
}

//export onRefreshQuickAccess
func onRefreshQuickAccess() {
	if instance == nil || instance.app == nil {
		return
	}
	sites := instance.app.GetTraySites()
	C.BeginQuickAccessRebuild(C.int(len(sites)))
	for i, s := range sites {
		cDomain := C.CString(s.Domain)
		cURL := C.CString(s.URL)
		cLabel := C.CString(s.Label)
		C.AddQuickAccessSite(C.int(i), cDomain, cURL, cLabel)
		C.free(unsafe.Pointer(cDomain))
		C.free(unsafe.Pointer(cURL))
		C.free(unsafe.Pointer(cLabel))
	}
	overflow := instance.app.GetTotalSiteCount() - len(sites)
	if overflow < 0 {
		overflow = 0
	}
	C.SetQuickAccessOverflow(C.int(overflow))
	webAdmins := instance.app.GetWebAdminItems()
	for _, wa := range webAdmins {
		if !wa.Installed {
			continue
		}
		cName := C.CString(wa.Name)
		cLabel := C.CString(wa.Label)
		C.AddQuickAccessWebAdmin(cName, cLabel)
		C.free(unsafe.Pointer(cName))
		C.free(unsafe.Pointer(cLabel))
	}
	C.CommitQuickAccessRebuild()
}
