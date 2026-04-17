---
phase: 05-remove-cloudbeaver
reviewed: 2026-04-16T00:00:00Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - app.go
  - internal/config/defaults.go
  - internal/services/service.go
  - frontend/src/pages/PgDatabasePage.tsx
  - docs/DEVELOPMENT.md
findings:
  critical: 0
  warning: 4
  info: 4
  total: 8
status: issues_found
---

# Phase 05: Code Review Report

**Reviewed:** 2026-04-16T00:00:00Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Five files were reviewed at standard depth. The changes correctly remove CloudBeaver from the
codebase and introduce the PostgreSQL database management page. No security vulnerabilities
were found. The main concerns are a silent data-loss bug in `SaveConfig`, loading-state leaks
in the React page that leave the UI stuck in a loading state on the happy path, and error
swallowing in two async handlers that would silently fail for the user. One additional
inconsistency exists between `GetAllStatuses` and the `ServiceStatus` constants defined in
`service.go`.

---

## Warnings

### WR-01: `SaveConfig` silently discards the caller-supplied config

**File:** `app.go:661-663`
**Issue:** `SaveConfig` accepts a `cfg config.AppConfig` parameter to convey what should be
persisted, but the implementation calls `a.Config.Save()` without first applying `cfg` to the
store. The incoming value is completely ignored. Any caller that passes a modified config
struct to this method will believe the save succeeded, but the store will persist its existing
in-memory state instead — silently dropping the caller's changes.
**Fix:**
```go
func (a *App) SaveConfig(cfg config.AppConfig) error {
    if err := a.Config.Set(cfg); err != nil { // or whatever the setter is named
        return err
    }
    return a.Config.Save()
}
```
If `config.Store` has no `Set` method, the parameter should be removed and the method
documented as "save current in-memory config", or the store should be updated to accept a new
value before saving.

---

### WR-02: `handleStart` and `handleStop` never reset `loading` on the happy path

**File:** `frontend/src/pages/PgDatabasePage.tsx:97-122`
**Issue:** Both `handleStart` and `handleStop` set `setLoading(true)` at the top, then call
`await loadStatus()` but never call `setLoading(false)` afterwards on the success path. The
`finally`-less try/catch means that if no exception is thrown, `loading` remains `true`
indefinitely. The Start/Stop buttons will stay permanently disabled until another action
reloads the component.
**Fix:**
```typescript
const handleStart = async () => {
  setLoading(true);
  setError(null);
  try {
    await StartService("postgresql");
    await loadStatus();
  } catch (e: any) {
    console.error(e);
    setError(String(e));
  } finally {
    setLoading(false);  // add this
  }
};
```
Apply the same `finally` block to `handleStop`.

---

### WR-03: Status-load and database-load errors are swallowed silently

**File:** `frontend/src/pages/PgDatabasePage.tsx:52-55, 63-65`
**Issue:** Errors in `loadStatus` and `loadDatabases` are only written to `console.error`.
The user receives no feedback when these background loads fail. `loadDatabases` in particular
is called after start/stop actions, so a failure there leaves the database list stale with no
indication that something went wrong.
**Fix:**
```typescript
const loadDatabases = async () => {
  try {
    const dbs = await ListServiceDatabases("postgresql");
    setDatabases((dbs || []).map((d: any) => ({ name: d.name, size: d.size || "" })));
  } catch (e) {
    console.error("Failed to list databases:", e);
    setDatabases([]);
    setError("Failed to load databases: " + String(e));  // add this
  }
};
```
In `loadStatus`, surface the error similarly via `setError`.

---

### WR-04: `GetAllStatuses` emits ad-hoc strings not aligned with `ServiceStatus` constants

**File:** `app.go:290-302`
**Issue:** The WebAdmin branch of `GetAllStatuses` returns the literal strings `"installed"`
and `"not_installed"`. The `"not_installed"` string matches `StatusNotInstalled` from
`service.go`, but `"installed"` has no corresponding constant. The frontend must handle this
as a special case that is undocumented in `service.go`, creating an implicit contract that
can diverge silently when new web admin services are added.
**Fix:** Add a constant to `service.go` and use it in `app.go`:
```go
// service.go
StatusInstalled ServiceStatus = "installed"

// app.go
result[name] = string(services.StatusInstalled)
```

---

## Info

### IN-01: `UninstallWebAdmin` hard-codes service names instead of delegating to the registry

**File:** `app.go:465-473`
**Issue:** `UninstallWebAdmin` uses a `switch` statement that explicitly names `"phpmyadmin"`.
The new `adminer` service (the goal of this milestone) will require adding another case here.
This is inconsistent with `InstallWebAdmin`, which correctly delegates to the registry via
`a.Manager.GetWebAdmin(name)`. If `WebAdminService` had an `Uninstall()` method on its
interface, both install and uninstall could use the same registry lookup.

---

### IN-02: `(d: any)` type assertion loses API safety at the database list boundary

**File:** `frontend/src/pages/PgDatabasePage.tsx:61`
**Issue:** The map callback casts each database entry to `any` before accessing `.name` and
`.size`. Wails generates TypeScript types for `DatabaseEntry` — importing and using
`main.DatabaseEntry` would catch mismatches at compile time instead of at runtime.
**Fix:**
```typescript
import { main } from "../../wailsjs/go/models";
// ...
setDatabases((dbs || []).map((d: main.DatabaseEntry) => ({ name: d.name, size: d.size || "" })));
```

---

### IN-03: Port 5432 is hardcoded in the running-state UI label

**File:** `frontend/src/pages/PgDatabasePage.tsx:248`
**Issue:** The service detail string contains `"Port 5432"` as a string literal. The canonical
port is defined as `config.DefaultPostgresPort = 5432` in `internal/config/defaults.go`. If
the port ever changes (or becomes user-configurable), the UI label will show the wrong value.
Since the frontend does not directly import Go constants, the fix is to pass the port via the
existing `DashboardStatus` or a dedicated config API call, or at minimum leave a comment
referencing the constant.

---

### IN-04: `docs/DEVELOPMENT.md` project-structure tree references no-longer-present
`cloudbeaver` paths

**File:** `docs/DEVELOPMENT.md:127-130`
**Issue:** Lines 127–130 show orphaned indentation under `phpmyadmin/` that references
`config.go`, `installer.go`, and `service_adapter.go` at an incorrect tree level, and the
preceding directory listing may still reference CloudBeaver artifacts depending on the state
of the tree. The documentation should be updated to reflect the post-removal structure and to
add `adminer/` once that service is added.

---

_Reviewed: 2026-04-16T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
