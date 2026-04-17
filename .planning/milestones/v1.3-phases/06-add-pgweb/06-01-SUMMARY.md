---
phase: 06-add-pgweb
plan: "01"
subsystem: pgweb
tags: [pgweb, postgresql, web-admin, lifecycle, ipc]
requirements: [PGW-01, PGW-02, PGW-03]

dependency_graph:
  requires:
    - internal/system/paths.go (PgwebDir method)
    - internal/services/postgres (PostgreSQL must be running for pgweb to connect)
  provides:
    - internal/services/pgweb package (DaemonWebAdminService lifecycle)
    - app.go IPC methods (GetPgwebStatus, InstallPgweb, StartPgweb, StopPgweb, OpenPgweb)
  affects:
    - app.go shutdown and StopService (D-08 auto-stop coupling)
    - internal/system/paths.go EnsureDirectories (pgweb dir created on startup)

tech_stack:
  added:
    - pgweb 0.17.0 binary (downloaded from github.com/sosedoff/pgweb at runtime)
  patterns:
    - exec.Command + cmd.Start() for daemon process management (non-blocking)
    - syscall.Signal(0) probe for IsRunning() liveness check
    - proc.Kill() + proc.Wait() for zombie reaping on Stop()
    - net.Listen preflight probe for port conflict detection
    - OsFileSystem/SystemCommandRunner type alias exports for constructor ergonomics

key_files:
  created:
    - internal/services/pgweb/interfaces.go
    - internal/services/pgweb/manager.go
    - internal/services/pgweb/manager_test.go
  modified:
    - internal/system/paths.go (PgwebDir method + EnsureDirectories entry)
    - internal/config/defaults.go (DefaultPgwebPort constant)
    - app.go (Pgweb field, NewApp wiring, shutdown D-08, StopService D-08, 5 IPC methods)

decisions:
  - "Used type alias exports (OsFileSystem = osFileSystem) so NewApp can reference concrete types without exposing unexported names across package boundary"
  - "Added PgwebDir() to paths.go in Task 1 commit (prerequisite for compilation) rather than Task 2 to avoid split-dependency commit"
  - "D-08 coupling placed in both shutdown() and StopService() to cover both graceful app exit and manual PostgreSQL stop from the frontend"

metrics:
  duration: "~12 minutes"
  completed: "2026-04-16"
  tasks_completed: 2
  tasks_total: 2
  files_created: 3
  files_modified: 3
---

# Phase 6 Plan 01: pgweb Backend Package and App Wiring Summary

**One-liner:** pgweb daemon lifecycle (install/start/stop/status via exec.Command) with D-08 auto-stop coupling and 5 Wails IPC bindings.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Create pgweb package — interfaces, manager, and tests | 72faa55 | internal/services/pgweb/{interfaces,manager,manager_test}.go, internal/system/paths.go |
| 2 | Wire pgweb into paths, config, and app.go | 6ec674e | app.go, internal/config/defaults.go, internal/system/paths.go |

## What Was Built

### Task 1: pgweb Package

**`internal/services/pgweb/interfaces.go`** defines:
- `DaemonWebAdminService` interface (Install, IsInstalled, URL, Version, Start, Stop, IsRunning)
- `ServiceStatus` struct (Installed, Running, Port JSON fields)
- `FileSystem` and `CommandRunner` testability interfaces
- Constants: `pgwebVersion = "0.17.0"`, `downloadURLTemplate`, `defaultPort = 8081`

**`internal/services/pgweb/manager.go`** implements:
- `Manager` struct with `paths`, `fs`, `cmd`, `proc *exec.Cmd`, `mu sync.Mutex`
- `Install()` — arch-aware download (arm64/amd64), curl + unzip + chmod, best-effort zip cleanup, post-install binary verification
- `Start()` — port preflight via `checkPort()`, `exec.Command` with `--bind=127.0.0.1 --skip-open`, non-blocking `cmd.Start()` (NOT `cmd.Run`)
- `Stop()` — `proc.Process.Kill()` + `proc.Wait()` (zombie reap), idempotent when `proc == nil`
- `IsRunning()` — `syscall.Signal(0)` probe on live process
- `checkPort()` — `net.Listen("tcp", "127.0.0.1:8081")` preflight (T-06-02 mitigation)
- Compile-time check: `var _ DaemonWebAdminService = (*Manager)(nil)`

**`internal/services/pgweb/manager_test.go`** — 8 unit tests, all passing:
- TestIsInstalled_FalseWhenBinaryAbsent
- TestIsInstalled_TrueWhenBinaryPresent
- TestInstall_Success
- TestInstall_DownloadFails
- TestStopIdempotent
- TestURL
- TestVersion
- TestStatus_NotInstalled

### Task 2: App Wiring

**`internal/system/paths.go`:**
- Added `PgwebDir() string` method returning `~/.lamboserver/pgweb/`
- Added `p.PgwebDir()` to `EnsureDirectories()` slice

**`internal/config/defaults.go`:**
- Added `DefaultPgwebPort = 8081` constant

**`app.go`:**
- Import: `github.com/rizkirmdhnnn/lamboserver/internal/services/pgweb`
- App struct: `Pgweb *pgweb.Manager` field
- NewApp: `pgwebMgr := pgweb.NewManager(paths, pgweb.OsFileSystem{}, pgweb.SystemCommandRunner{})`
- shutdown(): `a.Pgweb.Stop()` before `a.PostgreSQL.Stop()` (D-08)
- StopService(): `if err == nil && name == "postgresql" { _ = a.Pgweb.Stop() }` (D-08)
- 5 new IPC methods: `GetPgwebStatus`, `InstallPgweb`, `StartPgweb`, `StopPgweb`, `OpenPgweb`

## Threat Model Compliance

| Threat | Mitigation | Status |
|--------|-----------|--------|
| T-06-01 Elevation of Privilege | `--bind=127.0.0.1` always set in Start() | Implemented |
| T-06-02 DoS port 8081 | `checkPort()` preflight in Start() | Implemented |
| T-06-03 Tampering binary download | HTTPS from GitHub (accepted) | Accepted |
| T-06-04 PG trust auth disclosure | localhost-only, standard for local dev (accepted) | Accepted |

## Verification Results

```
go test ./internal/services/pgweb/... -v   → 8/8 PASS
go test ./internal/...                     → all packages PASS (no regressions)
go vet ./internal/... ./pkg/...            → clean
go build ./internal/... ./pkg/...          → clean
grep "a.Pgweb.Stop()" app.go               → lines 174 (shutdown) + 279 (StopService)
```

Note: `go build ./...` fails with `pattern all:frontend/dist: no matching files found` — this is a pre-existing condition in the worktree (no built frontend assets). All Go packages compile correctly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added PgwebDir() in Task 1 commit instead of Task 2**
- **Found during:** Task 1 test run — `m.paths.PgwebDir undefined`
- **Issue:** pgweb/manager.go references `m.paths.PgwebDir()` which is defined in Task 2's scope, but Task 1 needs it to compile and run tests
- **Fix:** Added `PgwebDir()` method to paths.go in the Task 1 commit. Task 2 commit then added only the EnsureDirectories entry.
- **Files modified:** internal/system/paths.go
- **Commit:** 72faa55

**2. [Rule 2 - Missing critical functionality] Exported OsFileSystem and SystemCommandRunner as type aliases**
- **Found during:** Task 2 — NewApp needs to construct the manager with concrete production types
- **Issue:** Plan spec said "if unexported, create helper `pgweb.NewDefaultManager(paths)`" but the types needed to be accessible from app.go. Type aliases (`type OsFileSystem = osFileSystem`) are the idiomatic Go approach that satisfies both goals — zero-cost, no wrapper needed.
- **Fix:** Added `type OsFileSystem = osFileSystem` and `type SystemCommandRunner = systemCommandRunner` in manager.go
- **Files modified:** internal/services/pgweb/manager.go

## Known Stubs

None — all methods are fully implemented and wired.

## Threat Flags

None — no new trust boundaries beyond those documented in the plan's threat model.

## Self-Check: PASSED

- [x] internal/services/pgweb/interfaces.go — FOUND
- [x] internal/services/pgweb/manager.go — FOUND
- [x] internal/services/pgweb/manager_test.go — FOUND
- [x] internal/system/paths.go modified — FOUND
- [x] internal/config/defaults.go modified — FOUND
- [x] app.go modified — FOUND
- [x] Commit 72faa55 — FOUND
- [x] Commit 6ec674e — FOUND
