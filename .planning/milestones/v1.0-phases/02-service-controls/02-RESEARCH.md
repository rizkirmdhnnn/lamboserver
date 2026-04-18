# Phase 2: Service Controls - Research

**Researched:** 2026-04-18
**Domain:** macOS NSMenu/NSMenuItem CGO bridge, service status display, submenu actions
**Confidence:** HIGH

## Summary

Phase 2 extends the existing Phase 1 tray menu (header, Show Window, Quit) with per-service status indicators and Start/Stop/Restart submenus. The core technical challenge is bridging Go service status queries and action dispatches through CGO into Objective-C `NSMenu` updates, using `NSMenuDelegate`'s `menuWillOpen:` callback to refresh status on every menu open.

The existing codebase already provides all the backend plumbing: `App.GetAllStatuses()` returns a `map[string]string` with status values from `services.ServiceStatus` constants, and `App.StartService/StopService/RestartService` accept service names as strings. The Phase 1 tray package establishes the CGO callback pattern (`//export` functions routing through a singleton `instance`). This phase's work is primarily in the ObjC layer (menu construction with submenus, `NSAttributedString` for colored dots, `NSMenuDelegate` for refresh) and the Go CGO bridge (new exported callbacks for service actions, a status query function).

**Primary recommendation:** Expand `AppController` interface with 4 methods (`GetAllStatuses`, `StartService`, `StopService`, `RestartService`), implement `NSMenuDelegate` on `LamboSystrayDelegate` with `menuWillOpen:` to refresh all service items, and use `NSAttributedString` with `NSForegroundColorAttributeName` for colored status dots.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Each service appears as a top-level menu item with a submenu arrow. Service items are inserted between the header separator and the "Show Window" item.
- **D-02:** Services are listed in dependency order: dnsmasq, Nginx, PHP, MySQL, PostgreSQL.
- **D-03:** Each service submenu contains only three items: Start, Stop, Restart.
- **D-04:** Running status shown as colored dot symbols. Green filled dot for running, hollow dot for stopped. Use `NSAttributedString` to colorize the dots.
- **D-05:** Submenu actions are context-aware. Running: Start disabled, Stop/Restart enabled. Stopped: Stop/Restart disabled, Start enabled.
- **D-06:** Fire-and-forget model. Action runs in background goroutine. No toast, no spinner.
- **D-07:** Statuses refreshed on every menu open via `NSMenuDelegate` `menuWillOpen:`. Synchronous query, no background polling.
- **D-08:** Not-installed services shown with dash and "Not Installed" label, disabled, no submenu.

### Claude's Discretion
- Exact `NSAttributedString` formatting (font size, color values for green/grey)
- Go-ObjC callback bridge structure for service actions and status queries
- Whether to use a single `RefreshMenu` C function or update items individually
- CGO callback naming conventions
- Thread safety approach for concurrent service action goroutines

### Deferred Ideas (OUT OF SCOPE)
None
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SRVC-01 | Tray menu shows per-service status indicators (running/stopped) for Nginx, MySQL, PHP, PostgreSQL, dnsmasq | Colored dot via `NSAttributedString` with `NSForegroundColorAttributeName`; status from `GetAllStatuses()` map; `menuWillOpen:` refresh |
| SRVC-02 | Each service has a submenu with Start, Stop, and Restart actions | `NSMenu` submenu attached to each service `NSMenuItem` via `setSubmenu:`; CGO callback bridge for actions |
| SRVC-03 | Tray menu refreshes service status after each action completes | `menuWillOpen:` delegate method queries all statuses synchronously before menu displays; fire-and-forget actions update state server-side |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Service status query | Go Backend (services.Manager) | -- | Status derived from launchd/process inspection in Go |
| Service Start/Stop/Restart | Go Backend (App methods) | -- | Delegates to service managers which manage launchd daemons |
| Menu construction with submenus | ObjC/Cocoa (tray_darwin.m) | -- | NSMenu/NSMenuItem are Cocoa APIs |
| Status indicator display | ObjC/Cocoa (tray_darwin.m) | -- | NSAttributedString for colored text is a Cocoa API |
| Menu refresh on open | ObjC/Cocoa (NSMenuDelegate) | Go Backend (status query) | Delegate triggers ObjC callback which calls into Go for fresh data |
| CGO bridge | Go CGO layer (controller_darwin.go) | ObjC header (tray_darwin.h) | Connects Go callbacks to ObjC and vice versa |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| AppKit (NSMenu, NSMenuItem, NSAttributedString) | macOS system | Menu construction, styled text, submenu hierarchy | Native macOS API, already used in Phase 1 |
| NSMenuDelegate | macOS system | `menuWillOpen:` callback for dynamic menu refresh | Official Apple pattern for pre-display menu updates |
| CGO | Go stdlib | Bridge between Go and Objective-C | Already established in Phase 1 tray package |

No new external dependencies are needed. This phase uses only existing macOS system frameworks and the established CGO bridge pattern.

## Architecture Patterns

### System Architecture Diagram

```
User clicks tray icon
        |
        v
[NSStatusItem menu opens]
        |
        v
[NSMenuDelegate menuWillOpen:]  -- calls --> [C function: RefreshServiceStatuses()]
        |                                              |
        |                                              v
        |                                    [Go: onRefreshStatuses() //export]
        |                                              |
        |                                              v
        |                                    [instance.app.GetAllStatuses()]
        |                                              |
        |                                              v
        |                                    [Returns status map to C as arrays]
        |                                              |
        v                                              v
[ObjC updates NSMenuItem attributedTitle with colored dots]
[ObjC enables/disables submenu items based on status]
        |
        v
[Menu displays with current statuses]

User clicks submenu action (e.g., "Start" on nginx)
        |
        v
[ObjC action selector] -- calls --> [C function: e.g., ServiceAction(name, action)]
        |                                              |
        v                                              v
[Menu closes (macOS default)]          [Go: onServiceAction(name, action) //export]
                                                       |
                                                       v
                                              [go func() { instance.app.StartService(name) }()]
                                              (fire-and-forget goroutine)
```

### Recommended Project Structure

No new files beyond existing tray package. All changes are to existing files:

```
internal/tray/
  controller.go          # Expand AppController interface (+4 methods)
  controller_darwin.go   # Add //export callbacks for service actions + status refresh
  tray_darwin.h          # Add C function declarations
  tray_darwin.m          # Add service menu items, submenus, NSMenuDelegate, refresh logic
  tray_unsupported.go    # No changes needed (stubs remain)
```

### Pattern 1: NSMenuDelegate for Dynamic Refresh

**What:** Implement `NSMenuDelegate` protocol on `LamboSystrayDelegate` to intercept menu open events and refresh service statuses before display.

**When to use:** Every time the user clicks the tray icon to open the menu.

**Example:**
```objc
// Source: Apple Developer Documentation - NSMenuDelegate
// https://developer.apple.com/documentation/appkit/nsmenudelegate

@interface LamboSystrayDelegate : NSObject <NSMenuDelegate>
@property (strong, nonatomic) NSStatusItem *statusItem;
@property (strong, nonatomic) NSMutableArray<NSMenuItem *> *serviceItems;
@property (strong, nonatomic) NSMutableArray<NSMenu *> *serviceSubmenus;
@end

@implementation LamboSystrayDelegate

- (void)menuWillOpen:(NSMenu *)menu {
    // Call into Go to get fresh statuses, then update menu items
    RefreshServiceStatuses();
}

@end
```

[VERIFIED: Apple Developer Documentation] `menuWillOpen:` is called just before the menu is displayed. The delegate must be strongly referenced (it already is -- `static LamboSystrayDelegate *delegate` in the existing code).

### Pattern 2: NSAttributedString for Colored Status Dots

**What:** Use `NSAttributedString` with `NSForegroundColorAttributeName` to render colored circle symbols in menu item titles.

**When to use:** For each service menu item to show running (green dot) or stopped (grey dot) state.

**Example:**
```objc
// Source: Apple Developer Documentation - NSMenuItem attributedTitle
// https://developer.apple.com/documentation/appkit/nsmenuitem/attributedtitle

- (NSAttributedString *)attributedTitleForService:(NSString *)name
                                           status:(NSString *)status {
    NSString *dot;
    NSColor *dotColor;

    if ([status isEqualToString:@"running"]) {
        dot = @"\u25CF ";  // ● filled circle
        dotColor = [NSColor systemGreenColor];
    } else if ([status isEqualToString:@"stopped"]) {
        dot = @"\u25CB ";  // ○ hollow circle
        dotColor = [NSColor secondaryLabelColor];
    } else {
        // not_installed handled separately (D-08)
        dot = @"\u2500 ";  // ─ dash
        dotColor = [NSColor tertiaryLabelColor];
    }

    NSMutableAttributedString *result = [[NSMutableAttributedString alloc] init];

    // Colored dot portion
    NSDictionary *dotAttrs = @{
        NSForegroundColorAttributeName: dotColor,
        NSFontAttributeName: [NSFont menuFontOfSize:14.0]
    };
    [result appendAttributedString:[[NSAttributedString alloc]
        initWithString:dot attributes:dotAttrs]];

    // Service name portion (default color -- adapts to selection/disabled state)
    NSDictionary *nameAttrs = @{
        NSFontAttributeName: [NSFont menuFontOfSize:14.0]
    };
    [result appendAttributedString:[[NSAttributedString alloc]
        initWithString:name attributes:nameAttrs]];

    return result;
}
```

[CITED: https://developer.apple.com/documentation/appkit/nsmenuitem/attributedtitle] `attributedTitle` allows styled text on menu items. When no `NSForegroundColorAttributeName` is set on the name portion, it automatically adapts to black/white/grey for normal/selected/disabled states.

### Pattern 3: CGO Callback Bridge for Service Actions

**What:** Extend the existing `//export` pattern to route service action requests from ObjC to Go.

**When to use:** When user clicks Start/Stop/Restart in a service submenu.

**Example:**
```go
// controller_darwin.go

//export onServiceAction
func onServiceAction(cName *C.char, cAction *C.char) {
    if instance == nil || instance.app == nil {
        return
    }
    name := C.GoString(cName)
    action := C.GoString(cAction)
    // Fire-and-forget per D-06
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
```

[VERIFIED: existing codebase] This follows the exact same pattern as `onShowWindow` and `onQuit` in `controller_darwin.go`, extended with C string parameters.

### Pattern 4: Submenu Construction

**What:** Create an `NSMenu` submenu for each service with Start/Stop/Restart items, attached to the service's parent menu item.

**Example:**
```objc
// Source: Apple Developer Documentation - NSMenu setSubmenu:forItem:
// https://developer.apple.com/documentation/appkit/nsmenu/1518194-setsubmenu

// For each service in the fixed order:
NSMenuItem *serviceItem = [[NSMenuItem alloc] initWithTitle:@""
                                                    action:nil
                                             keyEquivalent:@""];

NSMenu *submenu = [[NSMenu alloc] init];

NSMenuItem *startItem = [[NSMenuItem alloc] initWithTitle:@"Start"
                                                   action:@selector(serviceStart:)
                                            keyEquivalent:@""];
[startItem setTarget:delegate];
[startItem setRepresentedObject:@"nginx"];  // service name for callback
[submenu addItem:startItem];

// ... Stop and Restart items similarly

[menu setSubmenu:submenu forItem:serviceItem];
```

[CITED: https://developer.apple.com/documentation/appkit/nsmenu/1518194-setsubmenu] `setSubmenu:forItem:` associates a submenu with a menu item, causing the arrow indicator to appear.

### Anti-Patterns to Avoid

- **Polling for status changes:** Do NOT use a timer to periodically refresh statuses. D-07 explicitly requires refresh only on menu open via `menuWillOpen:`. Polling wastes resources and creates thread safety issues.
- **Blocking the main thread during status query:** The `menuWillOpen:` callback runs on the main thread. The Go `GetAllStatuses()` call queries launchd/process state. If any service status check is slow, it could freeze the menu. Keep the Go-side status checks fast (they already are -- they inspect launchd plist existence and `pgrep`). [ASSUMED]
- **Creating new menu items on every refresh:** Do NOT tear down and rebuild the entire menu on each `menuWillOpen:`. Instead, store references to the service `NSMenuItem` objects and update their `attributedTitle` and submenu item `enabled` state in place. This avoids flicker and is more efficient.
- **Forgetting GCD dispatch for UI updates:** All `NSMenuItem` mutations must happen on the main thread. Since `menuWillOpen:` already runs on the main thread, the refresh logic is safe. But if any Go callback tries to update menu items directly, it must dispatch to the main queue.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Colored text in menu items | Custom view (`NSView`) for menu item | `NSAttributedString` with `attributedTitle` | Custom views break native hover highlighting, keyboard navigation, and accessibility |
| Dynamic menu content | Remove/re-add all menu items on each open | Update existing items' `attributedTitle` and `enabled` state | In-place updates avoid flicker and are more performant |
| Service status model | New status tracking in tray package | `App.GetAllStatuses()` via `AppController` interface | Backend already queries launchd/process state; tray is a view layer |
| Service action dispatch | Direct service manager calls from tray | `App.StartService/StopService/RestartService` via `AppController` | These methods handle D-08 dependencies (pgweb stops with PostgreSQL) and logging |

**Key insight:** The tray package is purely a presentation layer. All service logic already exists in `app.go` methods. The tray should delegate everything through the `AppController` interface, never importing service packages directly.

## Common Pitfalls

### Pitfall 1: Service Name Mismatch Between Menu and Registry

**What goes wrong:** The tray hardcodes service display names but uses different strings when calling `StartService()`, causing "service not found" errors.
**Why it happens:** The service registry uses lowercase keys (`"nginx"`, `"dnsmasq"`, `"php"`, `"mysql"`, `"postgresql"`) while display names are capitalized.
**How to avoid:** Define a fixed service descriptor array in the ObjC code mapping display name to registry key. The 5 services per D-02 are: `dnsmasq`/`"DNSMasq"`, `nginx`/`"Nginx"`, `php`/`"PHP"`, `mysql`/`"MySQL"`, `postgresql`/`"PostgreSQL"`.
**Warning signs:** "service not found" errors in debug log after clicking tray actions.

### Pitfall 2: menuWillOpen: Not Firing

**What goes wrong:** The delegate's `menuWillOpen:` never gets called, so statuses are stale.
**Why it happens:** Forgetting to set `menu.delegate = delegate` on the `NSMenu` instance, or the delegate not conforming to `<NSMenuDelegate>` protocol.
**How to avoid:** Explicitly set `menu.delegate = delegate` in `CreateTray()` and add `<NSMenuDelegate>` to the `@interface` declaration.
**Warning signs:** Menu always shows the same status dots regardless of actual service state.

### Pitfall 3: Thread Safety on Concurrent Service Actions

**What goes wrong:** User rapidly clicks Start on multiple services; concurrent goroutines cause race conditions.
**Why it happens:** Fire-and-forget goroutines (D-06) run concurrently. The underlying service managers may not be designed for concurrent calls.
**How to avoid:** Each `App` method (`StartService`, etc.) delegates to individual service managers which operate independently. The `services.Manager` registry is read-only after startup. Concurrent operations on different services are safe. Concurrent operations on the same service would be a user error (menu closes after first click). No additional synchronization needed. [ASSUMED]
**Warning signs:** Corrupted launchd state or double-start errors in logs.

### Pitfall 4: C String Memory Leaks in CGO Bridge

**What goes wrong:** C strings passed from ObjC to Go via `onServiceAction(name, action)` leak memory.
**Why it happens:** If Go allocates C strings with `C.CString()`, they must be freed. If ObjC passes string literals or `NSString.UTF8String`, the memory is managed by ObjC autorelease pool.
**How to avoid:** When passing from ObjC to Go, use `[nsString UTF8String]` which is autoreleased. When passing from Go to ObjC (for status refresh results), use `C.CString()` in Go and `free()` in ObjC after use, or use Go-allocated arrays that ObjC copies from.
**Warning signs:** Growing memory usage over time as menu is repeatedly opened.

### Pitfall 5: Not-Installed Service with Submenu

**What goes wrong:** A not-installed service still shows a submenu with Start/Stop/Restart, which would fail.
**Why it happens:** Submenu is created at init time for all services and never removed.
**How to avoid:** Per D-08, not-installed services should have no submenu at all. In `menuWillOpen:`, check status and either attach or detach the submenu. Alternatively, create the submenu only for installed services and set the item to disabled with no submenu for not-installed ones.
**Warning signs:** Clicking "Start" on a not-installed service produces an error.

## Code Examples

### Expanding AppController Interface

```go
// controller.go - Source: existing codebase pattern

// AppController is the narrow interface the tray needs from the application.
// Follows consumer-defined interface convention (D-02).
type AppController interface {
    // Context returns the Wails runtime context for Show/Hide/Quit calls.
    Context() context.Context
    // GetAllStatuses returns status strings keyed by service name.
    GetAllStatuses() map[string]string
    // StartService starts the named service.
    StartService(name string) error
    // StopService stops the named service.
    StopService(name string) error
    // RestartService restarts the named service.
    RestartService(name string) error
}
```

[VERIFIED: existing codebase] `App` struct already implements all 4 new methods (`GetAllStatuses`, `StartService`, `StopService`, `RestartService` in `app.go:257-311`). Adding them to `AppController` requires no new backend code.

### Status Data Transfer from Go to ObjC

```go
// controller_darwin.go - passing status data back to ObjC

//export onRefreshStatuses
func onRefreshStatuses() {
    if instance == nil || instance.app == nil {
        return
    }
    statuses := instance.app.GetAllStatuses()

    // Fixed order per D-02
    serviceOrder := []string{"dnsmasq", "nginx", "php", "mysql", "postgresql"}
    for i, name := range serviceOrder {
        status := statuses[name]
        if status == "" {
            status = "not_installed"
        }
        cStatus := C.CString(status)
        C.UpdateServiceStatus(C.int(i), cStatus)
        C.free(unsafe.Pointer(cStatus))
    }
}
```

```c
// tray_darwin.h
void UpdateServiceStatus(int index, const char *status);
void RefreshServiceStatuses(void);
```

```objc
// tray_darwin.m

// Called from Go for each service during refresh
void UpdateServiceStatus(int index, const char *status) {
    // This runs on whatever thread Go calls from.
    // Must dispatch to main for UI updates.
    NSString *statusStr = [NSString stringWithUTF8String:status];
    int idx = index;
    dispatch_async(dispatch_get_main_queue(), ^{
        [delegate updateServiceItem:idx withStatus:statusStr];
    });
}
```

### Complete Menu Construction with Services

```objc
// tray_darwin.m - service menu setup (inside CreateTray, after header separator)

// Service descriptors: registry key, display name
static NSString *serviceKeys[] = {@"dnsmasq", @"nginx", @"php", @"mysql", @"postgresql"};
static NSString *serviceNames[] = {@"DNSMasq", @"Nginx", @"PHP", @"MySQL", @"PostgreSQL"};
static const int kServiceCount = 5;

// In CreateTray, after the header separator:
delegate.serviceItems = [NSMutableArray arrayWithCapacity:kServiceCount];
delegate.serviceSubmenus = [NSMutableArray arrayWithCapacity:kServiceCount];

for (int i = 0; i < kServiceCount; i++) {
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:serviceNames[i]
                                                 action:nil
                                          keyEquivalent:@""];

    // Create submenu with Start/Stop/Restart
    NSMenu *sub = [[NSMenu alloc] init];
    NSArray *actions = @[@"Start", @"Stop", @"Restart"];
    SEL selectors[] = {@selector(serviceStart:), @selector(serviceStop:),
                       @selector(serviceRestart:)};

    for (int j = 0; j < 3; j++) {
        NSMenuItem *actionItem = [[NSMenuItem alloc] initWithTitle:actions[j]
                                                            action:selectors[j]
                                                     keyEquivalent:@""];
        [actionItem setTarget:delegate];
        [actionItem setTag:i];  // identify which service
        [sub addItem:actionItem];
    }

    [delegate.serviceItems addObject:item];
    [delegate.serviceSubmenus addObject:sub];

    [menu addItem:item];
    [menu setSubmenu:sub forItem:item];
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `NSStatusItem` custom view for rich menus | `NSStatusItem.menu` with `NSAttributedString` titles | macOS 10.12+ | Custom views deprecated in favor of attributed strings for standard menus |
| Manual polling for status | `NSMenuDelegate menuWillOpen:` | Long-standing | Pull-on-demand is the standard macOS pattern |
| `[NSColor greenColor]` | `[NSColor systemGreenColor]` | macOS 10.10+ | System colors adapt to dark mode and accessibility settings |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `GetAllStatuses()` is fast enough to call synchronously in `menuWillOpen:` without perceptible delay | Anti-Patterns / Pitfall 2 | Menu would freeze briefly on open; would need async prefetch |
| A2 | Concurrent service actions on different services are safe (no shared mutable state between service managers) | Pitfall 3 | Race conditions; would need a mutex or serialization queue |
| A3 | `[NSColor systemGreenColor]` renders visibly on both light and dark macOS menu bar themes | Code Examples | Green dot might be invisible on certain themes; would need color testing |

## Open Questions

1. **menuWillOpen: synchronous Go call timing**
   - What we know: `menuWillOpen:` runs on the main thread. The Go status query involves `pgrep`/launchd checks.
   - What's unclear: Whether the Go call blocks the ObjC main thread until completion, or if CGO manages thread handoff.
   - Recommendation: Test empirically. If blocking is perceptible, consider pre-caching statuses and refreshing asynchronously with a brief delay before showing.

2. **Submenu attachment/detachment for not-installed services**
   - What we know: D-08 says no submenu for not-installed services.
   - What's unclear: Whether `setSubmenu:nil forItem:item` cleanly removes a previously attached submenu, or if items need to be recreated.
   - Recommendation: Set `[menu setSubmenu:nil forItem:item]` and test. If it doesn't work, hide the submenu items or create items without submenus for not-installed services.

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/tray/` package (Phase 1 implementation) - CGO bridge pattern, menu construction, singleton callback routing
- Existing codebase: `internal/services/service.go` - Service/ServiceStatus interfaces and constants
- Existing codebase: `app.go:257-311` - GetAllStatuses, StartService, StopService, RestartService implementations
- Existing codebase: `internal/services/manager.go` - Service registry with name-based lookup

### Secondary (MEDIUM confidence)
- [Apple Developer Documentation - NSMenuDelegate](https://developer.apple.com/documentation/appkit/nsmenudelegate) - `menuWillOpen:` protocol method
- [Apple Developer Documentation - NSMenuItem attributedTitle](https://developer.apple.com/documentation/appkit/nsmenuitem/attributedtitle) - Styled text on menu items
- [Apple Developer Documentation - NSMenu setSubmenu:forItem:](https://developer.apple.com/documentation/appkit/nsmenu/1518194-setsubmenu) - Submenu attachment

### Tertiary (LOW confidence)
- Web search results on NSAttributedString color behavior in menus (multiple sources agree on behavior)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Uses only existing macOS system frameworks already in use
- Architecture: HIGH - Extends established Phase 1 CGO bridge pattern with well-understood additions
- Pitfalls: MEDIUM - Thread timing and memory management in CGO bridges require empirical testing

**Research date:** 2026-04-18
**Valid until:** 2026-05-18 (stable macOS APIs, no rapid changes expected)
