---
phase: 01-tray-foundation-window-lifecycle
reviewed: 2026-04-17T16:24:01Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - app.go
  - internal/tray/controller.go
  - internal/tray/controller_darwin.go
  - internal/tray/doc.go
  - internal/tray/icon.go
  - internal/tray/tray_darwin.h
  - internal/tray/tray_darwin.m
  - internal/tray/tray_unsupported.go
  - main.go
findings:
  critical: 1
  warning: 4
  info: 3
  total: 8
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-04-17T16:24:01Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

The tray foundation work is well-structured. The package follows project conventions (doc.go, consumer-defined interface, build tags for platform gating, no-op stub for non-darwin). The CGO bridge is minimal and correct in the common path.

Two areas require attention before shipping: a crash-level nil dereference in `Start()` when the icon slice is empty, and a race on the module-level `instance` pointer that is written from Go and read from CGO callbacks without any synchronization. The remaining findings are warnings and info items about missing `OnBeforeClose` wiring, a `SaveConfig` bug that discards the caller-supplied config, and a few conventions gaps.

---

## Critical Issues

### CR-01: Nil dereference in `Start()` when `iconData` is empty

**File:** `internal/tray/controller_darwin.go:26`

**Issue:** `&c.iconData[0]` panics with an index-out-of-range (nil slice / zero-length slice) if `iconData` is empty. The embedded `icon_22x22.png` is always non-empty in a correctly built binary, but the constructor (`New`) accepts any `[]byte` with no validation. A caller who passes `nil` or an empty slice causes an immediate nil/index panic before any CGO boundary is reached.

**Fix:**
```go
func (c *Controller) Start() {
    if len(c.iconData) == 0 {
        // Fail fast with a clear message rather than a cryptic index panic.
        panic("tray: iconData must not be empty")
    }
    instance = c
    iconPtr := unsafe.Pointer(&c.iconData[0])
    // ...
}
```

Alternatively, guard in `New()` and return an error (preferred if callers can recover):
```go
func New(app AppController, iconData []byte, version string) (*Controller, error) {
    if len(iconData) == 0 {
        return nil, errors.New("tray: iconData must not be empty")
    }
    return &Controller{app: app, iconData: iconData, version: version}, nil
}
```

---

## Warnings

### WR-01: Data race on the `instance` package variable

**File:** `internal/tray/controller_darwin.go:21`

**Issue:** The package-level `instance *Controller` variable is written in `Start()` (Go goroutine) and read in `onShowWindow()` / `onQuit()` (CGO callbacks dispatched from the Obj-C main thread). There is no mutex, `sync/atomic`, or `sync.Once` protecting these accesses. Under Go's memory model, this is an unsynchronized concurrent read/write — a data race. While in practice the OS main thread and the Go startup goroutine are serialized by the time `dispatch_async` fires, the Go race detector will flag this and correctness relies on an undocumented ordering assumption.

**Fix:**
```go
import "sync"

var (
    instanceMu sync.RWMutex
    instance   *Controller
)

func (c *Controller) Start() {
    instanceMu.Lock()
    instance = c
    instanceMu.Unlock()
    // ...
}

func (c *Controller) Destroy() {
    C.DestroyTray()
    instanceMu.Lock()
    instance = nil
    instanceMu.Unlock()
}

//export onShowWindow
func onShowWindow() {
    instanceMu.RLock()
    c := instance
    instanceMu.RUnlock()
    if c == nil || c.app == nil {
        return
    }
    // ...
}
```

### WR-02: `DestroyTray` is async but `Destroy()` treats it as synchronous

**File:** `internal/tray/tray_darwin.m:75-83`, `internal/tray/controller_darwin.go:33-36`

**Issue:** `DestroyTray()` dispatches work to the main queue with `dispatch_async` (fire-and-forget). The Go `Destroy()` method returns immediately, then sets `instance = nil`. The `shutdown()` lifecycle in `app.go` calls `a.Tray.Destroy()` and then proceeds to stop services. If a CGO callback (`onShowWindow`, `onQuit`) fires between `Destroy()` returning and the async block executing, `instance` has already been set to nil in Go but the NSStatusItem is still live in ObjC — the callbacks can still arrive. This is a narrow but real window for a use-after-free on the Obj-C side (accessing `delegate` after `delegate = nil` was set by a concurrent async block).

**Fix:** Use `dispatch_sync` instead of `dispatch_async` in `DestroyTray` so the call blocks until the NSStatusItem is actually removed before returning to Go:

```objc
void DestroyTray(void) {
    // Use sync so the item is fully removed before returning to Go.
    dispatch_sync(dispatch_get_main_queue(), ^{
        if (delegate != nil && delegate.statusItem != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:delegate.statusItem];
            delegate.statusItem = nil;
        }
        delegate = nil;
    });
}
```

Note: `dispatch_sync` on the main queue from a non-main thread is safe. If `DestroyTray` could ever be called from the main thread (unlikely given the Wails shutdown path), guard with `NSThread.isMainThread`.

### WR-03: `SaveConfig` discards the caller-supplied `cfg` argument

**File:** `app.go:716`

**Issue:** The public method `SaveConfig(cfg config.AppConfig) error` receives a full config struct from the frontend but calls `a.Config.Save()` with no arguments — it persists whatever is already in the in-memory store, silently ignoring `cfg`. Any config changes the user made in the frontend are lost.

```go
// Current — cfg is never used:
func (a *App) SaveConfig(cfg config.AppConfig) error {
    return a.Config.Save()
}
```

**Fix:** Apply the new config to the store before saving:
```go
func (a *App) SaveConfig(cfg config.AppConfig) error {
    if err := a.Config.Set(cfg); err != nil {
        return err
    }
    return a.Config.Save()
}
```
(Exact method name on `config.Store` depends on the store API; the intent is to write `cfg` into the store before flushing to disk.)

### WR-04: `HideWindowOnClose` is set but `OnBeforeClose` is not wired — no graceful confirmation

**File:** `main.go:24`

**Issue:** `HideWindowOnClose: true` hides the window when the user clicks the red close button, which is correct for a tray app. However, the Wails `OnBeforeClose` callback is not set. The original `app.go` had a `beforeClose()` method (referenced in the CLAUDE.md architecture section as showing a confirmation dialog), but it is absent from the reviewed `app.go`. Without `OnBeforeClose`, there is no hook to show a "Are you sure you want to quit?" dialog when `onQuit()` triggers `wailsRuntime.Quit()`. The app quits immediately with no confirmation.

If a confirmation dialog is not desired, this is fine. If it is desired (matching the original behavior described in the CLAUDE.md architecture section), wire it up:

```go
// In main.go options:
OnBeforeClose: app.beforeClose,

// In app.go:
func (a *App) beforeClose(ctx context.Context) bool {
    dialog, _ := wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
        Type:    wailsRuntime.QuestionDialog,
        Title:   "Quit LamboServer?",
        Message: "This will stop all services.",
        Buttons: []string{"Quit", "Cancel"},
    })
    return dialog != "Quit" // return true to prevent close
}
```

---

## Info

### IN-01: `Icon2x` is embedded but never used

**File:** `internal/tray/icon.go:8-9`

**Issue:** `Icon2x []byte` embeds `icon_22x22@2x.png` but is not referenced anywhere in the codebase. The 2x asset is compiled into the binary unnecessarily.

**Fix:** Either pass `Icon2x` to `CreateTray` for Retina support (recommended), or remove the embed if retina density is not planned. The NSImage on macOS will automatically pick the @2x variant if the image is constructed from a multi-representation NSImage, but the current ObjC code uses a single `initWithData:` call that receives only the 1x data.

### IN-02: `version` is hardcoded as `"1.0.0"` in `startup()`

**File:** `app.go:154`

**Issue:** `tray.New(a, tray.Icon, "1.0.0")` hardcodes the version string. This will silently show the wrong version in the tray menu header as the project evolves.

**Fix:** Define a version constant or variable at the package level in `main.go` (or inject it at build time via `-ldflags`), then pass it through:

```go
// main.go
var version = "dev" // overridden by -ldflags "-X main.version=1.2.3"

// NewApp could accept version, or startup() could receive it
```

### IN-03: Import ordering in `app.go` does not follow stdlib / external / internal grouping

**File:** `app.go:3-25`

**Issue:** The imports in `app.go` mix stdlib (`context`, `fmt`, `os`), an external dependency (`github.com/pkg/browser`), and internal packages in two separate groups inconsistently — `"github.com/rizkirmdhnnn/lamboserver/internal/tray"` appears between other internal packages but after one external package. Go convention (and `goimports`) expects three groups: stdlib, external, internal. This is a minor style item but inconsistent with the rest of the codebase.

**Fix:** Reorder to the standard three-group format:
```go
import (
    "context"
    "fmt"
    "os"

    "github.com/pkg/browser"
    wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

    "github.com/rizkirmdhnnn/lamboserver/internal/cert"
    "github.com/rizkirmdhnnn/lamboserver/internal/config"
    // ... all internal packages
    "github.com/rizkirmdhnnn/lamboserver/internal/tray"
)
```

---

_Reviewed: 2026-04-17T16:24:01Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
