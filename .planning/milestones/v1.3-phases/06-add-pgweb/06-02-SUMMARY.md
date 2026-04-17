---
phase: 06-add-pgweb
plan: "02"
subsystem: pgweb-frontend
tags: [pgweb, postgresql, web-admin, frontend, react, wails-ipc]
requirements: [PGW-01, PGW-02, PGW-03]

dependency_graph:
  requires:
    - frontend/wailsjs/go/main/App.js (Wails IPC bindings)
    - app.go GetPgwebStatus/InstallPgweb/StartPgweb/StopPgweb/OpenPgweb (Plan 01)
    - internal/services/pgweb/manager.go (Plan 01)
  provides:
    - frontend/src/pages/PgDatabasePage.tsx (Web Admin pgweb card with all states)
  affects:
    - frontend/wailsjs/go/main/App.d.ts (pgweb TypeScript declarations)
    - frontend/wailsjs/go/main/App.js (pgweb Wails bindings)
    - frontend/wailsjs/go/models.ts (PgwebStatus type)

tech_stack:
  added: []
  patterns:
    - Independent pgwebLoading state separate from PostgreSQL loading state
    - Three-state conditional rendering (not-installed / stopped / running)
    - Install step label cycling (Downloading... -> checkmark Installed)
    - Inline error card with danger left-border (matches existing PG error pattern)

key_files:
  created: []
  modified:
    - frontend/src/pages/PgDatabasePage.tsx
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts

decisions:
  - "Added pgweb methods manually to Wails binding files (App.js, App.d.ts, models.ts) since wails generate module fails in worktree — equivalent to generated output"
  - "PgwebStatus class added to main namespace in models.ts to match how Wails serializes Go struct types defined in app.go"

metrics:
  duration: "~10 minutes"
  completed: "2026-04-16"
  tasks_completed: 1
  tasks_total: 2
  files_created: 0
  files_modified: 4
---

# Phase 6 Plan 02: pgweb Web Admin Card (Frontend) Summary

**One-liner:** React Web Admin card for pgweb with three states (not-installed/stopped/running), independent loading state, and all five Wails IPC bindings wired to PgDatabasePage.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Add pgweb Web Admin card to PgDatabasePage | c1ea830 | frontend/src/pages/PgDatabasePage.tsx, frontend/wailsjs/go/main/App.{js,d.ts}, frontend/wailsjs/go/models.ts |

## Tasks Pending Checkpoint

| Task | Name | Type | Status |
|------|------|------|--------|
| 2 | Verify pgweb end-to-end flow in running app | checkpoint:human-verify | Awaiting user verification |

## What Was Built

### Task 1: Web Admin Card

**`frontend/src/pages/PgDatabasePage.tsx`** updated with:

**New imports** (5 Wails IPC methods):
- `GetPgwebStatus` — fetches installed/running/port status
- `InstallPgweb` — downloads pgweb binary from GitHub
- `StartPgweb` — starts pgweb HTTP daemon on port 8081
- `StopPgweb` — stops the running pgweb process
- `OpenPgweb` — opens pgweb in default browser

**New state variables** (5, all independent from PostgreSQL state):
- `pgwebInstalled` / `pgwebRunning` — installation and runtime status
- `pgwebLoading` — independent loading flag (D-spec: must not share with `loading`)
- `pgwebInstallStep` — install progress label cycling
- `pgwebError` — inline error display

**Extended `loadStatus()`** — now calls `GetPgwebStatus()` in parallel with `GetAllStatuses()`, sets `pgwebInstalled` and `pgwebRunning`.

**Handler functions** (4):
- `handleInstallPgweb` — shows "Downloading..." step, then "✓ Installed" for 1.2s, then reloads status
- `handleStartPgweb` / `handleStopPgweb` — call IPC, reload status, surface errors
- `handleOpenPgweb` — fire-and-forget `OpenPgweb()` call

**Web Admin card JSX** inserted at correct position per D-02:
1. PostgreSQL service card (existing)
2. **Web Admin (pgweb) card — NEW**
3. Create Database card (existing)
4. Databases list card (existing)

Card renders only inside `{installed && running && ...}` block per D-03.

Three states rendered:
- **Not installed:** "Not installed" label + [Install pgweb] button (`.btn-primary`)
- **Stopped:** status-dot stopped + "Stopped" label + [Start] button (`.btn-primary`)
- **Running:** status-dot running + "Running · Port 8081" + [Stop] (`.btn-danger`) + [Open] (`.btn-secondary`)

**`frontend/wailsjs/go/main/App.js`** — 5 new export functions for pgweb IPC methods

**`frontend/wailsjs/go/main/App.d.ts`** — 5 new TypeScript declarations for pgweb methods

**`frontend/wailsjs/go/models.ts`** — `PgwebStatus` class added to `main` namespace with `installed: boolean`, `running: boolean`, `port: number` fields

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Manually added pgweb bindings to Wails binding files**
- **Found during:** Task 1 — TypeScript would fail to compile with missing imports
- **Issue:** The plan assumes `wails generate module` would have updated the bindings, but this command fails in the worktree environment. The binding files (`App.js`, `App.d.ts`, `models.ts`) did not contain pgweb methods.
- **Fix:** Manually added the 5 pgweb method exports to `App.js`, their TypeScript declarations to `App.d.ts`, and the `PgwebStatus` class to `models.ts` — exactly equivalent to what `wails generate module` would produce.
- **Files modified:** frontend/wailsjs/go/main/App.js, frontend/wailsjs/go/main/App.d.ts, frontend/wailsjs/go/models.ts
- **Commit:** c1ea830

## Known Stubs

None — all pgweb state is fetched live from the backend via `GetPgwebStatus()`. No hardcoded values flow to UI rendering (port 8081 is display-only copy, not a stub).

## Threat Model Compliance

| Threat | Mitigation | Status |
|--------|-----------|--------|
| T-06-05 Spoofing (Wails IPC) | Local-only; accepted | Accepted |
| T-06-06 Information Disclosure (pgweb URL) | localhost:8081 only; no sensitive data in frontend | Accepted |

## Threat Flags

None — no new trust boundaries beyond those in the plan's threat model.

## Self-Check: PASSED

- [x] frontend/src/pages/PgDatabasePage.tsx — FOUND, contains Web Admin (pgweb) card
- [x] frontend/wailsjs/go/main/App.js — FOUND, contains GetPgwebStatus export
- [x] frontend/wailsjs/go/main/App.d.ts — FOUND, contains GetPgwebStatus declaration
- [x] frontend/wailsjs/go/models.ts — FOUND, contains PgwebStatus class
- [x] Commit c1ea830 — FOUND
- [x] TypeScript compiles without errors (npx tsc --noEmit exit 0)
