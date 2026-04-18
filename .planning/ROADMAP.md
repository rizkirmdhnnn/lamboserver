# Roadmap: LamboServer System Tray

## Overview

This milestone adds macOS system tray integration to LamboServer. The work proceeds in four phases: first the high-risk CGO foundation and window lifecycle that everything else depends on, then service controls, then quick-access links, and finally the build pipeline changes needed to ship the feature. Each phase delivers a verifiable capability before the next begins.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Tray Foundation & Window Lifecycle** - Custom CGO tray package renders icon in menu bar; closing window hides to tray
- [x] **Phase 2: Service Controls** - Tray menu shows per-service status and exposes Start/Stop/Restart actions
- [ ] **Phase 3: Quick Access Links** - Tray menu lists configured sites and web admin tools for one-click browser launch
- [ ] **Phase 4: Build Pipeline** - Info.plist patching and CI pipeline updated to support CGO tray builds

## Phase Details

### Phase 1: Tray Foundation & Window Lifecycle
**Goal**: A functional tray icon exists in the macOS menu bar and the window hides/restores correctly from it
**Depends on**: Nothing (first phase)
**Requirements**: TRAY-01, TRAY-02, TRAY-03, WNDW-01, WNDW-02, WNDW-03, WNDW-04
**Success Criteria** (what must be TRUE):
  1. A monochrome template icon appears in the macOS menu bar when the app is running, adapting to dark/light mode
  2. Closing the main window hides it to tray without showing a confirmation dialog and without stopping services
  3. Selecting "Show Window" from the tray menu reopens the main application window
  4. Selecting "Quit" from the tray menu stops all running services and exits the app cleanly
  5. Services continue running while the window is hidden (verified via launchctl or service status check)
**Plans**: 2 plans
Plans:
- [x] 01-01-PLAN.md — Create custom CGO tray package with NSStatusBar icon, ObjC bridge, and template icon assets
- [x] 01-02-PLAN.md — Wire tray into App lifecycle (HideWindowOnClose, startup/shutdown integration) and verify
**UI hint**: yes

### Phase 2: Service Controls
**Goal**: Users can see the status of each service and trigger Start/Stop/Restart directly from the tray menu
**Depends on**: Phase 1
**Requirements**: SRVC-01, SRVC-02, SRVC-03
**Success Criteria** (what must be TRUE):
  1. Tray menu shows a running/stopped status indicator for each of Nginx, MySQL, PHP, PostgreSQL, and dnsmasq
  2. Each service entry has a submenu containing Start, Stop, and Restart actions
  3. After triggering a Start/Stop/Restart action, the tray menu status reflects the updated state when next opened
**Plans**: 2 plans
Plans:
- [x] 02-01-PLAN.md — Expand AppController interface with service methods and add CGO bridge callbacks for actions/status refresh
- [x] 02-02-PLAN.md — Implement ObjC service menu items with colored status dots, submenus, NSMenuDelegate refresh, and verify

### Phase 3: Quick Access Links
**Goal**: Users can open configured sites and web admin tools in the browser directly from the tray menu
**Depends on**: Phase 2
**Requirements**: QKAC-01, QKAC-02, QKAC-03
**Success Criteria** (what must be TRUE):
  1. Tray menu lists up to 15 configured local sites, each with an "Open in Browser" action that launches the default browser
  2. "Open phpMyAdmin" appears in the tray menu and launches phpMyAdmin in the browser (item hidden when phpMyAdmin is not installed)
  3. "Open pgweb" appears in the tray menu and launches pgweb in the browser (item hidden when pgweb is not installed)
**Plans**: 2 plans
Plans:
- [ ] 03-01-PLAN.md — Expand AppController interface, add CGO callbacks, C header declarations, and App methods for Quick Access data pipeline
- [ ] 03-02-PLAN.md — Implement ObjC Quick Access submenu with site items, web admin items, refresh logic, and verify

### Phase 4: Build Pipeline
**Goal**: The app builds and ships correctly with CGO tray support — Info.plist is patched and CI produces a valid artifact
**Depends on**: Phase 3
**Requirements**: TRAY-04, TRAY-05
**Success Criteria** (what must be TRUE):
  1. Post-build script patches Info.plist with the correct keys for system tray behavior (LSUIElement or equivalent)
  2. CI pipeline produces a signed DMG artifact with CGO compilation enabled and no build failures
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Tray Foundation & Window Lifecycle | 2/2 | Complete | 2026-04-17 |
| 2. Service Controls | 2/2 | Complete | 2026-04-18 |
| 3. Quick Access Links | 0/2 | Not started | - |
| 4. Build Pipeline | 0/TBD | Not started | - |
