---
phase: 01-tray-foundation-window-lifecycle
plan: "02"
subsystem: tray-lifecycle
status: partial
checkpoint_reached: task-2
tags:
  - tray
  - wails
  - lifecycle
  - window-management
dependency_graph:
  requires:
    - "01-01 (tray package: Controller, AppController, icon)"
  provides:
    - "Tray lifecycle wired into App startup/shutdown"
    - "HideWindowOnClose behavior"
    - "Context() method satisfying tray.AppController"
  affects:
    - main.go
    - app.go
tech_stack:
  added: []
  patterns:
    - "Nil-guard on a.Tray in shutdown() for T-02-01 race prevention"
    - "Tray init in startup() after a.ctx = ctx (per Pitfall 2)"
key_files:
  modified:
    - path: main.go
      change: "Added HideWindowOnClose: true; removed OnBeforeClose"
    - path: app.go
      change: "Added tray import, Tray field, Context() method, tray init in startup(), Destroy() in shutdown(), removed beforeClose()"
decisions:
  - "Tray initialized in startup() not NewApp() because it requires Wails context (available only after startup callback)"
  - "Nil-guard on a.Tray.Destroy() prevents panic if shutdown called before startup completes (T-02-01)"
  - "beforeClose() deleted entirely — HideWindowOnClose: true makes it dead code per Wails WindowDelegate.m behavior"
metrics:
  completed_date: "2026-04-17"
  tasks_completed: 1
  tasks_total: 2
  files_modified: 2
---

# Phase 01 Plan 02: Tray + Window Lifecycle Integration Summary

**One-liner:** Wires `tray.Controller` into `App` lifecycle — `HideWindowOnClose: true`, tray init in `startup()` after context, `Destroy()` in `shutdown()` before services, `beforeClose()` removed.

**Status: PARTIAL — awaiting human verification at Task 2 checkpoint.**

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Wire tray into main.go and app.go lifecycle | bd4a3d9 | main.go, app.go |

## Task 2 (Checkpoint — Not Executed)

**Type:** checkpoint:human-verify
**Name:** Verify tray icon, menu, and window lifecycle
**Status:** Awaiting human verification

## What Was Built (Task 1)

### main.go changes
- Added `HideWindowOnClose: true` to `options.App{}` after `MinHeight: 600`
- Removed `OnBeforeClose: app.beforeClose` line (D-08, D-11: no confirmation dialog)

### app.go changes
- Added import: `"github.com/rizkirmdhnnn/lamboserver/internal/tray"`
- Added `Tray *tray.Controller` field to `App` struct (after `Shell`)
- Added `Context() context.Context` method on `*App` to satisfy `tray.AppController` interface
- In `startup()`: added `tray.New(a, tray.Icon, "1.0.0")` + `Start()` after `a.ctx = ctx` is set
- Deleted entire `beforeClose()` method (dead code with `HideWindowOnClose: true`)
- In `shutdown()`: added nil-guarded `a.Tray.Destroy()` as first operation before `a.MySQL.Stop()`

## Verification Checklist (Task 1 Automated)

- [x] `main.go` contains `HideWindowOnClose: true`
- [x] `main.go` does NOT contain `OnBeforeClose`
- [x] `app.go` contains `import "github.com/rizkirmdhnnn/lamboserver/internal/tray"`
- [x] `app.go` contains `Tray *tray.Controller`
- [x] `app.go` contains `func (a *App) Context() context.Context`
- [x] `app.go` startup() contains `a.Tray = tray.New(a, tray.Icon, "1.0.0")` after `a.ctx = ctx`
- [x] `app.go` startup() contains `a.Tray.Start()`
- [x] `app.go` does NOT contain `func (a *App) beforeClose`
- [x] `app.go` shutdown() contains `a.Tray.Destroy()` before `a.MySQL.Stop()`
- [x] `go build ./internal/...` compiles without errors (frontend/dist embed skipped — pre-existing worktree limitation)

## Deviations from Plan

None — plan executed exactly as written.

## Threat Mitigations Applied

| Threat | Mitigation | Applied |
|--------|-----------|---------|
| T-02-01: shutdown race before startup | `if a.Tray != nil { a.Tray.Destroy() }` nil guard | Yes — shutdown() line 170 |

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or trust boundaries introduced.

## Self-Check: PASSED

- main.go exists and contains `HideWindowOnClose: true`: CONFIRMED
- app.go exists and contains `Tray *tray.Controller`: CONFIRMED
- app.go contains `Context()` method: CONFIRMED
- Commit bd4a3d9 exists: CONFIRMED
