---
phase: 01-remove-pgadmin
plan: "01"
subsystem: backend
tags: [pgadmin, removal, cleanup, go]
dependency_graph:
  requires: []
  provides: [pgadmin-go-code-removed]
  affects: [app.go, internal/system/paths.go]
tech_stack:
  added: []
  patterns: []
key_files:
  created: []
  modified:
    - app.go
    - internal/system/paths.go
    - internal/system/paths_test.go
  deleted:
    - internal/services/pgadmin/manager.go
    - internal/services/pgadmin/service_adapter.go
    - internal/services/pgadmin/interfaces.go
    - internal/services/pgadmin/manager_test.go
decisions:
  - "Deleted entire pgadmin package (4 files) rather than leaving empty stubs"
  - "Removed OpenPgAdmin Wails-exposed method from App struct since pgAdmin is no longer registered"
metrics:
  duration: "~5 minutes"
  completed: "2026-04-16"
  tasks_completed: 2
  tasks_total: 2
  files_modified: 3
  files_deleted: 4
requirements_satisfied: [REM-01, REM-02, REM-03, REM-04, REM-05]
---

# Phase 01 Plan 01: Delete pgAdmin Go Code Summary

**One-liner:** Deleted all pgAdmin Go code (4 files, ~440 lines) and stripped every reference from app.go and paths.go so the project compiles cleanly with zero pgAdmin symbols.

## What Was Built

Removed the entire `internal/services/pgadmin/` package and all references from the backend:

- **app.go**: Removed pgadmin import, `PgAdmin *pgadmin.Manager` struct field, `pgAdminMgr` variable creation, `RegisterWebAdmin("pgadmin", ...)` registration, `PgAdmin: pgAdminMgr` in return struct, and the entire `OpenPgAdmin()` method.
- **internal/system/paths.go**: Removed `PgAdminAppPath()`, `PgAdminSupportDir()`, and `PgAdminServersFile()` methods.
- **internal/system/paths_test.go**: Removed the `PgAdminAppPath` assertion from `TestPaths_PostgreSQLPaths`.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Delete pgAdmin package and strip app.go references | 639df5b | app.go, internal/services/pgadmin/* (deleted) |
| 2 | Remove pgAdmin path helpers and tests | 297a497 | internal/system/paths.go, internal/system/paths_test.go |

## Verification Results

- `internal/services/pgadmin/` directory: GONE
- `grep -ri pgadmin app.go internal/system/paths.go internal/system/paths_test.go`: NO MATCHES
- `go build ./...`: PASSES (with frontend/dist stub — frontend/dist missing from worktree is a pre-existing worktree constraint, not caused by this plan)
- `go build ./internal/...`: PASSES
- `go test ./internal/system/... -count=1`: PASSES
- All 14 internal test packages: PASS

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Threat Flags

No new security surface introduced. This plan only deletes code.

## Self-Check: PASSED

- `internal/services/pgadmin/` directory: confirmed absent
- `app.go`: confirmed no pgadmin/PgAdmin/pgAdmin references
- `internal/system/paths.go`: confirmed no PgAdmin references
- `internal/system/paths_test.go`: confirmed no PgAdmin references
- Commits 639df5b and 297a497 exist in git log
