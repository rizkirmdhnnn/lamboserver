---
phase: 03-quick-access-links
plan: "01"
subsystem: tray
tags: [cgo, tray, quick-access, interface]
dependency_graph:
  requires: []
  provides: [tray.AppController Quick Access methods, tray.SiteInfo, tray.WebAdminItem, CGO onOpenSite/onOpenWebAdmin/onRefreshQuickAccess callbacks, C Quick Access header declarations, App.GetTraySites/OpenSiteInBrowser/GetWebAdminItems/GetTotalSiteCount]
  affects: [internal/tray/controller.go, internal/tray/controller_darwin.go, internal/tray/tray_darwin.h, app.go]
tech_stack:
  added: []
  patterns: [CGO export callback with guard+goroutine, C string alloc/free in loop (no defer), consumer-defined interface extension]
key_files:
  created: []
  modified:
    - internal/tray/controller.go
    - internal/tray/controller_darwin.go
    - internal/tray/tray_darwin.h
    - app.go
decisions:
  - "GetTraySites named to avoid collision with Wails-bound App.GetSites() []sites.Site"
  - "pgweb handled via direct a.Pgweb Manager field in both GetWebAdminItems and OpenWebAdmin fallback — not in WebAdmin registry"
  - "C.free called immediately after each C function call in loops (no defer) to bound memory in onRefreshQuickAccess"
  - "URL scheme (https/http) determined from SSLEnabled config flag — no user-controllable scheme injection possible (T-03-03)"
  - "Site list capped at 15 entries in GetTraySites (D-05); overflow count passed via SetQuickAccessOverflow (T-03-04)"
metrics:
  duration: "~15 min"
  completed: "2026-04-18T03:06:15Z"
  tasks_completed: 2
  files_changed: 4
---

# Phase 3 Plan 01: Quick Access Go Bridge Summary

**One-liner:** Expanded AppController interface with 5 Quick Access methods, SiteInfo/WebAdminItem value types, three CGO export callbacks, and five C header declarations — complete Go data pipeline feeding site and web admin listings into the tray menu.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Expand AppController interface and define value types | a740f10 | internal/tray/controller.go |
| 2 | Add CGO callbacks, C header declarations, and App methods | ed05d79 | internal/tray/controller_darwin.go, internal/tray/tray_darwin.h, app.go |

## What Was Built

### AppController Interface (controller.go)

Five new methods added to the `AppController` interface:
- `GetTraySites() []SiteInfo` — sorted newest-first, capped at 15
- `OpenSiteInBrowser(domain string)` — opens site URL in browser
- `GetWebAdminItems() []WebAdminItem` — phpMyAdmin + pgweb install status
- `OpenWebAdmin(name string)` — opens named web admin tool
- `GetTotalSiteCount() int` — total site count for overflow label

Two new value types:
- `SiteInfo{Domain, URL, Label string}` — tray-facing site descriptor
- `WebAdminItem{Name, Label string; Installed bool; URL string}` — web admin descriptor

### CGO Bridge (controller_darwin.go)

Three new `//export` callbacks following the established guard+goroutine pattern:
- `onOpenSite(cDomain *C.char)` — routes to `app.OpenSiteInBrowser` in goroutine
- `onOpenWebAdmin(cName *C.char)` — routes to `app.OpenWebAdmin` in goroutine
- `onRefreshQuickAccess()` — drives full Quick Access data-passing protocol:
  - Calls `BeginQuickAccessRebuild` with site count
  - Loops sites: `AddQuickAccessSite` per site (C strings freed immediately)
  - Calls `SetQuickAccessOverflow` with capped overflow count
  - Loops installed web admins: `AddQuickAccessWebAdmin`
  - Calls `CommitQuickAccessRebuild`

Every `C.CString()` allocation has a matching `C.free()` called immediately after the C function call — no `defer` in loops to avoid unbounded accumulation.

### C Header (tray_darwin.h)

Five new declarations for the Quick Access data-passing protocol:
```c
void BeginQuickAccessRebuild(int siteCount);
void AddQuickAccessSite(int index, const char *domain, const char *url, const char *label);
void SetQuickAccessOverflow(int count);
void AddQuickAccessWebAdmin(const char *name, const char *label);
void CommitQuickAccessRebuild(void);
```

### App Methods (app.go)

- `GetTraySites()` — lists sites from `a.Sites.List()`, sorts by `CreatedAt` descending (RFC3339 lexicographic), caps at 15, computes scheme from `SSLEnabled`
- `OpenSiteInBrowser(domain)` — looks up domain in site config for scheme, falls back to https for unknown domains
- `GetWebAdminItems()` — returns phpMyAdmin via `a.Manager.GetWebAdmin("phpmyadmin")` and pgweb via `a.Pgweb` direct field
- `GetTotalSiteCount()` — returns `len(a.Sites.List())`
- `OpenWebAdmin(name)` — extended with pgweb fallback before final error return

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All methods return live data from existing managers. No hardcoded values or placeholder returns.

## Threat Surface Scan

No new network endpoints, auth paths, or file access patterns introduced. All new surface was covered by the plan's threat model:

| Flag | File | Description |
|------|------|-------------|
| T-03-03 mitigated | app.go | URL scheme hardcoded from SSLEnabled config, not user input |
| T-03-04 mitigated | app.go + controller_darwin.go | Site list capped at 15; C string allocations bounded and freed immediately |

## Self-Check: PASSED

All files found: internal/tray/controller.go, internal/tray/controller_darwin.go, internal/tray/tray_darwin.h, app.go, 03-01-SUMMARY.md
All commits found: a740f10 (Task 1), ed05d79 (Task 2)
