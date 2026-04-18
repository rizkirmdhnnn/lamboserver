# Phase 2: Service Controls - Pattern Map

**Mapped:** 2026-04-18
**Files analyzed:** 4 (all modifications to existing files)
**Analogs found:** 4 / 4

## File Classification

| Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------|------|-----------|----------------|---------------|
| `internal/tray/controller.go` | interface | request-response | itself (current Phase 1 version) | exact |
| `internal/tray/controller_darwin.go` | controller/bridge | request-response + event-driven | itself (current Phase 1 version) | exact |
| `internal/tray/tray_darwin.h` | config (C header) | N/A | itself (current Phase 1 version) | exact |
| `internal/tray/tray_darwin.m` | component (ObjC UI) | event-driven | itself (current Phase 1 version) | exact |

All files are modifications to the existing Phase 1 tray package. The analog for every file is its own current implementation -- the patterns established in Phase 1 are extended, not replaced.

## Pattern Assignments

### `internal/tray/controller.go` (interface, request-response)

**Analog:** itself -- current Phase 1 implementation

**Consumer-defined interface pattern** (lines 5-10):
```go
// AppController is the narrow interface the tray needs from the application.
// Follows consumer-defined interface convention (D-02).
type AppController interface {
	// Context returns the Wails runtime context for Show/Hide/Quit calls.
	Context() context.Context
}
```

**What to add:** Four new methods matching the signatures on `App` in `app.go` lines 257-311:
- `GetAllStatuses() map[string]string` (app.go line 298)
- `StartService(name string) error` (app.go line 258)
- `StopService(name string) error` (app.go line 270)
- `RestartService(name string) error` (app.go line 286)

**Doc comment pattern** -- each method gets a single-line `// MethodName verb-phrase.` comment, matching the existing `Context()` doc comment style on line 8.

---

### `internal/tray/controller_darwin.go` (controller/bridge, request-response + event-driven)

**Analog:** itself -- current Phase 1 implementation

**Import block pattern** (lines 1-17):
```go
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
```

**Singleton instance pattern** (line 20):
```go
// instance holds the singleton Controller for CGO callback routing.
var instance *Controller
```

**CGO export callback pattern -- no parameters** (lines 38-48):
```go
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
```

**Key patterns to replicate for new callbacks:**
1. Guard clause: `if instance == nil || instance.app == nil { return }` -- every exported callback must have this.
2. For `onServiceAction(cName *C.char, cAction *C.char)`: convert C strings with `C.GoString()`, then dispatch in a goroutine per D-06.
3. For `onRefreshStatuses()`: call `instance.app.GetAllStatuses()`, iterate in fixed order, pass each status to ObjC via `C.UpdateServiceStatus()`, free C strings with `C.free(unsafe.Pointer(...))`.

**C string lifecycle pattern** from `Start()` (lines 27-29):
```go
cVersion := C.CString(c.version)
defer C.free(unsafe.Pointer(cVersion))
C.CreateTray(iconPtr, C.int(len(c.iconData)), cVersion)
```

When passing strings from Go to C, always `C.CString()` then `C.free()`. For the refresh loop where multiple strings are created, free each immediately after the C call (no defer in a loop).

---

### `internal/tray/tray_darwin.h` (C header)

**Analog:** itself -- current Phase 1 implementation

**Header guard and declaration pattern** (lines 1-7):
```c
#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

void CreateTray(const void *iconData, int iconLen, const char *version);
void DestroyTray(void);

#endif
```

**What to add:** Two new function declarations:
- `void UpdateServiceStatus(int index, const char *status);` -- called from Go during refresh
- `void RefreshServiceStatuses(void);` -- called from ObjC `menuWillOpen:` into Go

---

### `internal/tray/tray_darwin.m` (ObjC component, event-driven)

**Analog:** itself -- current Phase 1 implementation

**Go callback extern declarations** (lines 4-6):
```objc
// Forward declarations for Go callbacks
extern void onShowWindow(void);
extern void onQuit(void);
```

New externs to add: `extern void onServiceAction(const char *name, const char *action);` and `extern void onRefreshStatuses(void);`.

**Delegate interface pattern** (lines 8-11):
```objc
@interface LamboSystrayDelegate : NSObject
@property (strong, nonatomic) NSStatusItem *statusItem;
@end
```

Extend with: `<NSMenuDelegate>` protocol conformance, `NSMutableArray<NSMenuItem *> *serviceItems` property, `NSMutableArray<NSMenu *> *serviceSubmenus` property.

**Action selector pattern** (lines 17-19, 21-23):
```objc
- (void)showWindow:(id)sender {
    onShowWindow();
}

- (void)quit:(id)sender {
    onQuit();
}
```

New service action selectors follow same pattern but extract the service index from `[sender tag]` and call `onServiceAction()` with the corresponding service key and action string.

**GCD dispatch pattern for all Cocoa operations** (lines 28, 76):
```objc
void CreateTray(const void *iconData, int iconLen, const char *version) {
    dispatch_async(dispatch_get_main_queue(), ^{
        // ... all NSMenu/NSStatusItem work inside this block
    });
}
```

All new C functions called from Go (`UpdateServiceStatus`) must wrap UI mutations in `dispatch_async(dispatch_get_main_queue(), ^{ ... })`.

**Menu construction pattern** (lines 41-71):
```objc
NSMenu *menu = [[NSMenu alloc] init];

// Header: "LamboServer v1.0.0" (disabled)
NSString *headerTitle = [NSString stringWithFormat:@"LamboServer v%s", version];
NSMenuItem *header = [[NSMenuItem alloc] initWithTitle:headerTitle
                                                action:nil
                                         keyEquivalent:@""];
[header setEnabled:NO];
[menu addItem:header];

[menu addItem:[NSMenuItem separatorItem]];

// Show Window
NSMenuItem *showItem = [[NSMenuItem alloc]
    initWithTitle:@"Show Window"
           action:@selector(showWindow:)
    keyEquivalent:@""];
[showItem setTarget:delegate];
[menu addItem:showItem];

[menu addItem:[NSMenuItem separatorItem]];

// Quit
NSMenuItem *quitItem = [[NSMenuItem alloc]
    initWithTitle:@"Quit"
           action:@selector(quit:)
    keyEquivalent:@""];
[quitItem setTarget:delegate];
[menu addItem:quitItem];

delegate.statusItem.menu = menu;
```

Service items must be inserted between the first separator (line 51) and the "Show Window" item (line 54). The final menu order becomes: header -> separator -> [5 service items] -> separator -> Show Window -> separator -> Quit.

**Disabled menu item pattern** (lines 47-48):
```objc
[header setEnabled:NO];
```

Used for not-installed services (D-08): set `[item setEnabled:NO]` and `[menu setSubmenu:nil forItem:item]`.

---

## Shared Patterns

### CGO Callback Guard
**Source:** `internal/tray/controller_darwin.go` lines 39-41
**Apply to:** All new `//export` functions (`onServiceAction`, `onRefreshStatuses`)
```go
if instance == nil || instance.app == nil {
    return
}
```

### GCD Main Queue Dispatch
**Source:** `internal/tray/tray_darwin.m` line 28
**Apply to:** All C functions called from Go that touch NSMenu/NSMenuItem
```objc
dispatch_async(dispatch_get_main_queue(), ^{
    // UI mutations here
});
```

### C String Memory Management (Go to C)
**Source:** `internal/tray/controller_darwin.go` lines 27-28
**Apply to:** `onRefreshStatuses` when passing status strings to ObjC
```go
cStatus := C.CString(status)
C.UpdateServiceStatus(C.int(i), cStatus)
C.free(unsafe.Pointer(cStatus))
```

### Service Registry Keys
**Source:** `app.go` lines 297-311, `internal/services/manager.go` line 72
**Apply to:** Service name mapping in both Go and ObjC layers

The fixed service order (D-02) with registry keys:
| Index | Registry Key | Display Name |
|-------|-------------|--------------|
| 0 | `"dnsmasq"` | DNSMasq |
| 1 | `"nginx"` | Nginx |
| 2 | `"php"` | PHP |
| 3 | `"mysql"` | MySQL |
| 4 | `"postgresql"` | PostgreSQL |

### ServiceStatus Constants
**Source:** `internal/services/service.go` lines 19-25
**Apply to:** Status comparison in ObjC `updateServiceItem:withStatus:` method
```go
StatusNotInstalled ServiceStatus = "not_installed"
StatusStopped      ServiceStatus = "stopped"
StatusStarting     ServiceStatus = "starting"
StatusRunning      ServiceStatus = "running"
StatusError        ServiceStatus = "error"
```

ObjC status comparisons must match these exact string values.

## No Analog Found

No files lack analogs. All modifications extend existing Phase 1 code in place. The patterns are fully self-contained within the tray package.

## Metadata

**Analog search scope:** `internal/tray/`, `app.go`, `internal/services/`
**Files scanned:** 7 (4 tray package files, app.go, service.go, manager.go)
**Pattern extraction date:** 2026-04-18
