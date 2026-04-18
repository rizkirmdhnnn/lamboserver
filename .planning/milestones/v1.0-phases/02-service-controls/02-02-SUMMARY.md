---
phase: 02-service-controls
plan: 02
status: complete
started: 2026-04-18
completed: 2026-04-18
---

# Summary: ObjC Service Menu with Status Dots and Submenus

## What Was Built

Complete ObjC implementation of service controls in the tray menu (`tray_darwin.m`):

1. **5 service menu items** in dependency order (D-02): DNSMasq, Nginx, PHP, MySQL, PostgreSQL
2. **Colored status dots** via `NSAttributedString`:
   - Green filled dot (`●`) for running services using `systemGreenColor`
   - Grey hollow dot (`○`) for stopped services using `secondaryLabelColor`
   - Dash (`─`) with "Not Installed" label for not-installed services using `tertiaryLabelColor`
3. **Start/Stop/Restart submenus** (D-03) with context-aware enable/disable (D-05):
   - Running: Start disabled, Stop/Restart enabled
   - Stopped: Start enabled, Stop/Restart disabled
4. **NSMenuDelegate `menuWillOpen:`** (D-07): Refreshes all statuses synchronously before menu displays
5. **Not-installed handling** (D-08): Disabled, no submenu, dash indicator
6. **`UpdateServiceStatus` C function**: Called from Go during refresh, updates individual menu items
7. **`RefreshServiceStatuses` C function**: Convenience wrapper calling Go `onRefreshStatuses` export

## Key Decisions

- No GCD dispatch in `UpdateServiceStatus` — the entire call chain from `menuWillOpen:` through Go and back is synchronous on the main thread
- Menu item tags store service index for action routing via `[sender tag]`
- Submenu dynamically attached/detached based on installation status

## Verification

- Human verification: APPROVED
- User confirmed: tray menu shows 5 services with correct status dots, submenus work with context-aware enabled/disabled state, actions trigger service operations, status refreshes on menu reopen

## Self-Check: PASSED

- `go build ./internal/tray/` compiles successfully
- All acceptance criteria verified (NSMenuDelegate, properties, selectors, colors, dots, context-aware enable/disable)
- `go test ./...` passes with no regressions

## Key Files

key-files.created:
- internal/tray/tray_darwin.m (modified — expanded from 84 to ~190 lines with full service menu UI)
