# Phase 03: Quick Access Links - Research

**Researched:** 2026-04-18
**Domain:** macOS ObjC/CGO tray menu construction, Go-to-ObjC data passing, site and web admin service queries
**Confidence:** HIGH

## Summary

Phase 3 adds a "Quick Access" submenu to the existing tray menu. Users can open configured local sites and web admin tools (phpMyAdmin, pgweb) directly from the tray with a single click. The implementation follows a well-established pattern from Phases 1 and 2: CGO callbacks from ObjC back into Go, data passing from Go into ObjC via C strings, and menu refresh in `menuWillOpen:`.

The codebase already contains all the Go-side building blocks: `App.GetSites()` returns sorted site data, `App.OpenWebAdmin("phpmyadmin")` opens phpMyAdmin via the service registry, and `pgweb.Manager.IsInstalled()` / `pgweb.Manager.URL()` provide pgweb status and URL. The only work is wiring these into the tray via new `AppController` methods, new `//export` callbacks, new C bridge functions, and new ObjC submenu construction.

One key asymmetry to handle: pgweb is NOT registered in the `services.Manager` WebAdmin registry — it has a dedicated `*pgweb.Manager` on `App`. This means tray needs a separate `GetPgwebStatus() (installed bool, url string)` method (or a combined `GetWebAdminItems()` method that bundles both tools), rather than relying on `Manager.AllWebAdmin()` alone.

**Primary recommendation:** Expand `AppController` with `GetSites()`, `OpenSiteInBrowser(domain string)`, `GetWebAdminItems() []WebAdminItem`, and `OpenWebAdmin(name string)`. Build the Quick Access submenu in ObjC, refreshing it entirely in `menuWillOpen:` using the same synchronous pattern as Phase 2.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Sites and web admin links live inside a single "Quick Access" parent submenu in the main tray menu.
- **D-02:** The "Quick Access" submenu is inserted between the services section and "Show Window" in the main menu.
- **D-03:** Clicking a site item opens `https://{domain}` (or `http://` if SSL not enabled) in the default browser using `browser.OpenURL`.
- **D-04:** Sites sorted by most recently created first (newest at top). Uses `CreatedAt` field from `sites.Site`.
- **D-05:** Site list capped at 15 entries. If more than 15 sites exist, show the first 15 followed by a disabled menu item: "(+N more — see main window)".
- **D-07:** phpMyAdmin and pgweb items are completely hidden when not installed. They only appear in the submenu when `IsInstalled()` returns true.
- **D-08:** Web admin items use "Open phpMyAdmin" / "Open pgweb" labels. Clicking opens the tool's URL via `browser.OpenURL`.
- **D-09:** The "Quick Access" submenu is always present in the main menu, even when empty. When no sites are configured and no web admin tools are installed, show a single disabled item: "No sites configured".
- **D-11:** Sites list and web admin install status are refreshed on every menu open, using the `NSMenuDelegate` `menuWillOpen:` pattern from Phase 2. No background polling.

### Claude's Discretion

- Site label format (domain only vs domain + path)
- Separator placement between sites and web admin tools inside the Quick Access submenu
- CGO callback naming for sites/web admin queries
- Whether to build site menu items in ObjC or pass pre-formatted data from Go

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QKAC-01 | Tray menu lists configured sites with "Open in Browser" action (capped at 15 entries) | `App.GetSites()` returns `[]sites.Site` with `Domain`, `SSLEnabled`, `CreatedAt`. Sort by `CreatedAt` descending (newest first per D-04), slice to 15. CGO callback `onOpenSite(domain)` calls `App.OpenSiteInBrowser(domain)` which constructs URL from domain + SSLEnabled and calls `browser.OpenURL`. |
| QKAC-02 | Tray menu has "Open phpMyAdmin" item (shown only when phpMyAdmin is installed) | `phpmyadmin.Manager.IsInstalled()` checks for `index.php` sentinel file. Registered in `services.Manager` as `"phpmyadmin"`. Accessible via `App.OpenWebAdmin("phpmyadmin")`. URL is `"https://phpmyadmin.test"`. |
| QKAC-03 | Tray menu has "Open pgweb" item (shown only when pgweb is installed) | `pgweb.Manager.IsInstalled()` checks binary existence. NOT in WebAdmin registry — uses `App.Pgweb.IsInstalled()` and `App.Pgweb.URL()` directly. Must be exposed via a new `AppController` method (e.g., `GetWebAdminItems()`). |
</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Site data query (list + sort) | Go (App method) | — | Sites live in `config.Store`; all data access goes through `App.GetSites()` |
| URL construction from domain | Go (AppController method) | — | SSL flag determines scheme; logic belongs in Go, not ObjC string concatenation |
| Browser launch | Go (browser.OpenURL) | — | `github.com/pkg/browser` is already wired; ObjC just fires the callback |
| phpMyAdmin install check | Go (phpmyadmin.Manager) | — | Filesystem check; exposed via `AppController` |
| pgweb install check | Go (pgweb.Manager) | — | Binary existence check; exposed via `AppController` |
| Menu construction + refresh | ObjC (LamboSystrayDelegate) | — | Follows established Phase 2 pattern; all NSMenu manipulation in ObjC |
| Menu item actions (open site/tool) | ObjC trigger → Go callback | — | ObjC fires `//export` callback, Go executes browser launch |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/pkg/browser` | v0.0.0-20240102092130 | Open URLs in default browser | Already imported and used in `app.go`; handles macOS `open` command |
| Cocoa / NSMenu | macOS SDK | ObjC menu item construction | Established in Phase 1+2; no alternatives within this CGO architecture |
| CGO (Go standard toolchain) | go 1.25.0 | Go-ObjC bridge | Locked by project constraint (TRAY-03) |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `sort` (Go stdlib) | stdlib | Sort sites by `CreatedAt` descending | Needed for D-04; no external dependency required |

**Installation:** No new dependencies required. [VERIFIED: codebase grep of go.mod and app.go imports]

---

## Architecture Patterns

### System Architecture Diagram

```
menuWillOpen: (main thread, ObjC)
       |
       v
onRefreshQuickAccess() [//export, Go]
       |
       +---> instance.app.GetSites()          --> []SiteInfo (domain, ssl, label)
       |
       +---> instance.app.GetWebAdminItems()  --> []WebAdminItem (name, label, installed, url)
       |
       v
RebuildQuickAccessSubmenu(sites[], count, webAdmins[], count) [C, ObjC]
       |
       +---> Clears existing items in quickAccessSubmenu
       +---> Adds site items (tag = index) up to 15
       +---> Adds "(+N more)" disabled item if overflow
       +---> Adds separator (discretion)
       +---> Adds web admin items for installed tools only
       +---> Falls back to "No sites configured" if empty

User clicks site item:
  ObjC -[openSite:] --> onOpenSite(domain) [//export, Go] --> browser.OpenURL(url)

User clicks web admin item:
  ObjC -[openWebAdmin:] --> onOpenWebAdmin(name) [//export, Go] --> browser.OpenURL(url)
```

### Recommended Project Structure

No new packages or files are introduced. Changes are confined to the existing tray package:

```
internal/tray/
├── controller.go          # AppController interface — add new methods
├── controller_darwin.go   # CGO bridge — add new //export callbacks + C calls
├── tray_darwin.h          # C header — add new function declarations
└── tray_darwin.m          # ObjC implementation — add Quick Access submenu
```

### Pattern 1: Expanding AppController (consumer-defined interface)

**What:** Add the minimum methods the tray package needs from the application.
**When to use:** Any time tray needs new data from the app — never import app packages directly.

```go
// Source: internal/tray/controller.go (established pattern)
type AppController interface {
    Context() context.Context
    GetAllStatuses() map[string]string
    StartService(name string) error
    StopService(name string) error
    RestartService(name string) error
    // New for Phase 3:
    GetSites() []SiteInfo            // Returns sorted, capped site data for the tray
    OpenSiteInBrowser(domain string) // Constructs URL from domain, calls browser.OpenURL
    GetWebAdminItems() []WebAdminItem // Returns install status + URL for phpMyAdmin and pgweb
    OpenWebAdmin(name string)        // Opens named web admin tool in browser
}
```

Define lightweight value types in `controller.go` (not importing `internal/sites`):

```go
// SiteInfo is the tray-facing site descriptor. Domain-only label is the default
// discretion choice — keeps items compact in the menu.
type SiteInfo struct {
    Domain string
    URL    string  // Pre-computed by Go: https://{domain} or http://{domain}
    Label  string  // Display string, e.g. "myapp.test"
}

// WebAdminItem is the tray-facing descriptor for a web admin tool.
type WebAdminItem struct {
    Name      string // Registry key: "phpmyadmin" or "pgweb"
    Label     string // Display: "Open phpMyAdmin" / "Open pgweb"
    Installed bool
    URL       string
}
```

[VERIFIED: controller.go — existing interface pattern confirmed by codebase read]

### Pattern 2: New //export callbacks in controller_darwin.go

**What:** Go functions exported to C for ObjC to call when user clicks a menu item.
**When to use:** Every user action in ObjC that requires a Go-side effect.

```go
// Source: internal/tray/controller_darwin.go (established //export pattern)

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
        instance.app.OpenWebAdmin(name)
    }()
}

//export onRefreshQuickAccess
func onRefreshQuickAccess() {
    if instance == nil || instance.app == nil {
        return
    }
    sites := instance.app.GetSites()
    for i, s := range sites {
        cDomain := C.CString(s.Domain)
        cURL := C.CString(s.URL)
        cLabel := C.CString(s.Label)
        C.AddQuickAccessSite(C.int(i), cDomain, cURL, cLabel)
        C.free(unsafe.Pointer(cDomain))
        C.free(unsafe.Pointer(cURL))
        C.free(unsafe.Pointer(cLabel))
    }
    // ... similarly for web admin items
}
```

[VERIFIED: controller_darwin.go — onServiceAction, onRefreshStatuses pattern confirmed]

### Pattern 3: menuWillOpen: refresh for Quick Access

**What:** ObjC delegate method triggers data refresh before submenu displays.
**When to use:** Any data that may change while the app is running (sites, install state).

```objc
// Source: tray_darwin.m (established NSMenuDelegate pattern from Phase 2)
- (void)menuWillOpen:(NSMenu *)menu {
    onRefreshStatuses();        // existing
    onRefreshQuickAccess();     // new: rebuild Quick Access submenu
}
```

The synchronous call chain (main thread → Go → C → ObjC) ensures all items are rebuilt before the menu renders. [VERIFIED: tray_darwin.m menuWillOpen: pattern confirmed]

### Pattern 4: Building site items in ObjC

**What:** ObjC creates NSMenuItem objects for each site using data passed from Go.
**Discretion decision:** Build items in ObjC (not Go). Go passes pre-formatted strings via C functions; ObjC constructs NSMenuItems. This avoids passing NSMenuItem pointers through CGO.

```objc
// In LamboSystrayDelegate
- (void)rebuildQuickAccessSubmenu {
    [self.quickAccessSubmenu removeAllItems];

    if (self.quickAccessSites.count == 0 && !self.phpMyAdminInstalled && !self.pgwebInstalled) {
        NSMenuItem *empty = [[NSMenuItem alloc]
            initWithTitle:@"No sites configured" action:nil keyEquivalent:@""];
        [empty setEnabled:NO];
        [self.quickAccessSubmenu addItem:empty];
        return;
    }

    for (int i = 0; i < (int)self.quickAccessSites.count; i++) {
        NSDictionary *site = self.quickAccessSites[i];
        NSMenuItem *item = [[NSMenuItem alloc]
            initWithTitle:site[@"label"]
                   action:@selector(openSite:)
            keyEquivalent:@""];
        [item setTarget:self];
        [item setTag:i];
        [self.quickAccessSubmenu addItem:item];
    }

    if (self.quickAccessOverflow > 0) {
        NSString *overflowTitle = [NSString
            stringWithFormat:@"(+%d more \u2014 see main window)", self.quickAccessOverflow];
        NSMenuItem *overflow = [[NSMenuItem alloc]
            initWithTitle:overflowTitle action:nil keyEquivalent:@""];
        [overflow setEnabled:NO];
        [self.quickAccessSubmenu addItem:overflow];
    }

    // Separator between sites and web admin tools (discretion: include separator)
    if (self.phpMyAdminInstalled || self.pgwebInstalled) {
        [self.quickAccessSubmenu addItem:[NSMenuItem separatorItem]];
    }

    if (self.phpMyAdminInstalled) {
        NSMenuItem *pma = [[NSMenuItem alloc]
            initWithTitle:@"Open phpMyAdmin"
                   action:@selector(openWebAdmin:)
            keyEquivalent:@""];
        [pma setTarget:self];
        [pma setRepresentedObject:@"phpmyadmin"];
        [self.quickAccessSubmenu addItem:pma];
    }

    if (self.pgwebInstalled) {
        NSMenuItem *pgweb = [[NSMenuItem alloc]
            initWithTitle:@"Open pgweb"
                   action:@selector(openWebAdmin:)
            keyEquivalent:@""];
        [pgweb setTarget:self];
        [pgweb setRepresentedObject:@"pgweb"];
        [self.quickAccessSubmenu addItem:pgweb];
    }
}
```

[ASSUMED: NSMenuItem `setRepresentedObject:` is the idiomatic way to attach arbitrary data to a menu item. The `tag` (int) approach used in Phase 2 for service indices could also be used with a parallel string array for names.]

### Anti-Patterns to Avoid

- **Storing NSMenuItem objects across calls for site items:** Site items are rebuilt from scratch on every `menuWillOpen:`. Do NOT cache per-site NSMenuItems in a mutable array between opens — the count and content may change.
- **Calling `browser.OpenURL` from the main thread synchronously in a callback:** Always `go func()` as done in Phase 2 for service actions (D-06). ObjC must not block waiting for browser launch.
- **Importing `internal/sites` or `internal/services` from `internal/tray`:** The `AppController` interface is the only coupling point. Pass plain value types (SiteInfo, WebAdminItem), not domain types.
- **Relying on `Manager.AllWebAdmin()` alone for pgweb:** pgweb is NOT registered as a WebAdminService in the registry (only phpmyadmin is). Tray must query pgweb via a separate path.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Opening URLs in browser | Custom `exec.Command("open", url)` | `github.com/pkg/browser` | Already imported; handles edge cases, error handling, cross-platform interface |
| Sorting by RFC3339 timestamp | Manual date parsing loop | `sort.Slice` with string comparison | RFC3339 is lexicographically sortable; no time.Parse needed |
| Menu item data attachment | Parallel index arrays | `setRepresentedObject:` or `tag` + parallel array | ObjC provides idiomatic data attachment; Phase 2 tag pattern already proven |

**Key insight:** The CGO/ObjC bridge layer already exists and is battle-tested from Phase 2. Phase 3 adds new menu items following the exact same structural pattern — no new architectural concepts are introduced.

---

## Common Pitfalls

### Pitfall 1: pgweb Not in WebAdmin Registry

**What goes wrong:** Code calls `instance.app.GetWebAdminItems()` and only phpMyAdmin appears — pgweb is silently missing.
**Why it happens:** `mgr.RegisterWebAdmin("phpmyadmin", ...)` is called in `NewApp()`, but pgweb has no corresponding `RegisterWebAdmin` call. pgweb is managed via `App.Pgweb *pgweb.Manager` directly.
**How to avoid:** The new `GetWebAdminItems()` method on `App` must explicitly include pgweb by querying `a.Pgweb.IsInstalled()` and `a.Pgweb.URL()` alongside the WebAdmin registry.
**Warning signs:** Test scenario with pgweb installed but not appearing in tray.

[VERIFIED: app.go lines 86-88 confirm only phpmyadmin is RegisterWebAdmin'd. pgweb.Manager is a separate struct field.]

### Pitfall 2: C Memory Leaks in onRefreshQuickAccess

**What goes wrong:** `C.CString()` allocations for site domain/URL/label strings are not freed, causing memory leaks on every menu open.
**Why it happens:** Each call to `C.CString()` allocates C memory that Go's GC does not track.
**How to avoid:** Always `defer C.free(unsafe.Pointer(cStr))` or free immediately after the C call, matching the pattern in `onRefreshStatuses`.
**Warning signs:** Memory growth visible in Activity Monitor when frequently opening the tray menu.

[VERIFIED: controller_darwin.go onRefreshStatuses — each C.CString is freed with C.free]

### Pitfall 3: Rebuilding vs. Updating Quick Access Items

**What goes wrong:** Trying to update existing site NSMenuItems in-place (like Phase 2 does for service status) instead of removing and recreating them.
**Why it happens:** Phase 2 uses a fixed-size array of service items (always 5) with in-place updates. Site count is variable.
**How to avoid:** Use `[submenu removeAllItems]` at the start of each rebuild. Rebuild from scratch on every `menuWillOpen:`. This is simpler and correct for variable-length lists.
**Warning signs:** Stale site items appearing after site deletion, or duplicate items accumulating.

[VERIFIED: sites.Manager.List() returns a fresh slice from config.Store on every call — no caching to worry about]

### Pitfall 4: Sort Direction for CreatedAt

**What goes wrong:** Sites appear oldest-first instead of newest-first.
**Why it happens:** Default `sort.Slice` ascending order sorts oldest (smallest timestamp string) first.
**How to avoid:** D-04 requires newest at top. Sort descending: `sites[i].CreatedAt > sites[j].CreatedAt`. RFC3339 strings are lexicographically comparable so string comparison works without `time.Parse`.
**Warning signs:** Newest site appears at the bottom of the Quick Access list.

[VERIFIED: config.Store.AddSite uses `time.Now().Format(time.RFC3339)` — RFC3339 is lexicographically sortable]

### Pitfall 5: Quick Access Submenu Inserted in Wrong Position

**What goes wrong:** "Quick Access" parent item appears after "Show Window" or at the bottom.
**Why it happens:** `CreateTray()` builds items in order. Quick Access must be inserted after the service items separator and before the "Show Window" separator (D-02).
**How to avoid:** Follow the menu order in `tray_darwin.m CreateTray()`: header → separator → service items → separator → [Quick Access parent] → separator → Show Window → separator → Quit.
**Warning signs:** Tray menu has wrong visual order after first launch.

[VERIFIED: tray_darwin.m CreateTray structure read directly]

---

## Code Examples

### Sorting sites by CreatedAt descending and capping at 15

```go
// Source: [ASSUMED pattern, standard Go sort.Slice]
import "sort"

func buildSiteInfoList(rawSites []sites.Site) (items []SiteInfo, overflow int) {
    // Sort newest first (D-04). RFC3339 is lexicographically sortable.
    sort.Slice(rawSites, func(i, j int) bool {
        return rawSites[i].CreatedAt > rawSites[j].CreatedAt
    })
    const maxSites = 15
    if len(rawSites) > maxSites {
        overflow = len(rawSites) - maxSites
        rawSites = rawSites[:maxSites]
    }
    items = make([]SiteInfo, len(rawSites))
    for i, s := range rawSites {
        scheme := "https"
        if !s.SSLEnabled {
            scheme = "http"
        }
        items[i] = SiteInfo{
            Domain: s.Domain,
            URL:    scheme + "://" + s.Domain,
            Label:  s.Domain, // Discretion: domain-only label
        }
    }
    return
}
```

### GetWebAdminItems on App (bridging phpmyadmin registry + pgweb direct)

```go
// Source: [ASSUMED — derived from app.go OpenWebAdmin + pgweb.Manager patterns]
// Place this on App in app.go or as a new tray-facing method.
func (a *App) GetWebAdminItems() []tray.WebAdminItem {
    var items []tray.WebAdminItem

    // phpMyAdmin via WebAdmin registry
    if pma, err := a.Manager.GetWebAdmin("phpmyadmin"); err == nil {
        items = append(items, tray.WebAdminItem{
            Name:      "phpmyadmin",
            Label:     "Open phpMyAdmin",
            Installed: pma.IsInstalled(),
            URL:       pma.URL(),
        })
    }

    // pgweb via direct Manager field (not in WebAdmin registry)
    items = append(items, tray.WebAdminItem{
        Name:      "pgweb",
        Label:     "Open pgweb",
        Installed: a.Pgweb.IsInstalled(),
        URL:       a.Pgweb.URL(),
    })

    return items
}
```

### C header additions (tray_darwin.h)

```c
// New functions for Quick Access data passing
void BeginQuickAccessRebuild(void);
void AddQuickAccessSite(int index, const char *domain, const char *url, const char *label);
void SetQuickAccessOverflow(int count);
void AddQuickAccessWebAdmin(const char *name, const char *label);
void CommitQuickAccessRebuild(void);
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Mutating existing menu items in-place | Rebuild submenu from scratch via removeAllItems | N/A — design choice for Phase 3 | Simpler correctness for variable-length lists |
| Passing Go struct slices through CGO | Passing individual C strings via repeated C calls | Phase 1 established this | Avoids CGO struct alignment issues |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `NSMenuItem setRepresentedObject:` is the idiomatic way to attach name string to web admin items | Architecture Patterns | Low — `tag` + parallel string array (Phase 2 pattern) is a proven alternative |
| A2 | RFC3339 string comparison (`>` operator) is sufficient for newest-first sort without time.Parse | Code Examples | Low — RFC3339 is a fixed-width ISO 8601 format, lexicographic order equals chronological order |
| A3 | `GetWebAdminItems()` should be a new method on `App` rather than using two separate `AppController` methods | Standard Stack / Architecture | Low — either approach works; single method is cleaner for the interface |
| A4 | A separator between sites and web admin tools should be included when both are present (discretion call) | Architecture Patterns | None — user explicitly gave Claude discretion on this |

---

## Open Questions (RESOLVED)

1. **Site label format (Claude's discretion)**
   - What we know: D-06 gives discretion; options are domain-only ("myapp.test") or domain + path hint ("myapp.test — ~/projects/myapp")
   - Recommendation: Domain-only. Path hints make items wide and inconsistent in length. Users who need the path can open the main window.

2. **Import cycle risk: tray importing app types**
   - What we know: `controller.go` defines `AppController` as an interface — no import of `internal/sites` or `internal/services/pgweb`. `SiteInfo` and `WebAdminItem` must be defined in `internal/tray/controller.go`, not imported from domain packages.
   - What's unclear: Whether `GetWebAdminItems()` can reference `tray.WebAdminItem` from `app.go` — this requires `app.go` to import `internal/tray`. Check if this is already the case.
   - Recommendation: Verify `app.go` already imports `internal/tray` (it does, via `a.Tray *tray.Controller`). Define `WebAdminItem` and `SiteInfo` in `internal/tray/controller.go`. The `App.GetWebAdminItems()` return type references `tray.WebAdminItem` — this is a circular import if `tray` imports `app`. Resolution: `AppController` is the interface that `App` implements. `App` imports `tray`, not the reverse. Return type on `App` method uses `tray.WebAdminItem`. This is valid.

---

## Environment Availability

Step 2.6: SKIPPED — this phase is code and CGO changes only. No new external tools, services, runtimes, or CLI utilities are required beyond the existing Go/CGO/Cocoa toolchain used in Phases 1 and 2.

---

## Sources

### Primary (HIGH confidence)
- `internal/tray/controller.go` — AppController interface and Controller struct [VERIFIED: read]
- `internal/tray/controller_darwin.go` — CGO callback pattern, //export functions [VERIFIED: read]
- `internal/tray/tray_darwin.m` — ObjC menu construction, menuWillOpen: pattern [VERIFIED: read]
- `internal/tray/tray_darwin.h` — C function declarations [VERIFIED: read]
- `internal/sites/manager.go` — Site struct, List() method [VERIFIED: read]
- `internal/services/service.go` — WebAdminService interface [VERIFIED: read]
- `internal/services/manager.go` — AllWebAdmin(), RegisterWebAdmin() [VERIFIED: read]
- `internal/services/phpmyadmin/service_adapter.go` — IsInstalled, URL() [VERIFIED: read]
- `internal/services/phpmyadmin/manager.go` — IsInstalled filesystem check [VERIFIED: read]
- `internal/services/pgweb/manager.go` — IsInstalled, URL() [VERIFIED: read]
- `app.go` — GetSites(), OpenWebAdmin(), OpenPgweb(), service registrations [VERIFIED: read]
- `internal/config/store.go` — SiteConfig.CreatedAt format (RFC3339) [VERIFIED: read]

### Secondary (MEDIUM confidence)
- None required — all research was done against the actual codebase.

### Tertiary (LOW confidence)
- None.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already present in codebase, no new dependencies
- Architecture: HIGH — follows identical patterns to Phase 2; all integration points verified by reading source
- Pitfalls: HIGH — all pitfalls derived from direct code inspection (pgweb registry gap, C memory pattern, sort direction)

**Research date:** 2026-04-18
**Valid until:** Until Phase 2 tray code changes (stable codebase)
