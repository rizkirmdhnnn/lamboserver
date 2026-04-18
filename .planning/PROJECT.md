# LamboServer — System Tray Integration

## What This Is

LamboServer is a macOS local development environment manager that controls Nginx, MySQL, PHP, PostgreSQL, Node.js, and related tools through a native desktop GUI. It features system tray integration for quick access to service controls without opening the full window.

## Core Value

Users can control their local dev services instantly from the system tray without opening the full application window.

## Current Milestone: v1.1 Build Pipeline Overhaul

**Goal:** Fix the "app is damaged" launch error and overhaul the entire build-to-distribution pipeline for reliable, professional DMG releases.

**Target features:**
- Fix "app is damaged" error — app must launch without quarantine workarounds on macOS Sequoia
- Professional DMG installer — custom background, icon layout, drag-to-Applications visual
- Improved code signing — investigate options beyond ad-hoc (self-signed cert, Developer ID)
- Robust CI pipeline — better versioning, artifact caching, build verification, reliable release workflow

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
- ✓ System tray icon visible when app is running — v1.0
- ✓ Tray menu shows service status indicators (running/stopped) — v1.0
- ✓ Tray menu has Start/Stop/Restart submenus per service — v1.0
- ✓ Tray menu has quick links to open sites in browser — v1.0
- ✓ Tray menu has quick links to open web admin tools — v1.0
- ✓ "Show Window" menu item to toggle main window visibility — v1.0
- ✓ Closing the window hides to tray instead of quitting — v1.0
- ✓ "Quit" menu item that stops all services and exits the app — v1.0
- ✓ Services keep running while app is minimized to tray — v1.0

### Active

- [ ] Fix "app is damaged" launch error on macOS Sequoia
- [ ] Professional DMG installer with custom background and drag-to-Applications
- [ ] Improved code signing beyond ad-hoc
- [ ] Robust CI pipeline with better versioning and build verification

### Out of Scope

- Tray-only mode (no main window at all) — users still need the full UI for complex tasks
- Notification badges on tray icon — adds complexity, not needed for v1
- Auto-start at login — can be added later (TRAY-08 deferred to v2)
- Custom tray icon themes — single icon is sufficient
- Dynamic Dock icon hiding — deferred to v2 (TRAY-06)
- Tray icon state on service failure — deferred to v2 (TRAY-07)
- Per-site PHP version switching from tray — deferred to v2 (TRAY-09)

## Context

- Shipped v1.0 System Tray milestone with custom CGO package using NSStatusBar
- 4 phases, 7 plans executed across 2 days
- Custom ObjC bridge (`internal/tray/`) integrates with Wails v2 event loop
- CI pipeline updated with CGO_ENABLED=1 and binary verification
- Ad-hoc code signing only (no Apple Developer account)

## Constraints

- **Platform**: macOS only — system tray implementation is Darwin-specific
- **Framework**: Wails v2 — tray must integrate with Wails lifecycle without conflicts
- **Signing**: Ad-hoc only — no notarization, no entitlements beyond current set
- **Architecture**: Must maintain existing service adapter pattern and App struct as central coordinator

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Custom CGO package using NSStatusBar directly | All standard systray libraries cause ObjC linker conflicts with Wails v2 | ✓ Good — works reliably |
| Hide to tray on close (no dialog) | Cleaner UX, less friction | ✓ Good |
| Services persist while minimized | Tray is a lightweight controller, not a service lifecycle boundary | ✓ Good |
| Fire-and-forget service actions | Menu closes naturally, status refreshes on next open | ✓ Good — simple and reliable |
| No LSUIElement — Dock icon stays visible | Users expect Dock presence for a GUI app; dynamic hiding deferred to v2 | ✓ Good |
| No Info.plist patching needed | NSStatusBar works at runtime via CGO, no plist keys required | ✓ Good — simplified build |

## Evolution

This document evolves at phase transitions and milestone boundaries.

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
*Last updated: 2026-04-18 after v1.1 milestone start*
