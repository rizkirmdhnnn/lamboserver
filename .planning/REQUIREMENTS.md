# Requirements: LamboServer System Tray

**Defined:** 2026-04-17
**Core Value:** Users can control their local dev services instantly from the system tray without opening the full application window.

## v1 Requirements

Requirements for system tray milestone. Each maps to roadmap phases.

### Tray Foundation

- [ ] **TRAY-01**: System tray icon is visible in macOS menu bar when app is running
- [ ] **TRAY-02**: Tray icon uses monochrome template image that adapts to macOS dark/light mode
- [ ] **TRAY-03**: Custom CGO package (`internal/tray/`) using NSStatusBar directly, no third-party systray libraries
- [ ] **TRAY-04**: Post-build script patches Info.plist for correct tray behavior
- [ ] **TRAY-05**: CI pipeline updated to support CGO builds with system tray

### Window Lifecycle

- [ ] **WNDW-01**: Closing the main window hides to tray instead of quitting (no confirmation dialog)
- [ ] **WNDW-02**: "Show Window" menu item in tray reopens the main application window
- [ ] **WNDW-03**: "Quit" menu item stops all running services and exits the application
- [ ] **WNDW-04**: Services keep running while app is minimized to tray

### Service Controls

- [ ] **SRVC-01**: Tray menu shows per-service status indicators (running/stopped) for Nginx, MySQL, PHP, PostgreSQL, dnsmasq
- [ ] **SRVC-02**: Each service has a submenu with Start, Stop, and Restart actions
- [ ] **SRVC-03**: Tray menu refreshes service status after each action completes

### Quick Access

- [ ] **QKAC-01**: Tray menu lists configured sites with "Open in Browser" action (capped at 15 entries)
- [ ] **QKAC-02**: Tray menu has "Open phpMyAdmin" item (shown only when phpMyAdmin is installed)
- [ ] **QKAC-03**: Tray menu has "Open pgweb" item (shown only when pgweb is installed)

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Tray Enhancements

- **TRAY-06**: Dynamic Dock icon hiding (hide Dock icon when minimized, show when window is open)
- **TRAY-07**: Tray icon state changes on unexpected service failure (degraded state indicator)
- **TRAY-08**: Auto-start at macOS login
- **TRAY-09**: Per-site PHP version switching from tray submenu

## Out of Scope

| Feature | Reason |
|---------|--------|
| Tray-only mode (no main window) | Users need full UI for complex tasks like site creation, version management |
| Notification badges on tray icon | Adds complexity, not needed for v1 |
| Custom tray icon themes | Single monochrome template icon is sufficient |
| Left-click-to-show-window shortcut | macOS NSStatusItem cannot have both click handler and menu simultaneously |
| Log quick-access from tray | Low frequency use case, available in main window |
| Timer-based status polling | Pull-on-action model is simpler and sufficient |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| TRAY-01 | Phase 1 | Pending |
| TRAY-02 | Phase 1 | Pending |
| TRAY-03 | Phase 1 | Pending |
| TRAY-04 | Phase 4 | Pending |
| TRAY-05 | Phase 4 | Pending |
| WNDW-01 | Phase 1 | Pending |
| WNDW-02 | Phase 1 | Pending |
| WNDW-03 | Phase 1 | Pending |
| WNDW-04 | Phase 1 | Pending |
| SRVC-01 | Phase 2 | Pending |
| SRVC-02 | Phase 2 | Pending |
| SRVC-03 | Phase 2 | Pending |
| QKAC-01 | Phase 3 | Pending |
| QKAC-02 | Phase 3 | Pending |
| QKAC-03 | Phase 3 | Pending |

**Coverage:**
- v1 requirements: 15 total
- Mapped to phases: 15
- Unmapped: 0

---
*Requirements defined: 2026-04-17*
*Last updated: 2026-04-17 after roadmap creation*
