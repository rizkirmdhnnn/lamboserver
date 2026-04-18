---
phase: 03-quick-access-links
plan: "02"
subsystem: tray
tags: [objc, tray, quick-access, submenu, cgo]
dependency_graph:
  requires: [03-01]
  provides: [ObjC Quick Access submenu construction, C bridge functions BeginQuickAccessRebuild/AddQuickAccessSite/SetQuickAccessOverflow/AddQuickAccessWebAdmin/CommitQuickAccessRebuild, menuWillOpen refresh integration]
  affects: [internal/tray/tray_darwin.m]
tech_stack:
  added: []
  patterns: [ObjC NSMenuItem representedObject for string data, removeAllItems rebuild pattern, NSMenuDelegate menuWillOpen refresh chaining]
key_files:
  created: []
  modified:
    - internal/tray/tray_darwin.m
decisions:
  - "rebuildQuickAccessSubmenu uses removeAllItems on every rebuild (not in-place updates) to handle variable site count correctly (Pitfall 3)"
  - "Separator between sites and web admin tools included only when both sections are present (D-10 discretion)"
  - "Quick Access submenu initialized in CreateTray with empty NSMutableArrays so C bridge functions are safe to call before first menuWillOpen"
  - "C bridge functions called synchronously from Go on main thread via menuWillOpen chain — no GCD dispatch needed inside them"
metrics:
  duration: "~10 min"
  completed: "2026-04-18T04:10:00Z"
  tasks_completed: 1
  files_changed: 1
---

# Phase 3 Plan 02: ObjC Quick Access Submenu Summary

**One-liner:** ObjC Quick Access submenu with 5 C bridge functions, site items, web admin items, overflow indicator, empty state, and menuWillOpen: refresh integration — completing the user-facing tray feature.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement ObjC Quick Access submenu with C bridge functions | 8d1f454 | internal/tray/tray_darwin.m |

## What Was Built

### Forward Declarations (tray_darwin.m lines 9-11)

Three new extern callbacks added alongside existing declarations:
- `extern void onOpenSite(const char *domain);`
- `extern void onOpenWebAdmin(const char *name);`
- `extern void onRefreshQuickAccess(void);`

### LamboSystrayDelegate Properties (lines 23-26)

Four new properties added to the delegate interface:
- `@property (strong, nonatomic) NSMenu *quickAccessSubmenu;` — the NSMenu attached to the "Quick Access" parent item
- `@property (strong, nonatomic) NSMutableArray<NSDictionary *> *quickAccessSites;` — site data from Go
- `@property (assign, nonatomic) int quickAccessOverflow;` — overflow count from Go
- `@property (strong, nonatomic) NSMutableArray<NSDictionary *> *quickAccessWebAdmins;` — web admin data from Go

### Action Methods

- `openSite:` — retrieves domain from `representedObject`, calls `onOpenSite([domain UTF8String])`
- `openWebAdmin:` — retrieves name from `representedObject`, calls `onOpenWebAdmin([name UTF8String])`

### rebuildQuickAccessSubmenu Method

Full rebuild on every call using `removeAllItems`:
- Empty state: disabled "No sites configured" item when no sites and no web admins (D-09)
- Site items: iterates `quickAccessSites`, sets `representedObject` to domain for `openSite:` action
- Overflow indicator: disabled "(+N more — see main window)" item when `quickAccessOverflow > 0` (D-05)
- Separator: added between sites and web admin sections only when both are present (D-10)
- Web admin items: iterates `quickAccessWebAdmins`, sets `representedObject` to name for `openWebAdmin:` action (D-07/D-08)

### menuWillOpen: Update (line 125)

Added `onRefreshQuickAccess()` call after existing `onRefreshStatuses()` — sites and web admin status refresh on every menu open (D-11).

### CreateTray Menu Insertion (lines 261-271)

Quick Access parent item inserted between the services separator and Show Window:
- Menu order: LamboServer header -> separator -> 5 service items -> separator -> Quick Access -> separator -> Show Window -> separator -> Quit
- `quickAccessSubmenu`, `quickAccessSites`, `quickAccessWebAdmins` initialized as empty collections at creation time

### 5 C Bridge Functions (lines 321-352)

Implement the data-passing protocol declared in `tray_darwin.h` (Plan 01):
- `BeginQuickAccessRebuild(siteCount)` — clears all three data collections
- `AddQuickAccessSite(index, domain, url, label)` — appends NSDictionary to `quickAccessSites`
- `SetQuickAccessOverflow(count)` — sets `quickAccessOverflow` integer
- `AddQuickAccessWebAdmin(name, label)` — appends NSDictionary to `quickAccessWebAdmins`
- `CommitQuickAccessRebuild()` — calls `[delegate rebuildQuickAccessSubmenu]`

All functions operate synchronously on the main thread — no GCD dispatch needed.

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. The submenu reads live data passed from Go via the C bridge functions. All display logic uses actual runtime data.

## Threat Surface Scan

No new network endpoints, auth paths, or file access patterns introduced. All surface covered by plan's threat model (T-03-05, T-03-06, T-03-07).

## Checkpoint: Task 2 Pending Human Verification

Task 2 is a `checkpoint:human-verify` — requires running `wails dev` and visually confirming the Quick Access submenu in the macOS tray.

## Self-Check: PASSED

- File found: internal/tray/tray_darwin.m (119 lines added)
- Commit found: 8d1f454
- Build verified: `go build ./internal/tray/...` passes with no errors
- Pattern count: 22 occurrences of key symbols across rebuildQuickAccessSubmenu/quickAccessSubmenu/openSite/openWebAdmin/C bridge functions
