---
phase: 02-verify-cloudbeaver-and-clean-up
reviewed: 2026-04-16T00:00:00Z
depth: standard
files_reviewed: 1
files_reviewed_list:
  - docs/DEVELOPMENT.md
findings:
  critical: 0
  warning: 3
  info: 2
  total: 5
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-04-16
**Depth:** standard
**Files Reviewed:** 1
**Status:** issues_found

## Summary

`docs/DEVELOPMENT.md` is the primary developer guide for the LamboServer project. The file
is well-structured and covers prerequisites, architecture, the manager pattern, and a thorough
step-by-step walkthrough for adding new services.

However, the document contains several stale references resulting from the recent replacement
of `pgadmin` with `cloudbeaver` (commit `2e696ce`). The project structure tree still lists
the deleted `pgadmin/` directory and makes no mention of `cloudbeaver/`, which now exists at
`internal/services/cloudbeaver/`. Additionally, a file referenced in both the tree
(`app_compat.go`) and the walkthrough section no longer exists on disk.

There are also two documentation accuracy issues: a contradictory conventions statement about
constant casing, and an example `Stop()` return-type inconsistency that conflicts with the
stated error-handling convention.

No security vulnerabilities or runtime bugs were found — this file is documentation only.
All issues are documentation correctness problems that could mislead contributors.

## Warnings

### WR-01: Project structure tree references deleted `pgadmin/` directory

**File:** `docs/DEVELOPMENT.md:127`
**Issue:** The project structure tree lists `pgadmin/` as a service package under
`internal/services/`. This directory was deleted as part of the OpenPgAdmin removal
(commit `2e696ce`). The actual filesystem has `internal/services/cloudbeaver/` in its
place, but `cloudbeaver` is entirely absent from the tree. A contributor following this
guide will not know the CloudBeaver service exists.
**Fix:** Replace the `pgadmin/` block in the tree with the `cloudbeaver/` block, listing
its actual files (`cloudbeaver.go`, `config.go`, `installer.go`, `service_adapter.go`):

```
├── cloudbeaver/  # CloudBeaver web-based DB admin
│   ├── cloudbeaver.go       # Manager struct and lifecycle methods
│   ├── config.go            # Configuration helpers
│   ├── installer.go         # Download and install logic
│   └── service_adapter.go   # Adapts Manager to services.WebAdminService
```

---

### WR-02: `app_compat.go` referenced in structure tree and walkthrough does not exist

**File:** `docs/DEVELOPMENT.md:49` and `docs/DEVELOPMENT.md:496`
**Issue:** Line 49 lists `app_compat.go` in the project structure tree with the description
"Legacy Wails bindings (backward-compatible, to be removed after frontend migration)". Line
496 instructs contributors: "For backward-compatible legacy methods, add them in
`app_compat.go`". The file does not exist in the repository. Root Go files are: `app.go`,
`doc.go`, `main.go`.
**Fix:** Remove the `app_compat.go` entry from the structure tree (line 49) and update the
Step 8 note (line 496) to direct contributors to `app.go` directly:

```
For new methods on the App struct, add them directly to `app.go`.
```

---

### WR-03: Example `Stop()` signature violates stated error-handling convention

**File:** `docs/DEVELOPMENT.md:416`
**Issue:** The Step 4 `manager.go` example defines `Stop()` with no return value:
```go
func (m *Manager) Stop() {
    // ... implementation
}
```
The Conventions section (line 595) states "Return errors explicitly — no panic except in
`init()` or unreachable paths". The service adapter on line 483-484 wraps the call as
`a.mgr.Stop(); return nil`, silently discarding any failure. This inconsistency could lead
contributors to implement managers whose Stop failures are invisible to callers.
**Fix:** Update the example to return an error, consistent with the rest of the walkthrough
and the project convention:

```go
func (m *Manager) Stop() error {
    // ... implementation
    return nil
}
```

And update the adapter accordingly:
```go
func (a *serviceAdapter) Stop() error { return a.mgr.Stop() }
func (a *serviceAdapter) Restart() error {
    if err := a.mgr.Stop(); err != nil {
        return err
    }
    return a.mgr.Start()
}
```

---

## Info

### IN-01: Conventions section incorrectly describes constant casing

**File:** `docs/DEVELOPMENT.md:580`
**Issue:** The Conventions section states: "Constants: UPPER_CASE (`ServiceLabel`,
`TestTLD`)". The given examples `ServiceLabel` and `TestTLD` are PascalCase (exported Go
identifiers), not UPPER_CASE (which would be `SERVICE_LABEL`). Go convention for exported
constants is PascalCase; UPPER_CASE is a C/shell convention not used in idiomatic Go.
**Fix:** Correct the description to match the examples and Go convention:

```
Constants: PascalCase for exported (`ServiceLabel`, `DefaultPort`), camelCase for unexported (`serviceName`, `defaultPort`)
```

---

### IN-02: CloudBeaver service is not mentioned anywhere in the guide

**File:** `docs/DEVELOPMENT.md` (whole file)
**Issue:** The CloudBeaver service now lives at `internal/services/cloudbeaver/` and has a
notable structural difference from other services: it does not follow the launchd pattern
(it uses `process.PIDFileManager` directly and has no `interfaces.go` file). The guide
makes no mention of CloudBeaver at all, including no note that the service adapter pattern
for process-managed services differs from the launchd-managed pattern.
**Fix:** Add a brief note after the project structure tree (around line 155) that identifies
CloudBeaver as a process-managed service (not launchd-managed) and points contributors to
`internal/services/cloudbeaver/` as the reference when adding similar Java/external-process
services.

---

_Reviewed: 2026-04-16_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
