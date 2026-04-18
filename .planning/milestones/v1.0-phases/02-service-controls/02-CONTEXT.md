# Phase 2: Service Controls - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Add per-service status indicators and Start/Stop/Restart actions to the existing tray menu. The tray menu built in Phase 1 (header → Show Window → Quit) is extended with a services section showing Nginx, MySQL, PHP, PostgreSQL, and dnsmasq. Each service displays its current running state and exposes a submenu for control actions. Status is refreshed each time the menu is opened.

</domain>

<decisions>
## Implementation Decisions

### Menu Layout
- **D-01:** Each service appears as a top-level menu item with a submenu arrow (▸). Service items are inserted between the header separator and the "Show Window" item.
- **D-02:** Services are listed in dependency order: dnsmasq → Nginx → PHP → MySQL → PostgreSQL. This matches the startup order in `app.go`.
- **D-03:** Each service submenu contains only three items: Start, Stop, Restart. No status text or version info in the submenu — status is conveyed by the indicator on the parent item.

### Status Indicators
- **D-04:** Running status shown as colored dot symbols next to the service name. Green filled dot (●) for running, hollow dot (○) for stopped. Use `NSAttributedString` to colorize the dots in the menu item title.
- **D-05:** Submenu actions are context-aware based on current status. If running: Start is disabled, Stop and Restart enabled. If stopped: Stop and Restart disabled, Start enabled. Prevents no-op actions.

### Action Feedback
- **D-06:** Fire-and-forget model. When the user clicks an action (Start/Stop/Restart), the menu closes naturally (macOS default behavior). The action runs in a background Go goroutine. No toast notification, no spinner.
- **D-07:** Service statuses are refreshed on every menu open using `NSMenuDelegate`'s `menuWillOpen:` callback. This calls into Go to get all statuses, then updates menu item titles (dots) and enabled/disabled state before the menu displays. No background polling.

### Not-Installed Services
- **D-08:** Services that are not installed are shown in the menu with a dash (─) and "Not Installed" label. The item is disabled (greyed out) with no submenu. All 5 services always appear in the menu for consistency.

### Claude's Discretion
- Exact `NSAttributedString` formatting for colored dots (font size, color values for green/grey)
- How the Go ↔ ObjC callback bridge is structured for service actions and status queries (expanding on the Phase 1 `onShowWindow`/`onQuit` pattern)
- Whether to use a single `RefreshMenu` C function or update items individually
- CGO callback naming conventions for service actions (e.g., `onServiceStart`, `onServiceAction`)
- Thread safety approach for concurrent service action goroutines

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Tray Package (Phase 1 Foundation)
- `internal/tray/controller.go` — `AppController` interface (currently only `Context()`), `Controller` struct, `New()` constructor. Must be expanded with service operation methods.
- `internal/tray/controller_darwin.go` — CGO bridge: `Start()`, `Destroy()`, `onShowWindow`, `onQuit` exports. Pattern for adding new Go↔ObjC callbacks.
- `internal/tray/tray_darwin.m` — ObjC implementation: `CreateTray()`, `DestroyTray()`, `LamboSystrayDelegate`. Menu construction pattern to extend with service items and submenus.
- `internal/tray/tray_darwin.h` — C header for CGO bridge functions. Add new function declarations here.

### Service Architecture
- `internal/services/service.go` — `Service` interface (`Start()`, `Stop()`, `Restart()`, `Status()`), `ServiceStatus` enum (`StatusNotInstalled`, `StatusStopped`, `StatusRunning`, etc.), `VersionedService`, `WebAdminService`.
- `app.go:257-295` — `StartService()`, `StopService()`, `RestartService()` methods on `App` struct. These delegate to `services.Manager.Get(name)`. Note D-08 pgweb dependency in `StopService`.
- `app.go:297-310` — `GetAllStatuses()` returns `map[string]string` for all registered services and web admin tools.
- `app.go:40` — `App.Manager` field (`*services.Manager`) — the service registry.

### Phase 1 Context
- `.planning/phases/01-tray-foundation-window-lifecycle/01-CONTEXT.md` — Phase 1 decisions (D-01 through D-13). Especially D-03 (GCD dispatch), D-05/D-06 (menu structure), D-09 (runtime.Show/Hide).

### Existing Architecture
- `.planning/codebase/ARCHITECTURE.md` — Full architecture analysis.
- `.planning/codebase/CONVENTIONS.md` — Consumer-defined interfaces (D-02 pattern), naming conventions.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/tray/tray_darwin.m` — Existing ObjC menu construction with `NSMenu`, `NSMenuItem`, `NSStatusItem`. Extend with `NSMenu` submenus for each service.
- `app.go:GetAllStatuses()` — Returns all service statuses as `map[string]string`. Can be called from tray via CGO callback to refresh menu.
- `app.go:StartService/StopService/RestartService` — Existing service action methods that take a service name string. Tray actions should delegate to these.
- CGO callback pattern (`//export onShowWindow`, `//export onQuit`) — Established pattern for Go functions called from ObjC. New service action callbacks follow this same pattern.

### Established Patterns
- **Consumer-defined interfaces (D-02):** `AppController` interface in `controller.go` defines only what the tray needs. Must be expanded to include service operations (`GetAllStatuses`, `StartService`, `StopService`, `RestartService`).
- **GCD dispatch (D-03):** All Cocoa UI mutations via `dispatch_async(dispatch_get_main_queue(), ...)`. Status refresh in `menuWillOpen:` must follow this.
- **Singleton instance for CGO:** `var instance *Controller` in `controller_darwin.go` routes CGO callbacks to the controller.

### Integration Points
- `internal/tray/controller.go:6-9` — `AppController` interface: add `GetAllStatuses() map[string]string`, `StartService(name string) error`, `StopService(name string) error`, `RestartService(name string) error`.
- `internal/tray/tray_darwin.m:41-71` — Menu construction in `CreateTray()`: insert service items with submenus between header separator and "Show Window".
- `internal/tray/tray_darwin.h` — Add function declarations for `RefreshMenu` and/or service action callbacks.
- `internal/tray/controller_darwin.go` — Add `//export` functions for service actions and menu refresh.

</code_context>

<specifics>
## Specific Ideas

- Menu should feel like Docker Desktop's container list — compact service names with status dots, expand for actions
- All 5 services always visible (even not-installed) so users know what's available
- No confirmation dialogs for Start/Stop/Restart from tray — user clicked intentionally, just do it
- Status refresh should feel instant when opening menu (synchronous query before menu displays)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-service-controls*
*Context gathered: 2026-04-17*
