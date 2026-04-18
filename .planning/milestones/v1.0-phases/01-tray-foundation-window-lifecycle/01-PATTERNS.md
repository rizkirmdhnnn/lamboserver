# Phase 01: Tray Foundation & Window Lifecycle - Pattern Map

**Mapped:** 2026-04-17
**Files analyzed:** 9 new/modified files
**Analogs found:** 8 / 9

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/tray/doc.go` | config | N/A | `internal/services/dnsmasq/doc.go` | exact |
| `internal/tray/controller.go` | controller | event-driven | `internal/services/pgweb/manager.go` | role-match |
| `internal/tray/controller_darwin.go` | controller | event-driven | None (CGO bridge) | no-analog |
| `internal/tray/tray_darwin.h` | config | N/A | None (ObjC header) | no-analog |
| `internal/tray/tray_darwin.m` | controller | event-driven | None (ObjC implementation) | no-analog |
| `internal/tray/tray_unsupported.go` | controller | N/A | N/A (stub) | no-analog |
| `internal/tray/icon.go` | utility | file-I/O | `main.go` (embed pattern) | role-match |
| `main.go` | config | request-response | Self (modify in place) | exact |
| `app.go` | controller | request-response | Self (modify in place) | exact |

## Pattern Assignments

### `internal/tray/doc.go` (package documentation)

**Analog:** `internal/services/dnsmasq/doc.go`

**Doc comment pattern** (lines 1-9):
```go
// Package dns manages dnsmasq for local DNS resolution in LamboServer.
//
// The Manager type handles dnsmasq installation detection, writes resolver
// configuration files under /etc/resolver/ for the local TLD, and controls
// the dnsmasq LaunchDaemon lifecycle (install, start, stop, restart). It uses
// FileSystem, LaunchdService, and AdminRunner interfaces for testability.
//
// This package does not manage external DNS records or upstream resolvers.
package dnsmasq
```

**Apply:** Follow same structure -- package purpose, what Controller type does, what it does NOT do. Replace existing placeholder in `internal/tray/tray.go` (currently just a doc comment + `package tray`). The new `doc.go` replaces `tray.go` as the package-level documentation file.

---

### `internal/tray/controller.go` (controller, event-driven)

**Analog:** `internal/services/pgweb/manager.go`

**Imports pattern** (lines 1-14):
```go
package pgweb

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)
```

**Struct with mutex pattern** (lines 36-42):
```go
// Manager handles pgweb binary download, process lifecycle, and status reporting.
type Manager struct {
	paths *system.Paths
	fs    FileSystem
	cmd   CommandRunner
	proc  *exec.Cmd // nil when stopped
	mu    sync.Mutex
}
```

**Constructor injection pattern** (lines 47-50):
```go
// NewManager creates a Manager with injected dependencies.
func NewManager(paths *system.Paths, fs FileSystem, cmd CommandRunner) *Manager {
	return &Manager{
		paths: paths,
```

**Apply:** The tray `Controller` struct should follow the same constructor-injection pattern. Fields: `app AppController`, `iconData []byte`, `version string`. Use `New(app, iconData, version) *Controller` constructor. No mutex needed in Phase 1 (single-threaded tray operations dispatched to main queue via GCD).

---

### `internal/tray/controller.go` -- Consumer-Defined Interface

**Analog:** `internal/services/nginx/interfaces.go`

**Consumer-defined interface pattern** (lines 1-29):
```go
package nginx

import (
	"io/fs"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FileSystem is the filesystem subset that nginx.Manager needs.
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}

// LaunchdService is the launchd subset that nginx.Manager needs.
type LaunchdService interface {
	Install(cfg system.ServiceConfig) error
	Uninstall(cfg system.ServiceConfig) error
	IsRunning(label string) bool
}

// HelperRunner executes privileged helper actions.
type HelperRunner interface {
	Run(args ...string) (string, error)
}
```

**Also see:** `internal/cert/interfaces.go` (lines 1-18):
```go
// Package cert manages local CA generation and self-signed certificate creation
// for development sites.
package cert

import (
	"io/fs"
	"os"
)

// FileSystem is the filesystem subset that cert.Manager needs.
// Defined at the consumer site per Go interface convention (D-02).
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	Create(name string) (*os.File, error)
	OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error)
	Remove(name string) error
}

// AdminRunner abstracts privilege elevation for CA trust operations.
type AdminRunner interface {
	RunWithPrivileges(command string) error
}
```

**Apply:** Define `AppController` interface in `controller.go` (not a separate `interfaces.go` -- the tray only needs one narrow interface). Interface should contain only `Context() context.Context` as that's all the tray needs to call `runtime.Show`/`runtime.Quit`. Doc comment should reference D-02 convention.

---

### `internal/tray/icon.go` (utility, embed)

**Analog:** `main.go` lines 1-13

**Embed pattern** (lines 1-13):
```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS
```

**Apply:** Use `//go:embed icon_22x22.png` (and optionally `icon_22x22@2x.png`) to embed the template icon PNG as `[]byte`. The icon.go file should export the embedded bytes for the Controller constructor to consume.

---

### `main.go` (config modification -- Wails options)

**Analog:** Self -- current `main.go` lines 18-47

**Current Wails options** (lines 18-47):
```go
err := wails.Run(&options.App{
	Title:     "LamboServer",
	Width:     1100,
	Height:    700,
	MinWidth:  900,
	MinHeight: 600,
	AssetServer: &assetserver.Options{
		Assets: assets,
	},
	BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 1},
	OnStartup:         app.startup,
	OnBeforeClose:     app.beforeClose,
	OnShutdown:        app.shutdown,
	Mac: &mac.Options{
		TitleBar: &mac.TitleBar{
			TitlebarAppearsTransparent: true,
			HideTitle:                 true,
			FullSizeContent:           true,
		},
		WebviewIsTransparent: true,
		WindowIsTranslucent:  true,
		About: &mac.AboutInfo{
			Title:   "LamboServer",
			Message: "Open-source local development environment manager for Laravel & PHP",
		},
	},
	Bind: []interface{}{
		app,
	},
})
```

**Changes needed:**
1. Add `HideWindowOnClose: true` after `MinHeight` (D-08)
2. Remove `OnBeforeClose: app.beforeClose` line (D-08)

---

### `app.go` (controller modification -- lifecycle hooks)

**Analog:** Self -- current `app.go`

**App struct pattern** (lines 33-54):
```go
type App struct {
	ctx context.Context

	Paths      *system.Paths
	Config     *config.Store
	Launchd    *system.LaunchdManager
	Manager    *services.Manager
	Php        *php.Manager
	// ... more managers ...
	Debug      *logger.Logger
	Shell      *system.Integration
}
```

**Composition root pattern** (lines 59-108):
```go
func NewApp() *App {
	paths := system.NewPaths()
	// ... dependency creation ...
	return &App{
		Paths:      paths,
		// ... wiring ...
	}
}
```

**Startup lifecycle pattern** (lines 112-155):
```go
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Paths.EnsureDirectories()
	// ... initialization steps ...
}
```

**Shutdown lifecycle pattern** (lines 171-183):
```go
func (a *App) shutdown(ctx context.Context) {
	a.Debug.Info("LamboServer shutting down, stopping services...")
	a.MySQL.Stop()
	// ... service stops in dependency order ...
	a.Debug.Info("LamboServer shutdown complete")
}
```

**beforeClose pattern to remove** (lines 157-169):
```go
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	dialog, err := wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:          wailsRuntime.QuestionDialog,
		Title:         "Quit LamboServer?",
		Message:       "Menutup aplikasi akan menghentikan semua service (Nginx, DNS, PHP-FPM). Lanjutkan?",
		DefaultButton: "No",
		Buttons:       []string{"Yes", "No"},
	})
	if err != nil {
		return false
	}
	return dialog != "Yes"
}
```

**Changes needed:**
1. Add `Tray *tray.Controller` field to `App` struct (after `Shell`)
2. Add `import "github.com/rizkirmdhnnn/lamboserver/internal/tray"` to imports
3. In `NewApp()`: no tray initialization here (tray needs `a.ctx` from `startup`)
4. In `startup()`: after `a.ctx = ctx`, create and start tray: `a.Tray = tray.New(a, iconData, "1.0.0"); a.Tray.Start()`
5. Add `Context() context.Context` method on `App` to satisfy `tray.AppController` interface
6. Remove entire `beforeClose()` method (D-08)
7. In `shutdown()`: add `a.Tray.Destroy()` before service stops

**Debug.Action logging pattern** (from `app.go` various methods):
```go
a.Debug.Action(fmt.Sprintf("StartService(%s)", name), err)
```

**Apply:** Use `a.Debug.Action("TrayInit", err)` for tray initialization in `startup()`.

---

### `internal/tray/controller_darwin.go` (CGO bridge -- no analog)

No existing CGO files in the project. Use patterns from RESEARCH.md Example 3:

```go
// controller_darwin.go
package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "tray_darwin.h"
*/
import "C"

import "unsafe"
```

**Apply:** This file bridges Go to ObjC. Contains `//export` functions (`onShowWindow`, `onQuit`) called from ObjC, and Go functions that call C functions. Uses package-level `instance *Controller` variable for CGO callback routing.

---

### `internal/tray/tray_darwin.m` (ObjC implementation -- no analog)

No ObjC files exist in the project. Use patterns from RESEARCH.md Examples 3-4. Key requirements:
- Class name `LamboSystrayDelegate` (D-02)
- All UI mutations via `dispatch_async(dispatch_get_main_queue(), ...)` (D-03)
- `setTemplate:YES` on NSImage (D-04)
- Menu structure: disabled header, separator, Show Window, separator, Quit (D-06)

---

### `internal/tray/tray_unsupported.go` (stub -- no analog)

Standard Go build-tag stub pattern. File should have `//go:build !darwin` tag and provide no-op implementations of `Start()`, `Destroy()`.

---

## Shared Patterns

### Consumer-Defined Interfaces (D-02)
**Source:** `internal/services/nginx/interfaces.go` lines 1-29
**Apply to:** `internal/tray/controller.go` -- define `AppController` interface with only methods the tray calls

### Constructor Injection
**Source:** `internal/services/pgweb/manager.go` lines 47-50
**Apply to:** `internal/tray/controller.go` -- `New(app, iconData, version)` constructor

### Lifecycle Integration in App
**Source:** `app.go` lines 112-183 (`startup`, `shutdown`)
**Apply to:** `app.go` modifications -- tray init in `startup()` after `a.ctx = ctx`, tray destroy in `shutdown()` before service stops

### Debug Logging
**Source:** `app.go` line 119
```go
a.Debug.Info("LamboServer started")
```
**Apply to:** `app.go` tray initialization logging

### Embed Directive
**Source:** `main.go` line 13
```go
//go:embed all:frontend/dist
var assets embed.FS
```
**Apply to:** `internal/tray/icon.go` -- embed PNG icon files

### Doc Comment Convention
**Source:** `internal/services/dnsmasq/doc.go` lines 1-9
**Apply to:** `internal/tray/doc.go` -- package-level doc comment

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/tray/controller_darwin.go` | controller | event-driven | No CGO bridge files exist in the project; use RESEARCH.md patterns |
| `internal/tray/tray_darwin.h` | config | N/A | No ObjC header files exist; follow Apple conventions |
| `internal/tray/tray_darwin.m` | controller | event-driven | No ObjC implementation files exist; use RESEARCH.md patterns |
| `internal/tray/tray_unsupported.go` | controller | N/A | Standard Go build-tag stub; no complex analog needed |

## Metadata

**Analog search scope:** `internal/`, `main.go`, `app.go`
**Files scanned:** ~40 Go source files
**Pattern extraction date:** 2026-04-17
