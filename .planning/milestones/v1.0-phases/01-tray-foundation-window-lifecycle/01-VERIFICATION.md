---
phase: 01-tray-foundation-window-lifecycle
verified: 2026-04-17T17:00:00Z
status: human_needed
score: 4/5
overrides_applied: 0
human_verification:
  - test: "Launch app with wails dev and inspect the macOS menu bar"
    expected: "A monochrome LamboServer icon appears in the menu bar; it is dark on light menu bars and light on dark menu bars (template adaptation). Clicking it shows: 'LamboServer v1.0.0' (grayed out), separator, 'Show Window', separator, 'Quit'."
    why_human: "Template icon rendering and dark/light mode adaptation require a running macOS display context — cannot be verified with grep or build checks."
  - test: "Click the red close button (X) on the main window"
    expected: "Window disappears silently — NO confirmation dialog. The tray icon remains in the menu bar. Services continue running (verify with: launchctl list | grep lamboserver)."
    why_human: "HideWindowOnClose: true is set in code, but the actual hide behavior and absence of dialog require a live Wails runtime."
  - test: "With window hidden, click tray icon and select 'Show Window'"
    expected: "The main application window reappears and comes to the front."
    why_human: "wailsRuntime.Show(ctx) is wired correctly in code, but the actual window restoration requires a live macOS Wails session."
  - test: "Click tray icon and select 'Quit'"
    expected: "Application exits completely. Tray icon disappears. Services stop (verify with: launchctl list | grep lamboserver shows no entries or all stopped)."
    why_human: "Quit path calls wailsRuntime.Quit -> shutdown() -> service stops. The end-to-end clean exit and service teardown require a live running app."
---

# Phase 1: Tray Foundation & Window Lifecycle — Verification Report

**Phase Goal:** A functional tray icon exists in the macOS menu bar and the window hides/restores correctly from it
**Verified:** 2026-04-17T17:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (from ROADMAP.md Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | A monochrome template icon appears in the macOS menu bar when the app is running, adapting to dark/light mode | ? HUMAN NEEDED | `setTemplate:YES` is set in tray_darwin.m (line 36); icon_22x22.png is 22x22px, icon_22x22@2x.png is 44x44px; `go build ./internal/tray/` compiles clean. Visual rendering requires human. |
| 2 | Closing the main window hides it to tray without showing a confirmation dialog and without stopping services | ? HUMAN NEEDED | `HideWindowOnClose: true` present in main.go (line 24); `OnBeforeClose` absent; `beforeClose()` method deleted. Runtime behavior requires human. |
| 3 | Selecting "Show Window" from the tray menu reopens the main application window | ? HUMAN NEEDED | `onShowWindow()` calls `wailsRuntime.Show(ctx)` in controller_darwin.go (line 47); Show Window menu item wired to delegate selector. End-to-end window restore requires human. |
| 4 | Selecting "Quit" from the tray menu stops all running services and exits the app cleanly | ? HUMAN NEEDED | `onQuit()` calls `wailsRuntime.Quit(ctx)` (line 59); shutdown() calls `a.Tray.Destroy()` then all service stops; nil-guard present. End-to-end exit requires human. |
| 5 | Services continue running while the window is hidden (verified via launchctl or service status check) | ✓ VERIFIED (structural) | `HideWindowOnClose: true` means no shutdown() is called on window close. Services are launchd daemons managed independently of window state. No code path stops services on window hide. Human still needed to confirm launchctl output at runtime. |

**Score (automated):** Structural prerequisites for all 5 truths are verified. 4 truths require human runtime confirmation.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/tray/doc.go` | Package documentation | ✓ VERIFIED | Contains `package tray` and full doc comment |
| `internal/tray/controller.go` | Controller struct, AppController interface, New constructor | ✓ VERIFIED | Exports `Controller`, `AppController`, `New`; `AppController` has `Context() context.Context` |
| `internal/tray/icon.go` | Embedded template icon bytes | ✓ VERIFIED | `//go:embed icon_22x22.png` and `//go:embed icon_22x22@2x.png` present |
| `internal/tray/icon_22x22.png` | 22x22 monochrome template icon | ✓ VERIFIED | Confirmed 22x22 px via sips |
| `internal/tray/icon_22x22@2x.png` | 44x44 Retina template icon | ✓ VERIFIED | Confirmed 44x44 px via sips |
| `internal/tray/controller_darwin.go` | CGO bridge to ObjC tray implementation | ✓ VERIFIED | `//go:build darwin`, `#cgo LDFLAGS: -framework Cocoa`, `Start()`, `Destroy()`, `onShowWindow`, `onQuit` |
| `internal/tray/tray_darwin.h` | ObjC header for C function declarations | ✓ VERIFIED | Contains `void CreateTray` and `void DestroyTray` |
| `internal/tray/tray_darwin.m` | ObjC implementation with NSStatusBar, NSMenu, LamboSystrayDelegate | ✓ VERIFIED | `LamboSystrayDelegate`, `dispatch_async(dispatch_get_main_queue())`, `setTemplate:YES`, "Show Window", "Quit" |
| `internal/tray/tray_unsupported.go` | No-op stubs for non-darwin builds | ✓ VERIFIED | `//go:build !darwin`, no-op `Start()` and `Destroy()` |
| `internal/tray/tray.go` (deleted) | Old placeholder must be gone | ✓ VERIFIED | File does not exist |
| `main.go` | HideWindowOnClose: true, no OnBeforeClose | ✓ VERIFIED | Line 24: `HideWindowOnClose: true`; `OnBeforeClose` absent |
| `app.go` | Tray field, Context() method, tray lifecycle | ✓ VERIFIED | Line 55: `Tray *tray.Controller`; line 164: `Context()`; line 154: tray init; line 171: `Tray.Destroy()` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/tray/controller_darwin.go` | `internal/tray/tray_darwin.m` | `#include "tray_darwin.h"` | ✓ WIRED | Line 9 of controller_darwin.go |
| `internal/tray/controller.go` | `internal/tray/icon.go` | Controller receives iconData via `tray.Icon` call | ✓ WIRED | app.go line 154: `tray.New(a, tray.Icon, "1.0.0")` |
| `app.go` | `internal/tray/controller.go` | `App` satisfies `tray.AppController` interface | ✓ WIRED | `func (a *App) Context() context.Context` at line 164 |
| `app.go startup()` | `internal/tray/controller_darwin.go Start()` | `a.Tray.Start()` after `a.ctx = ctx` | ✓ WIRED | ctx assigned at line 115; tray init at line 154-155 (ordering confirmed) |
| `app.go shutdown()` | `internal/tray/controller_darwin.go Destroy()` | `a.Tray.Destroy()` before service stops | ✓ WIRED | Line 171 (Destroy) before line 173 (MySQL.Stop) |
| `main.go` | Wails WindowDelegate.m | `HideWindowOnClose: true` triggers hide on close | ✓ WIRED | Line 24 of main.go |

### Data-Flow Trace (Level 4)

Not applicable — this phase delivers Go infrastructure (tray controller, CGO bridge) and Wails configuration, not React components rendering dynamic data.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| `internal/tray/` package compiles | `go build ./internal/tray/` | Exit 0, no output | ✓ PASS |
| Full project compiles | `go build -o /dev/null .` | Exit 0, no output | ✓ PASS |
| icon_22x22.png dimensions | `sips --getProperty pixelWidth pixelHeight` | 22x22 | ✓ PASS |
| icon_22x22@2x.png dimensions | `sips --getProperty pixelWidth pixelHeight` | 44x44 | ✓ PASS |
| tray icon appears in menu bar | Requires running `wails dev` | — | ? SKIP (needs live app) |
| window close hides to tray | Requires running `wails dev` | — | ? SKIP (needs live app) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| TRAY-01 | 01-01-PLAN.md | Tray icon visible in macOS menu bar | ? HUMAN NEEDED | NSStatusItem created via CreateTray; icon embedded and passed; compiles; visual appearance requires human |
| TRAY-02 | 01-01-PLAN.md | Monochrome template icon adapts to dark/light mode | ? HUMAN NEEDED | `[icon setTemplate:YES]` confirmed in tray_darwin.m line 36; actual rendering requires human |
| TRAY-03 | 01-01-PLAN.md | Custom CGO package using NSStatusBar directly | ✓ VERIFIED | `internal/tray/` package uses NSStatusBar via CGO ObjC bridge with no third-party systray library |
| WNDW-01 | 01-02-PLAN.md | Closing window hides to tray (no dialog) | ? HUMAN NEEDED | `HideWindowOnClose: true` set; `beforeClose()` deleted; `OnBeforeClose` absent; behavior requires human |
| WNDW-02 | 01-02-PLAN.md | "Show Window" reopens main window | ? HUMAN NEEDED | `wailsRuntime.Show(ctx)` wired in `onShowWindow()`; requires live runtime |
| WNDW-03 | 01-02-PLAN.md | "Quit" stops services and exits | ? HUMAN NEEDED | `wailsRuntime.Quit(ctx)` -> `shutdown()` -> service stops wired; requires live runtime |
| WNDW-04 | 01-02-PLAN.md | Services keep running while window hidden | ✓ VERIFIED (structural) | No code path stops services on window hide; launchd daemons are independent |

**Orphaned requirements check:** TRAY-04 and TRAY-05 are mapped to Phase 4 in REQUIREMENTS.md — not orphaned for this phase.

### Anti-Patterns Found

None detected. Scan of all modified files (internal/tray/*, main.go, app.go) found:
- No TODO/FIXME/placeholder comments
- No stub return patterns (return null, empty handlers)
- No hardcoded empty data flowing to rendering
- Nil-guards on CGO callbacks are correct safety patterns, not stubs

### Human Verification Required

#### 1. Tray Icon Appearance and Dark/Light Mode Adaptation

**Test:** Run `wails dev` in the project directory. Wait for the main window to appear. Look at the macOS menu bar.

**Expected:** A LamboServer icon is visible in the menu bar. On a light menu bar the icon is dark; on a dark menu bar the icon is light (macOS template image inversion). Clicking the icon shows a menu with "LamboServer v1.0.0" (grayed out, non-clickable), a separator, "Show Window", a separator, and "Quit".

**Why human:** Template icon rendering and dark/light mode inversion require a running macOS display context with NSStatusBar. Cannot be verified with static analysis or build checks. Covers TRAY-01 and TRAY-02.

#### 2. Window Close Hides to Tray (No Dialog)

**Test:** With the app running, click the red close button (X) on the main LamboServer window.

**Expected:** Window disappears instantly with no confirmation dialog. The tray icon remains in the menu bar. Run `launchctl list | grep lamboserver` — services should still be listed.

**Why human:** `HideWindowOnClose: true` is set in code but the actual Wails runtime behavior (suppressing the close event and hiding the window rather than quitting) requires a live session. Covers WNDW-01 and WNDW-04.

#### 3. Show Window Restores the Main Window

**Test:** With the window hidden (after step 2), click the tray icon and select "Show Window".

**Expected:** The main application window reappears and comes to the front of all windows.

**Why human:** `wailsRuntime.Show(ctx)` is wired in `onShowWindow()` which is the ObjC selector target, but the CGO callback chain (ObjC -> Go -> Wails runtime -> WebView show) requires a live running app. Covers WNDW-02.

#### 4. Quit Stops Services and Exits Cleanly

**Test:** With the app running (window visible or hidden), click the tray icon and select "Quit".

**Expected:** The application exits completely. The tray icon disappears from the menu bar. Running `launchctl list | grep lamboserver` shows no running service entries (or all are stopped).

**Why human:** The Quit path invokes `wailsRuntime.Quit(ctx)` which triggers the Wails shutdown hook -> `app.shutdown()` -> service stops in dependency order. The complete end-to-end teardown and service stop verification requires a live running app. Covers WNDW-03.

### Gaps Summary

No structural gaps. All artifacts exist, are substantive, and are correctly wired. The full project compiles cleanly. The 4 human verification items above are the only remaining gate — they confirm runtime behavior that static analysis cannot reach.

---

_Verified: 2026-04-17T17:00:00Z_
_Verifier: Claude (gsd-verifier)_
