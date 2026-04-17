---
phase: 6
slug: add-pgweb
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-16
updated: 2026-04-17
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (existing) |
| **Config file** | none — standard Go test infrastructure |
| **Quick run command** | `go test ./internal/services/pgweb/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/services/pgweb/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | PGW-01 | — | N/A | unit | `go test ./internal/services/pgweb/...` | ✅ | ✅ green |
| 06-01-02 | 01 | 1 | PGW-02 | T-06-02 | Port preflight | unit | `go test ./internal/services/pgweb/...` | ✅ | ✅ green |
| 06-01-03 | 01 | 1 | PGW-03 | — | N/A | manual | Browser opens pgweb at localhost:8081 | N/A | ⬜ manual |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/services/pgweb/manager_test.go` — 15 tests covering PGW-01, PGW-02

*Existing go test infrastructure covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Open pgweb in browser | PGW-03 | Requires real browser + running PostgreSQL | Start PostgreSQL, install pgweb, start pgweb, click Open — verify browser opens to localhost:8081 with databases listed |
| pgweb auto-stop on PG stop | D-08 | Requires running processes | Start PostgreSQL, start pgweb, stop PostgreSQL — verify pgweb process is also terminated |

---

## Test Inventory (15 tests)

| # | Test Name | Requirement | Behavior |
|---|-----------|-------------|----------|
| 1 | TestIsInstalled_FalseWhenBinaryAbsent | PGW-01 | Binary missing → false |
| 2 | TestIsInstalled_TrueWhenBinaryPresent | PGW-01 | Binary exists → true |
| 3 | TestInstall_Success | PGW-01 | Full install flow (download, unzip, chmod) |
| 4 | TestInstall_DownloadFails | PGW-01 | Download failure → error |
| 5 | TestStopIdempotent | PGW-02 | Stop on fresh manager → no panic |
| 6 | TestURL | PGW-02 | Returns http://127.0.0.1:8081 |
| 7 | TestVersion | PGW-01 | Returns 0.17.0 |
| 8 | TestStatus_NotInstalled | PGW-02 | Status struct when not installed |
| 9 | TestIsRunning_FalseWhenProcNil | PGW-02 | No process → false |
| 10 | TestIsRunning_TrueWhenProcessAlive | PGW-02 | Live subprocess → true |
| 11 | TestStart_IdempotentWhenAlreadyRunning | PGW-02 | Double-start → no error |
| 12 | TestStart_PortConflictReturnsError | PGW-02, T-06-02 | Port occupied → error |
| 13 | TestStart_MissingBinaryReturnsWrappedError | PGW-02 | Binary absent → wrapped error |
| 14 | TestCheckPort_FreePortAllowsStart | PGW-02 | Free port → no port error |
| 15 | TestCheckPort_OccupiedPortErrors | PGW-02, T-06-02 | Port occupied → specific error msg |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** complete

---

## Validation Audit 2026-04-17

| Metric | Count |
|--------|-------|
| Gaps found | 3 |
| Resolved | 3 |
| Escalated | 0 |
