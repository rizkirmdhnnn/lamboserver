---
phase: 01-tray-foundation-window-lifecycle
plan: "01"
subsystem: tray
tags: [cgo, objc, macos, system-tray, nsstatusbar]
dependency_graph:
  requires: []
  provides: [internal/tray CGO package]
  affects: []
tech_stack:
  added: [CGO ObjC bridge, NSStatusBar, NSMenu, GCD dispatch]
  patterns: [consumer-defined interface (D-02), constructor injection, embed directive]
key_files:
  created:
    - internal/tray/doc.go
    - internal/tray/controller.go
    - internal/tray/icon.go
    - internal/tray/controller_darwin.go
    - internal/tray/tray_darwin.h
    - internal/tray/tray_darwin.m
    - internal/tray/tray_unsupported.go
    - internal/tray/icon_22x22.png
    - internal/tray/icon_22x22@2x.png
  modified: []
  deleted:
    - internal/tray/tray.go
decisions:
  - Custom CGO NSStatusBar package with LamboSystrayDelegate class (D-01, D-02)
  - All Cocoa mutations dispatched via GCD main queue (D-03)
  - setTemplate:YES on NSImage for automatic dark/light mode adaptation (D-04)
  - Menu structure: disabled header, separator, Show Window, separator, Quit (D-05, D-06)
  - Nil-guard all CGO callbacks (T-01-02 mitigation)
  - stdlib.h included in CGO preamble to make C.free available
metrics:
  duration_minutes: 1
  completed_date: "2026-04-17T16:09:55Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 9
  files_modified: 0
  files_deleted: 1
requirements:
  - TRAY-01
  - TRAY-02
  - TRAY-03
---

# Phase 01 Plan 01: Tray Foundation Package Summary

**One-liner:** Custom CGO NSStatusBar package with LamboSystrayDelegate, GCD dispatch, embedded template icons, and cross-platform no-op stubs.

## What Was Built

Created the complete `internal/tray/` package that provides the foundational system tray implementation for LamboServer. The package uses CGO to bridge Go to macOS AppKit APIs directly — no third-party systray library — because all existing Go tray libraries conflict with Wails v2's ObjC symbols at the linker level.

The package provides:
- `Controller` struct with `New()`, `Start()`, `Destroy()` public API
- `AppController` interface (consumer-defined per D-02) requiring only `Context() context.Context`
- ObjC `LamboSystrayDelegate` class with `NSStatusBar`/`NSStatusItem`/`NSMenu` implementation
- All Cocoa mutations dispatched via `dispatch_async(dispatch_get_main_queue(), ...)` (D-03)
- Template icon (22x22 1x, 44x44 2x Retina) with `setTemplate:YES` for dark/light mode (D-04)
- Menu: disabled "LamboServer v{version}" header, Show Window, Quit (D-05/D-06)
- Cross-platform no-op stubs for non-darwin builds
- Nil-guard on all CGO callbacks to prevent nil context panics (T-01-02)

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Create tray template icon assets (22x22 + 44x44 PNG) | f221465 | icon_22x22.png, icon_22x22@2x.png |
| 2 | Create internal/tray/ package with CGO bridge and ObjC implementation | 21c1b1d | doc.go, controller.go, icon.go, tray_darwin.h, tray_darwin.m, controller_darwin.go, tray_unsupported.go |

## Verification Results

- `go build ./internal/tray/` compiles without errors on darwin
- `LamboSystrayDelegate` ObjC class confirmed unique (no collision with Wails)
- `dispatch_async(dispatch_get_main_queue())` wraps all NSStatusItem and NSMenu mutations
- `setTemplate:YES` applied to icon for dark/light mode adaptation
- Icon files: icon_22x22.png (22x22 px), icon_22x22@2x.png (44x44 px) verified via `sips`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Added stdlib.h include to CGO preamble**
- **Found during:** Task 2
- **Issue:** `C.free` requires `stdlib.h` to be explicitly included when using CGO preambles; without it the compiler may warn or error on some toolchain versions
- **Fix:** Added `#include <stdlib.h>` to the CGO comment block in `controller_darwin.go`
- **Files modified:** internal/tray/controller_darwin.go
- **Commit:** 21c1b1d

None of the plan's main items were skipped or altered. The plan was followed exactly.

## Known Stubs

None. All icon embed directives reference real files at the correct dimensions.

## Threat Flags

No new network endpoints, auth paths, or trust boundaries introduced beyond those documented in the plan's threat model. The `onShowWindow` and `onQuit` CGO callbacks are nil-guarded per T-01-02 mitigation.

## Self-Check: PASSED

- internal/tray/doc.go: FOUND
- internal/tray/controller.go: FOUND
- internal/tray/icon.go: FOUND
- internal/tray/controller_darwin.go: FOUND
- internal/tray/tray_darwin.h: FOUND
- internal/tray/tray_darwin.m: FOUND
- internal/tray/tray_unsupported.go: FOUND
- internal/tray/icon_22x22.png: FOUND
- internal/tray/icon_22x22@2x.png: FOUND
- internal/tray/tray.go (deleted): CONFIRMED DELETED
- Commit f221465: FOUND
- Commit 21c1b1d: FOUND
- go build ./internal/tray/: PASSED
