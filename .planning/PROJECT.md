# LamboServer — System Tray Integration

## What This Is

LamboServer is a macOS local development environment manager that controls Nginx, MySQL, PHP, PostgreSQL, Node.js, and related tools through a native desktop GUI. This milestone adds system tray integration so the app runs in the background with a tray icon, giving users quick access to service controls, status, and navigation without needing the full window open.

## Core Value

Users can control their local dev services instantly from the system tray without opening the full application window.

## Requirements

### Validated

- ✓ Manage Nginx (install, start, stop, restart, configure) — existing
- ✓ Manage MySQL (install, start, stop, restart, databases, users) — existing
- ✓ Manage PHP (install, start, stop, switch versions, configure) — existing
- ✓ Manage PostgreSQL (install, start, stop, databases) — existing
- ✓ Manage Node.js (install, switch versions) — existing
- ✓ Manage dnsmasq for .test domain resolution — existing
- ✓ Create and manage local sites with SSL — existing
- ✓ phpMyAdmin and pgweb web admin tools — existing
- ✓ Debug logging and service log viewing — existing
- ✓ Dashboard with aggregated service status — existing

### Active

- [ ] System tray icon visible when app is running
- [ ] Tray menu shows service status indicators (running/stopped)
- [ ] Tray menu has Start/Stop/Restart submenus per service
- [ ] Tray menu has quick links to open sites in browser
- [ ] Tray menu has quick links to open web admin tools (phpMyAdmin, pgweb)
- [ ] "Show Window" menu item to toggle main window visibility
- [ ] Closing the window hides to tray instead of quitting
- [ ] "Quit" menu item that stops all services and exits the app
- [ ] Services keep running while app is minimized to tray

### Out of Scope

- Tray-only mode (no main window at all) — users still need the full UI for complex tasks like site creation, version management
- Notification badges on tray icon — adds complexity, not needed for v1
- Auto-start at login — can be added later as a separate feature
- Custom tray icon themes — single icon is sufficient

## Context

- LamboServer is built with Wails v2 (Go backend + React frontend)
- Wails v2 does not have built-in system tray support — will need a Go system tray library
- The app currently uses `beforeClose` hook to show a confirmation dialog — this needs to change to hide-to-tray behavior
- Services are managed via macOS launchd and persist independently of the app process
- The `App` struct in `app.go` is the central orchestrator — tray integration will need access to all service managers
- Current shutdown flow in `app.go:shutdown()` stops all services in dependency order — "Quit" from tray must trigger this same flow
- Ad-hoc code signing only (no Apple Developer account)

## Constraints

- **Platform**: macOS only — system tray implementation is Darwin-specific
- **Framework**: Wails v2 — tray must integrate with Wails lifecycle without conflicts
- **Signing**: Ad-hoc only — no notarization, no entitlements beyond current set
- **Architecture**: Must maintain existing service adapter pattern and App struct as central coordinator

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Hide to tray on close (no dialog) | Cleaner UX, less friction | — Pending |
| Services persist while minimized | Tray is a lightweight controller, not a service lifecycle boundary | — Pending |
| Full menu with submenus per service | Users want quick access without opening window | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-17 after initialization*
