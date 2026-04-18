# Phase 3: Quick Access Links - Context

**Gathered:** 2026-04-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a "Quick Access" submenu to the existing tray menu that lists configured local sites (for one-click browser launch) and web admin tool links (phpMyAdmin, pgweb). Sites are dynamically populated on each menu open. Web admin items appear only when the tool is installed.

</domain>

<decisions>
## Implementation Decisions

### Menu Layout
- **D-01:** Sites and web admin links live inside a single "Quick Access" parent submenu in the main tray menu. This keeps the main menu compact — one extra item instead of up to 17.
- **D-02:** The "Quick Access" submenu is inserted between the services section and "Show Window" in the main menu.
- **D-03:** Clicking a site item opens `https://{domain}` (or `http://` if SSL not enabled) in the default browser using `browser.OpenURL`.

### Site Display
- **D-04:** Sites sorted by most recently created first (newest at top). Uses `CreatedAt` field from `sites.Site`.
- **D-05:** Site list capped at 15 entries. If more than 15 sites exist, show the first 15 followed by a disabled menu item: "(+N more — see main window)".
- **D-06:** Claude has discretion on site label format (domain only vs domain + path hint).

### Web Admin Tools
- **D-07:** phpMyAdmin and pgweb items are completely hidden when not installed. They only appear in the submenu when `IsInstalled()` returns true. This matches QKAC-02/QKAC-03 requirements.
- **D-08:** Web admin items use "Open phpMyAdmin" / "Open pgweb" labels (matching existing `app.go` naming). Clicking opens the tool's URL via `browser.OpenURL`.

### Empty State
- **D-09:** The "Quick Access" submenu is always present in the main menu, even when empty. When no sites are configured and no web admin tools are installed, show a single disabled item: "No sites configured".
- **D-10:** Claude has discretion on whether to use a separator between the sites list and web admin tool items inside the submenu.

### Refresh Behavior
- **D-11:** Sites list and web admin install status are refreshed on every menu open, using the same `NSMenuDelegate` `menuWillOpen:` pattern established in Phase 2 (D-07). No background polling.

### Claude's Discretion
- Site label format (domain only vs domain + path)
- Separator placement between sites and web admin tools inside the Quick Access submenu
- CGO callback naming for sites/web admin queries
- Whether to build site menu items in ObjC or pass pre-formatted data from Go

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Tray Package (Phase 1+2 Foundation)
- `internal/tray/controller.go` — `AppController` interface. Must be expanded with site listing and web admin query methods.
- `internal/tray/controller_darwin.go` — CGO bridge with `//export` callbacks. Pattern for adding new Go→ObjC data passing.
- `internal/tray/tray_darwin.m` — ObjC menu construction. `menuWillOpen:` refresh pattern. Service items/submenus pattern to replicate for Quick Access.
- `internal/tray/tray_darwin.h` — C header for bridge functions.

### Site Data
- `internal/sites/manager.go` — `Site` struct: `Domain`, `Path`, `PhpVersion`, `SSLEnabled`, `CreatedAt` fields.
- `app.go:670` — `GetSites() []sites.Site` — returns all configured sites.
- `internal/config/store.go:150` — `GetSites()` returns thread-safe copy of site configs.

### Web Admin Tools
- `internal/services/service.go:72-88` — `WebAdminService` interface: `IsInstalled()`, `URL()`, `Version()`.
- `internal/services/manager.go:86` — `AllWebAdmin()` returns all registered web admin services.
- `app.go:442-456` — `OpenWebAdmin(name)` uses `GetWebAdmin` registry + `browser.OpenURL`.
- `app.go:524-527` — `OpenPgweb()` hardcoded URL. Consider using `OpenWebAdmin` pattern instead.

### Browser Integration
- `github.com/pkg/browser` — Already imported and used in `app.go`. `browser.OpenURL(url)` opens default browser.

### Phase 1+2 Context
- `.planning/phases/01-tray-foundation-window-lifecycle/01-CONTEXT.md` — D-03 (GCD dispatch), D-05/D-06 (menu structure).
- `.planning/phases/02-service-controls/02-CONTEXT.md` — D-01 (menu item placement), D-07 (menuWillOpen refresh pattern).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `App.GetSites()` — Already returns all sites as `[]sites.Site`. Can be called from tray via CGO callback.
- `App.OpenWebAdmin(name)` — Generic web admin opener using the service registry. Can be called from tray for both phpMyAdmin and pgweb.
- `services.Manager.AllWebAdmin()` — Returns `map[string]WebAdminService`. Can query `IsInstalled()` and `URL()` for each.
- `browser.OpenURL()` — Already used throughout `app.go` for opening URLs.
- `NSMenuDelegate` `menuWillOpen:` — Phase 2 established the pattern of refreshing menu state on every open. Sites and web admin status can be refreshed the same way.

### Established Patterns
- **Consumer-defined interfaces (D-02):** `AppController` in `controller.go` defines only what tray needs. Must expand with `GetSites()` and web admin query methods.
- **CGO callback pattern:** `//export onServiceAction` etc. in `controller_darwin.go`. New callbacks for opening sites/web admin follow the same pattern.
- **GCD dispatch (Phase 1 D-03):** All Cocoa UI mutations via `dispatch_async(dispatch_get_main_queue(), ...)`.

### Integration Points
- `internal/tray/controller.go` — Add to `AppController`: `GetSites()`, `OpenSiteInBrowser(domain)`, `GetWebAdminStatus()`, `OpenWebAdmin(name)`.
- `internal/tray/tray_darwin.m` — Add Quick Access submenu construction in `CreateTray()` or `menuWillOpen:`.
- `internal/tray/tray_darwin.h` — Add C function declarations for site/web admin data passing.
- `internal/tray/controller_darwin.go` — Add `//export` callbacks for site open and web admin open actions.

</code_context>

<specifics>
## Specific Ideas

- Follow the "Quick Access" submenu pattern — keeps main menu clean as the app grows
- Web admin items hidden, not greyed out, when not installed — users shouldn't see tools they can't use
- Refresh sites on every menu open — user may add/remove sites while app is running
- The 15-site cap with "(+N more — see main window)" prevents the menu from becoming unwieldy

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-quick-access-links*
*Context gathered: 2026-04-18*
