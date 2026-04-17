---
phase: 06-add-pgweb
reviewed: 2026-04-16T00:00:00Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - internal/services/pgweb/interfaces.go
  - internal/services/pgweb/manager.go
  - internal/services/pgweb/manager_test.go
  - internal/system/paths.go
  - internal/config/defaults.go
  - app.go
  - frontend/src/pages/PgDatabasePage.tsx
findings:
  critical: 0
  warning: 4
  info: 3
  total: 7
status: issues_found
---

# Phase 06: Code Review Report

**Reviewed:** 2026-04-16  
**Depth:** standard  
**Files Reviewed:** 7  
**Status:** issues_found

## Summary

The pgweb integration is well-structured overall: the `Manager` uses proper dependency injection, the interface boundary is clean, and lifecycle handling (mutex, signal-0 probe, idempotent Stop) follows established patterns in the codebase. Tests are meaningful and use testify mocks correctly.

Four warnings were found — all in `manager.go` and `app.go`. The most significant is a zombie-process risk in `Manager.Start()` when the pgweb process crashes between calls: the old `*exec.Cmd` is overwritten without being reaped. There is also pervasive hardcoding of port `8081` as a string literal in places that should reference the `defaultPort` constant (or `config.DefaultPgwebPort`). Three info items cover type safety and a missing `await` on the frontend.

---

## Warnings

### WR-01: Zombie process leak when pgweb crashes between Start calls

**File:** `internal/services/pgweb/manager.go:124-149`

`Start()` calls `IsRunning()` at line 128 to guard against double-starts. `IsRunning()` returns `false` when the process has exited (signal 0 returns an error). In that case `Start()` proceeds to `exec.Command(...).Start()` and assigns the new `*exec.Cmd` to `m.proc` at line 148, silently overwriting the old one. Because `Wait()` was never called on the crashed process, the OS keeps a zombie entry until the parent (this app) exits.

**Fix:** Before launching a new process, reap any previously-stored process that is no longer running:

```go
func (m *Manager) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Reap a crashed process before checking / relaunching.
    if m.proc != nil && !m.isRunningLocked() {
        m.proc.Wait() //nolint:errcheck — reap zombie
        m.proc = nil
    }

    if m.proc != nil {
        return nil // already running, idempotent
    }
    // ... rest of Start unchanged
}
```

(Extract the signal-0 check into a private `isRunningLocked()` helper that does not re-acquire the mutex, to be called from both `Start` and `IsRunning`.)

---

### WR-02: Port 8081 hardcoded as string literals instead of constant

**File:** `internal/services/pgweb/manager.go:114,138`

`checkPort()` at line 114 and the `exec.Command` args at line 138 both embed `"8081"` as raw string literals. The package already defines `defaultPort = 8081` in `interfaces.go:15`, and `config.DefaultPgwebPort = 8081` exists in `internal/config/defaults.go:29`. If the port ever changes, these literals will silently diverge.

```go
// line 114 — replace
ln, err := net.Listen("tcp", "127.0.0.1:8081")
// with
ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", defaultPort))

// line 138 — replace
fmt.Sprintf("--listen=8081")
// with
fmt.Sprintf("--listen=%d", defaultPort)
```

---

### WR-03: `URL()` hardcodes address instead of using the port constant

**File:** `internal/services/pgweb/manager.go:178`

```go
func (m *Manager) URL() string {
    return "http://127.0.0.1:8081"
}
```

This duplicates the literal again and will silently diverge if the port is changed in `defaultPort`.

**Fix:**
```go
func (m *Manager) URL() string {
    return fmt.Sprintf("http://127.0.0.1:%d", defaultPort)
}
```

---

### WR-04: `OpenPgweb()` in app.go hardcodes URL instead of delegating to `Manager.URL()`

**File:** `app.go:526`

```go
func (a *App) OpenPgweb() error {
    a.Debug.Info("OpenPgweb called")
    return browser.OpenURL("http://127.0.0.1:8081")
}
```

This duplicates the address for the third time across the codebase and bypasses the `Manager.URL()` method that exists for exactly this purpose.

**Fix:**
```go
func (a *App) OpenPgweb() error {
    a.Debug.Info("OpenPgweb called")
    return browser.OpenURL(a.Pgweb.URL())
}
```

---

## Info

### IN-01: Duplicate port constant — `defaultPort` vs `config.DefaultPgwebPort`

**File:** `internal/services/pgweb/interfaces.go:15` and `internal/config/defaults.go:29`

Both files define the same value 8081 under different names. The package-level `defaultPort` is unexported and inaccessible to callers outside the `pgweb` package, which is why `config.DefaultPgwebPort` was added. Having two authoritative sources risks them drifting out of sync.

**Fix:** Remove `defaultPort` from `interfaces.go` and have `manager.go` import and use `config.DefaultPgwebPort` directly (or re-export it as an exported constant).

---

### IN-02: `any` type cast on database list items in PgDatabasePage

**File:** `frontend/src/pages/PgDatabasePage.tsx:76`

```ts
setDatabases((dbs || []).map((d: any) => ({ name: d.name, size: d.size || "" })));
```

The Wails-generated `DatabaseEntry` type is available in `wailsjs/go/main/App`. Using `any` here bypasses TypeScript's type checking.

**Fix:**
```ts
import type { main } from "../../wailsjs/go/models";
// then:
setDatabases((dbs || []).map((d: main.DatabaseEntry) => ({ name: d.name, size: d.size || "" })));
```

---

### IN-03: `OpenPgweb()` IPC call not awaited and errors are silently dropped on the frontend

**File:** `frontend/src/pages/PgDatabasePage.tsx:222-224`

```ts
const handleOpenPgweb = () => {
    OpenPgweb();
};
```

`OpenPgweb` returns a `Promise<void>` (it can fail if `browser.OpenURL` fails). The promise is not awaited and any rejection is unhandled, meaning errors opening the browser are invisible to the user.

**Fix:**
```ts
const handleOpenPgweb = async () => {
    try {
        await OpenPgweb();
    } catch (e: any) {
        setPgwebError(String(e));
    }
};
```

---

_Reviewed: 2026-04-16_  
_Reviewer: Claude (gsd-code-reviewer)_  
_Depth: standard_
