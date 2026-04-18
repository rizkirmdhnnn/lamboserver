# Phase 01: Tray Foundation & Window Lifecycle - Research

**Researched:** 2026-04-17
**Domain:** macOS system tray (NSStatusBar) + Wails v2 window lifecycle
**Confidence:** HIGH

## Summary

This phase adds a system tray icon to LamboServer and changes the window close behavior to hide-to-tray instead of quit. The core technical challenge is that Wails v2 has no built-in tray support and all third-party Go systray libraries conflict with Wails' Objective-C symbols at the linker level. The solution is a custom CGO package (`internal/tray/`) that calls macOS `NSStatusBar`/`NSStatusItem` APIs directly with a uniquely-named ObjC delegate class.

The window lifecycle is straightforward: Wails v2.12.0 has a `HideWindowOnClose` option that calls `[NSApp hide:nil]` when the close button is pressed, which is the exact behavior needed. The existing `beforeClose` confirmation dialog is removed. "Show Window" from the tray calls `runtime.Show(ctx)` which calls `[NSApp unhide:self]` + `activateIgnoringOtherApps`. "Quit" calls `runtime.Quit(ctx)` which triggers the existing `shutdown()` flow.

**Primary recommendation:** Build a minimal CGO tray package with 4 files (Go API, Darwin CGO bridge, ObjC implementation, unsupported stub). Wire it into the App struct via the existing composition root pattern. Set `HideWindowOnClose: true` in Wails options and remove `OnBeforeClose` entirely.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Custom CGO package at `internal/tray/` using NSStatusBar/NSStatusItem directly. No third-party systray libraries.
- **D-02:** Use unique ObjC delegate class name `LamboSystrayDelegate` to avoid symbol collisions with Wails.
- **D-03:** All Cocoa UI mutations dispatched via `dispatch_async(dispatch_get_main_queue(), ...)`.
- **D-04:** Monochrome silhouette of existing LamboServer app logo, 22x22px macOS template PNG.
- **D-05:** Menu shows "LamboServer v{version}" as disabled text header.
- **D-06:** Menu: header, separator, Show Window, separator, Quit.
- **D-07:** Claude has discretion on exact separator placement and menu item ordering within Phase 1 structure.
- **D-08:** Set `HideWindowOnClose: true` in `main.go`. Remove existing `OnBeforeClose` / `beforeClose`.
- **D-09:** Use `runtime.Hide` / `runtime.Show` (application-level). Avoid `runtime.WindowHide` / `runtime.WindowShow`.
- **D-10:** "Show Window" tray menu item calls `runtime.Show(a.ctx)`.
- **D-11:** "Quit" from tray quits immediately with no confirmation dialog.
- **D-12:** "Quit" triggers existing `shutdown()` flow via `runtime.Quit(a.ctx)`, not `os.Exit`.
- **D-13:** Services keep running while app is minimized to tray. Only "Quit" stops them.

### Claude's Discretion
- Menu item ordering and separator placement within the Phase 1 structure
- CGO file organization within `internal/tray/` (tray.go, tray_darwin.go, tray.m, tray_unsupported.go, etc.)
- Whether to use `NSApplicationActivationPolicyAccessory` for Dock icon hiding at runtime (optional in Phase 1)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TRAY-01 | System tray icon visible in macOS menu bar when app is running | Custom CGO NSStatusBar package; icon embedded via `//go:embed`; initialized in `startup()` after `a.ctx` is set |
| TRAY-02 | Tray icon uses monochrome template image adapting to dark/light mode | NSImage `setTemplate:YES` on the icon; 22x22 black-on-transparent PNG; macOS handles inversion automatically |
| TRAY-03 | Custom CGO package using NSStatusBar directly, no third-party systray libraries | `internal/tray/` with tray.go + tray_darwin.go + tray_darwin.m + tray_unsupported.go; `LamboSystrayDelegate` ObjC class |
| WNDW-01 | Closing main window hides to tray instead of quitting | `HideWindowOnClose: true` in Wails options calls `[NSApp hide:nil]`; remove `OnBeforeClose` callback |
| WNDW-02 | "Show Window" menu item reopens main application window | `runtime.Show(a.ctx)` calls `[NSApp unhide:self]` + `activateIgnoringOtherApps:TRUE` |
| WNDW-03 | "Quit" menu item stops all services and exits | `runtime.Quit(a.ctx)` triggers `OnShutdown` -> existing `shutdown()` flow |
| WNDW-04 | Services keep running while app is minimized to tray | `HideWindowOnClose` only hides the NSApplication; launchd daemons are independent processes unaffected by hide |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Tray icon rendering | OS / Native (AppKit) | -- | NSStatusBar is a macOS system API; all tray UI is native Cocoa |
| Tray menu handling | OS / Native (AppKit) | Go backend | NSMenu renders natively; click callbacks dispatch to Go via CGO |
| Window hide/show | Wails runtime | OS / Native | `runtime.Show`/`Hide` wrap `[NSApp unhide/hide]`; Wails owns the window lifecycle |
| Service lifecycle on quit | Go backend | OS / Native (launchd) | `shutdown()` in app.go calls service managers which control launchd daemons |
| Tray initialization | Go backend | OS / Native | `tray.Controller` created in `startup()`, dispatches NSStatusItem creation via GCD |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Wails v2 | v2.12.0 | Desktop framework, window lifecycle, `HideWindowOnClose` | Already in use; `HideWindowOnClose` verified in source [VERIFIED: Wails v2.12.0 source `pkg/options/options.go`] |
| macOS AppKit (NSStatusBar) | System framework | Tray icon and menu | Only way to create menu bar items on macOS [CITED: Apple NSStatusBar docs] |
| CGO | Go toolchain | Bridge Go to Objective-C | Required for NSStatusBar; already enabled by Wails [VERIFIED: Wails darwin `#cgo LDFLAGS`] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `//go:embed` | Go stdlib | Embed tray icon PNG in binary | Icon asset embedding for the 22x22 template PNG |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom CGO NSStatusBar | getlantern/systray | Linker conflict with Wails ObjC classes -- unusable [VERIFIED: Wails issue #3003, Discussion #4514] |
| Custom CGO NSStatusBar | fyne.io/systray | Same ObjC linker conflict despite `RunWithExternalLoop` [VERIFIED: Wails Discussion #4514] |
| Custom CGO NSStatusBar | energye/systray | Same ObjC conflict [VERIFIED: Wails issue #3003] |
| `HideWindowOnClose: true` | Custom `beforeClose` handler | `HideWindowOnClose` is built-in, calls `[NSApp hide:nil]` directly in ObjC -- cleaner [VERIFIED: WindowDelegate.m source] |

**Installation:**
```bash
# No new dependencies needed. Internal package + system frameworks only.
```

**Dependency delta:** Zero new `go.mod` entries. The tray uses only macOS system frameworks (Foundation, Cocoa/AppKit) already linked by Wails.

## Architecture Patterns

### System Architecture Diagram

```
User clicks close (X)
  |
  v
Wails WindowDelegate.windowShouldClose
  |-- hideOnClose=true --> [NSApp hide:nil]  --> Window hidden, app alive
  |                                               Services keep running (launchd)
  |                                               Tray icon remains visible
  |
User clicks tray icon
  |
  v
macOS shows NSMenu (built by tray.Controller)
  |
  +-- "Show Window" --> runtime.Show(ctx)
  |                       |
  |                       v
  |                     [NSApp unhide:self] + activateIgnoringOtherApps
  |                       |
  |                       v
  |                     Window visible again
  |
  +-- "Quit" --> runtime.Quit(ctx)
                   |
                   v
                 Wails OnShutdown
                   |
                   v
                 app.shutdown(ctx)
                   |
                   v
                 Stop MySQL, pgweb, PostgreSQL, PHP-FPM, Nginx, DNS
                   |
                   v
                 tray.Controller.Destroy() -- remove NSStatusItem
                   |
                   v
                 Process exits
```

### Recommended Project Structure

```
internal/tray/
  doc.go               # Package doc comment
  controller.go        # Controller struct, public API: New(), Start(), Destroy()
  controller_darwin.go  # CGO bridge: Go functions wrapping C calls
  tray_darwin.m        # Objective-C: LamboSystrayDelegate, NSStatusItem, NSMenu
  tray_darwin.h        # ObjC header for the delegate class
  tray_unsupported.go  # Build tag !darwin -- stub with no-op methods
  icon.go              # //go:embed for the template icon PNG
  icon_22x22.png       # Monochrome template icon (1x)
  icon_22x22@2x.png    # Monochrome template icon (2x Retina)
```

### Pattern 1: CGO Dispatch to Main Queue via GCD

**What:** All Cocoa/AppKit calls must execute on the main thread. The tray ObjC code uses GCD to dispatch work onto the main queue from any goroutine.

**When to use:** Every time Go code needs to create, update, or destroy an NSStatusItem, NSMenu, or NSMenuItem.

**Example:**
```objc
// Source: Apple GCD docs + Wails community pattern (Discussion #4514)
// tray_darwin.m
#import <Cocoa/Cocoa.h>

@interface LamboSystrayDelegate : NSObject <NSMenuDelegate>
@property (strong, nonatomic) NSStatusItem *statusItem;
@property (strong, nonatomic) NSMenu *menu;
@end

@implementation LamboSystrayDelegate

void CreateTray(const void *iconData, int iconLen) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSStatusBar *bar = [NSStatusBar systemStatusBar];
        NSStatusItem *item = [bar statusItemWithLength:NSVariableStatusItemLength];

        NSData *data = [NSData dataWithBytes:iconData length:iconLen];
        NSImage *icon = [[NSImage alloc] initWithData:data];
        [icon setTemplate:YES];  // Critical: enables dark/light mode adaptation
        [icon setSize:NSMakeSize(22, 22)];
        item.button.image = icon;

        // Menu setup...
        item.menu = /* built NSMenu */;
    });
}

@end
```

```go
// Source: Standard CGO pattern for macOS
// controller_darwin.go
package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "tray_darwin.h"
*/
import "C"

import "unsafe"

func (c *Controller) createNative() {
    iconPtr := unsafe.Pointer(&c.iconData[0])
    C.CreateTray(iconPtr, C.int(len(c.iconData)))
}
```

### Pattern 2: Consumer-Defined Interface for App Handler (D-02)

**What:** The tray package defines a narrow interface listing only the App methods it calls, following the project's consumer-defined interface convention.

**When to use:** For the tray's reference to the App struct. The tray needs `ShowWindow()`, `Quit()`, and access to the Wails context.

**Example:**
```go
// internal/tray/controller.go
package tray

import "context"

// AppController defines the narrow interface the tray needs from the application.
// Follows consumer-defined interface pattern (D-02).
type AppController interface {
    // Context returns the Wails runtime context for Show/Hide/Quit calls.
    Context() context.Context
}
```

### Pattern 3: Template Icon for Dark/Light Mode

**What:** Set `[icon setTemplate:YES]` on the NSImage to let macOS automatically adapt the icon color for menu bar appearance.

**When to use:** Always, for any menu bar icon on macOS.

**Example:**
```objc
// Source: Apple NSImage setTemplate docs
NSImage *icon = [[NSImage alloc] initWithData:data];
[icon setTemplate:YES];  // macOS will invert for dark menu bar
[icon setSize:NSMakeSize(22, 22)];
statusItem.button.image = icon;
```

**Icon requirements:**
- Black silhouette on transparent background [CITED: Apple Human Interface Guidelines - Menu Bar Extras]
- 22x22 points (1x), 44x44 pixels (2x Retina)
- PNG with alpha channel
- Source: derive from `build/appicon.png` (1024x1024 RGBA) [VERIFIED: `sips --getProperty` on build/appicon.png]

### Anti-Patterns to Avoid

- **Using `os.Exit(0)` for Quit:** Bypasses Wails shutdown sequence. Services remain running as orphaned launchd daemons. Always use `runtime.Quit(ctx)`. [VERIFIED: Wails source -- `runtime.Quit` calls `appFrontend.Quit()` which triggers `OnShutdown`]
- **Using `runtime.WindowHide`/`runtime.WindowShow` instead of `runtime.Hide`/`runtime.Show`:** Window-level hide has a documented freeze bug (Wails #2572). Application-level hide (`[NSApp hide:nil]`) is what `HideWindowOnClose` uses internally. [VERIFIED: WindowDelegate.m line 18]
- **Initializing Cocoa objects in Go `init()`:** Wails' own darwin `init()` calls `runtime.LockOSThread()`. Adding another `init()` that does Cocoa setup will race with Wails. Only create NSStatusItem in the explicitly called `Start()` method after `startup()` runs. [VERIFIED: Wails darwin/window.go line 30]
- **Using `systray.Run()` from any Go library:** Attempts to own the main event loop, conflicts with Wails' NSApplication loop. [VERIFIED: multiple Wails issues]
- **Polling service status on a timer:** Not needed for Phase 1. The existing architecture is pull-on-action. Status display is Phase 2 scope.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dark/light mode icon adaptation | Custom color detection + icon swapping | `[NSImage setTemplate:YES]` | macOS handles inversion automatically; custom detection is fragile across OS versions |
| Window hide on close | Custom `beforeClose` with `runtime.Hide` | `HideWindowOnClose: true` in Wails options | Built into Wails v2.12.0; handles the ObjC-level `[NSApp hide:nil]` correctly, avoids #2572 bug |
| Quit with service shutdown | Direct `os.Exit` + manual shutdown | `runtime.Quit(ctx)` | Triggers Wails' `OnShutdown` which calls existing `shutdown()` with correct dependency ordering |
| GCD main queue dispatch | Custom thread management | `dispatch_async(dispatch_get_main_queue(), ^{...})` | Standard macOS pattern; Wails already runs NSApplication on main thread |

## Common Pitfalls

### Pitfall 1: ObjC Class Name Collision with Wails
**What goes wrong:** Using common ObjC class names (`AppDelegate`, `WindowDelegate`, `SystrayDelegate`) causes linker duplicate symbol errors.
**Why it happens:** Wails v2 defines `AppDelegate`, `WindowDelegate`, `WailsContext`, `WailsWindow`, `WailsMenu`, `WailsMenuItem`, `WailsWebView`, `WailsAlert`, `CustomProtocolSchemeHandler` as ObjC classes. [VERIFIED: grep of Wails darwin `.h` files]
**How to avoid:** Use `LamboSystrayDelegate` as the ObjC class name (decision D-02). Avoid any class name starting with `Wails` or matching the names above.
**Warning signs:** `ld: duplicate symbol '_OBJC_CLASS_$_...'` during `wails build`.

### Pitfall 2: Nil Context Panic in Tray Callbacks
**What goes wrong:** Tray menu click callbacks call `runtime.Show(ctx)` or `runtime.Quit(ctx)` with a nil context, causing a panic.
**Why it happens:** `a.ctx` is set in `startup()` but tray callbacks can fire from ObjC at any time. If `runtime.Quit(nil)` is called, Wails calls `log.Fatalf` which crashes with no cleanup. [VERIFIED: Wails source `pkg/runtime/runtime.go` line 65-66]
**How to avoid:** Guard all tray callbacks: `if ctx == nil { return }`. Initialize the tray after `a.ctx` is assigned in `startup()`. Never expose tray menu items before context is ready.
**Warning signs:** App crashes on tray menu click with `cannot call 'runtime.Quit': context not available`.

### Pitfall 3: HideWindowOnClose Bypasses OnBeforeClose
**What goes wrong:** Setting `HideWindowOnClose: true` means `OnBeforeClose` is never called. Any logic in `beforeClose` stops executing.
**Why it happens:** In Wails' `WindowDelegate.m`, when `hideOnClose` is true, the method returns `false` immediately after `[NSApp hide:nil]`, before reaching `processMessage("Q")` which triggers `OnBeforeClose`. [VERIFIED: WindowDelegate.m lines 16-20]
**How to avoid:** This is the desired behavior -- we want to remove `beforeClose`. But be aware that `OnBeforeClose` becomes a dead code path when `HideWindowOnClose` is set. Remove it entirely to avoid confusion.
**Warning signs:** `beforeClose` logic still present in code but never executing.

### Pitfall 4: Shutdown Race Between Tray Quit and Wails Lifecycle
**What goes wrong:** Double-invocation of shutdown logic if both tray `onQuit` and Wails `OnShutdown` try to stop services.
**Why it happens:** `runtime.Quit(ctx)` triggers Wails shutdown which calls `OnShutdown` -> `app.shutdown()`. If the tray quit handler also calls service stop methods, services are stopped twice.
**How to avoid:** Tray "Quit" handler should only call `runtime.Quit(ctx)`. All service shutdown stays in `app.shutdown()`. Tray cleanup (removing NSStatusItem) also goes in `shutdown()` or uses `sync.Once`.
**Warning signs:** "Service already stopped" errors in logs after quitting.

### Pitfall 5: Icon Invisible on Light Menu Bar
**What goes wrong:** A white or light-colored icon is invisible against a light macOS menu bar.
**Why it happens:** The icon is not set as a template image. Without `[icon setTemplate:YES]`, macOS renders the PNG as-is without adapting to the menu bar appearance.
**How to avoid:** Always call `setTemplate:YES`. Use a black silhouette on transparent background. macOS will invert it for dark mode automatically. [CITED: Apple NSImage Template docs]
**Warning signs:** Tray area shows blank space where the icon should be.

## Code Examples

### Example 1: Wails Options Configuration (main.go changes)

```go
// Source: Verified from Wails v2.12.0 source pkg/options/options.go
// and WindowDelegate.m behavior

err := wails.Run(&options.App{
    Title:     "LamboServer",
    Width:     1100,
    Height:    700,
    MinWidth:  900,
    MinHeight: 600,
    HideWindowOnClose: true,  // NEW: hides to tray on close
    AssetServer: &assetserver.Options{
        Assets: assets,
    },
    BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 1},
    OnStartup:    app.startup,
    // OnBeforeClose: REMOVED (D-08)
    OnShutdown:   app.shutdown,
    Mac: &mac.Options{
        // ... existing mac options unchanged ...
    },
    Bind: []interface{}{
        app,
    },
})
```

### Example 2: Controller struct and lifecycle

```go
// internal/tray/controller.go
package tray

import "context"

// AppController is the narrow interface the tray needs from the application.
type AppController interface {
    Context() context.Context
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
```

### Example 3: CGO callback from ObjC to Go

```go
// controller_darwin.go
package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "tray_darwin.h"
*/
import "C"

//export onShowWindow
func onShowWindow() {
    if instance == nil || instance.app == nil {
        return
    }
    ctx := instance.app.Context()
    if ctx == nil {
        return
    }
    // Use Wails runtime.Show via the context
    // (import wailsRuntime in the calling package, not here)
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
    // Use Wails runtime.Quit via the context
}
```

### Example 4: ObjC Menu Construction

```objc
// tray_darwin.m -- menu creation
void SetupMenu(const char* version) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSMenu *menu = [[NSMenu alloc] init];

        // Header: "LamboServer v1.0.0" (disabled)
        NSString *headerTitle = [NSString stringWithFormat:@"LamboServer v%s",
                                 version];
        NSMenuItem *header = [[NSMenuItem alloc] initWithTitle:headerTitle
                                                        action:nil
                                                 keyEquivalent:@""];
        [header setEnabled:NO];
        [menu addItem:header];

        [menu addItem:[NSMenuItem separatorItem]];

        // Show Window
        NSMenuItem *showItem = [[NSMenuItem alloc]
            initWithTitle:@"Show Window"
                   action:@selector(onShowWindow:)
            keyEquivalent:@""];
        [showItem setTarget:delegate];
        [menu addItem:showItem];

        [menu addItem:[NSMenuItem separatorItem]];

        // Quit
        NSMenuItem *quitItem = [[NSMenuItem alloc]
            initWithTitle:@"Quit"
                   action:@selector(onQuit:)
            keyEquivalent:@""];
        [quitItem setTarget:delegate];
        [menu addItem:quitItem];

        statusItem.menu = menu;
    });
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| getlantern/systray | Custom CGO (for Wails v2) | 2024-2025 | Only working approach; all Go systray libs conflict with Wails ObjC |
| `runtime.WindowHide` for close-to-tray | `HideWindowOnClose: true` option | Wails v2.x | Built-in option avoids freeze bug (#2572); cleaner lifecycle |
| `OnBeforeClose` for quit confirmation | Remove entirely with `HideWindowOnClose` | This phase | `HideWindowOnClose` bypasses `OnBeforeClose` at ObjC level |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `dispatch_async(dispatch_get_main_queue())` from a CGO callback will correctly serialize onto Wails' NSApplication main thread | Architecture Patterns | HIGH -- if GCD dispatch doesn't coexist with Wails' event loop, the entire tray approach fails. Mitigation: verified by community implementation in Wails Discussion #4514. |
| A2 | `runtime.Show(ctx)` correctly restores the window after `HideWindowOnClose` hides it | Code Examples | MEDIUM -- verified that both use `[NSApp hide/unhide]` semantics, but round-trip behavior not tested on v2.12.0 specifically. Must be tested at implementation time. |
| A3 | The 22x22 template icon derived from the existing app logo will be visually recognizable at that size | Architecture Patterns | LOW -- if the logo has too much detail, a simpler geometric icon may be needed. Easy to iterate. |

## Open Questions (RESOLVED)

1. **Icon design: silhouette from existing logo vs. custom icon**
   - What we know: The existing `build/appicon.png` is 1024x1024 RGBA, likely a detailed color logo. Decision D-04 says "monochrome silhouette of existing LamboServer app logo."
   - What's unclear: Whether the logo's detail level is suitable for 22x22px rendering as a monochrome silhouette.
   - RESOLVED: Create silhouette from build/appicon.png; fall back to simpler icon if unrecognizable at 22x22. This is a design iteration, not a blocker.

2. **Dock icon behavior when window is hidden**
   - What we know: `HideWindowOnClose` calls `[NSApp hide:nil]` which hides the app but leaves the Dock icon visible. D-09 in context marks Dock icon hiding as optional for Phase 1. TRAY-06 (dynamic Dock hiding) is deferred to v2.
   - What's unclear: Whether users will be confused by the Dock icon remaining while "hidden to tray."
   - RESOLVED: Dock icon stays visible in Phase 1 — matches [NSApp hide] behavior (same as Cmd+H). TRAY-06 deferred to v2.

3. **App version string for menu header**
   - What we know: `wails.json` has `info.productVersion: "1.0.0"`. The app can read this at build time or from the app bundle at runtime.
   - What's unclear: Best mechanism to pass the version string to the tray controller.
   - RESOLVED: Hardcode "1.0.0" in tray.New() call in startup(); version sourced from wails.json productVersion via hardcode for Phase 1.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Build | Yes | go1.26.1 | -- |
| Wails CLI | Build | Yes | v2.12.0 | -- |
| Xcode Command Line Tools | CGO compilation of ObjC | Yes (implied by working Wails build) | -- | -- |
| macOS AppKit framework | NSStatusBar, NSMenu | Yes (system) | macOS 12.0+ | -- |

**Missing dependencies:** None. All required tools are available.

## Security Domain

Security enforcement is not explicitly disabled in config. However, this phase has minimal security surface.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | N/A -- tray is local UI only |
| V3 Session Management | No | N/A |
| V4 Access Control | No | N/A -- tray delegates to existing App methods which already handle auth |
| V5 Input Validation | No | Tray menu has fixed items, no user text input |
| V6 Cryptography | No | N/A |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Tray menu spoofing | Spoofing | Not applicable -- NSStatusItem is OS-managed, cannot be spoofed by external process |
| Unauthorized service control | Elevation | Services already require privileged helper (sudoers); tray just calls existing App methods |

## Sources

### Primary (HIGH confidence)
- Wails v2.12.0 source code -- verified `HideWindowOnClose`, `WindowDelegate.m`, `runtime.Show/Hide/Quit`, ObjC class names [VERIFIED: local source in GOMODCACHE]
- Apple NSStatusBar/NSStatusItem documentation [CITED: developer.apple.com/documentation/appkit/nsstatusbar]
- Apple NSImage setTemplate documentation [CITED: developer.apple.com/documentation/appkit/nsimage/1520017-template]
- Project source: `app.go`, `main.go`, `wails.json` [VERIFIED: local files]

### Secondary (MEDIUM confidence)
- Wails Discussion #4514 -- community implementation of custom CGO tray for Wails v2 [CITED: github.com/wailsapp/wails/discussions/4514]
- Wails Issue #2572 -- WindowHide freeze bug confirmation [CITED: github.com/wailsapp/wails/issues/2572]

### Tertiary (LOW confidence)
- Exact behavior of `dispatch_async` coexisting with Wails' NSApplication run loop -- verified conceptually but not tested in v2.12.0 specifically [ASSUMED -- A1]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- verified in Wails v2.12.0 source; no external dependencies
- Architecture: HIGH -- verified `HideWindowOnClose` ObjC implementation, `runtime.Show/Hide/Quit` code paths, ObjC class names to avoid
- Pitfalls: HIGH -- verified against Wails source for WindowDelegate behavior, runtime.Quit nil-guard, ObjC class names

**Research date:** 2026-04-17
**Valid until:** 2026-05-17 (stable -- Wails v2 is in maintenance mode, no breaking changes expected)
