# Phase 1: Tray Foundation & Window Lifecycle - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 01-tray-foundation-window-lifecycle
**Areas discussed:** Tray icon design, Menu layout, Quit behavior

---

## Tray Icon Design

| Option | Description | Selected |
|--------|-------------|----------|
| App logo silhouette | A simplified monochrome version of the LamboServer logo/icon | ✓ |
| Letter mark | A bold 'L' or 'LS' monogram in the menu bar | |
| Server/stack icon | A generic server or stacked-layers symbol | |
| You decide | Claude picks something appropriate | |

**User's choice:** App logo silhouette
**Notes:** Existing app icon at `build/appicon.png` and `build/darwin/LamboServer.iconset/` will be used as the source for deriving a 22x22px monochrome template PNG.

---

## Menu Layout

### Menu Order

| Option | Description | Selected |
|--------|-------------|----------|
| Show Window first | Show Window at top, separator, Quit at bottom | |
| Quit at bottom only | App title/version at top, separator, Show Window, separator, Quit | |
| You decide | Claude picks the layout | ✓ |

**User's choice:** You decide (Claude's discretion)

### App Title Header

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, show header | "LamboServer v1.0.0" as disabled text at top | ✓ |
| No header | Jump straight to menu items | |
| You decide | Claude picks | |

**User's choice:** Yes, show header
**Notes:** Like Docker Desktop and Raycast pattern with non-clickable app name and version.

---

## Quit Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Quit immediately | No dialog — just stop services and exit | ✓ |
| Confirm with dialog | Show confirmation: 'This will stop all services. Quit?' | |
| You decide | Claude picks based on UX best practices | |

**User's choice:** Quit immediately
**Notes:** User chose intentional quit without friction. The existing Indonesian-language confirmation dialog in `app.go:157-169` will be removed entirely.

---

## Claude's Discretion

- Menu item ordering and separator placement within the Phase 1 structure
- CGO file organization within `internal/tray/`
- Dock icon hiding behavior (optional)
