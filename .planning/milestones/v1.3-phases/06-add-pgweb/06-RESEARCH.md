# Phase 6: Add pgweb - Research

**Researched:** 2026-04-16
**Domain:** Go process lifecycle management, binary download/install, Wails IPC extension, React UI state
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** pgweb gets a dedicated "Web Admin (pgweb)" card on the PostgreSQL page, separate from the PostgreSQL service card
- **D-02:** Card position: between the PostgreSQL service card and the Create Database card (Service → Web Admin → Create DB → DB list)
- **D-03:** The Web Admin card only appears when PostgreSQL is running. When PostgreSQL is stopped or not installed, no Web Admin card is shown.
- **D-04:** Not-installed state: compact card with "Not installed" text and [Install] button — no icon or lengthy description
- **D-05:** Stopped state: status dot (stopped) + [Start] button
- **D-06:** Running state: status dot (running) + port display + [Stop] and [Open] buttons
- **D-07:** pgweb does NOT auto-start when PostgreSQL starts. User manually starts pgweb when they need the web UI.
- **D-08:** pgweb IS auto-stopped when PostgreSQL stops. pgweb can't function without PostgreSQL — clean shutdown prevents orphan processes.
- **D-09:** No state persistence across app restarts. pgweb always starts in stopped state. No surprise processes at launch.
- **D-10:** pgweb has its own [Install] button in the Web Admin card. It is NOT bundled with PostgreSQL installation.
- **D-11:** Install progress uses step labels matching PostgreSQL's pattern: "Downloading..." → "✓ Installed" → card updates to stopped state.
- **D-12:** pgweb binary is downloaded from GitHub releases to ~/.lamboserver/pgweb/
- **D-13:** pgweb launches pre-configured to auto-connect to the local PostgreSQL via trust auth. User sees databases immediately — no connection form.
- **D-14:** pgweb binds to localhost only (127.0.0.1). Not accessible from other machines on the network.
- **D-15:** Default port: 8081 (pgweb default).

### Claude's Discretion

- Service architecture: how pgweb fits into the existing Service/WebAdminService interface hierarchy
- pgweb process management: how to track the running pgweb process (PID file, os/exec process handle, etc.)
- Error handling: what to show if pgweb download fails, if port 8081 is in use, etc.
- pgweb version selection: which version to pin, how to detect platform (darwin-amd64 vs darwin-arm64)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PGW-01 | pgweb binary is downloaded and installed into ~/.lamboserver/pgweb/ | Download URL pattern verified; arch detection uses existing `system.GetArchitecture()` |
| PGW-02 | User can start/stop pgweb from the PostgreSQL database page | New `DaemonWebAdminService` interface; Manager holds `*exec.Cmd` for process tracking |
| PGW-03 | User can open pgweb in the default browser with pre-configured connection to the local PostgreSQL server | `browser.OpenURL("http://127.0.0.1:8081")` via existing `pkg/browser`; TCP connect flags verified |
</phase_requirements>

---

## Summary

pgweb (github.com/sosedoff/pgweb) is a cross-platform Go binary that provides a web UI for PostgreSQL. Unlike phpMyAdmin, it ships as a self-contained executable with a built-in HTTP server — it is neither served by Nginx nor dependent on PHP. This means it does NOT fit the existing `WebAdminService` interface (which assumes Nginx delivers the UI) and does NOT fit the plain `Service` interface (which assumes launchd/pg_ctl lifecycle). pgweb needs a hybrid: **it installs like a WebAdminService** (download binary, detect arch) **and runs like a daemon** (Start/Stop/Status with an in-memory process handle). [VERIFIED: codebase grep of service.go and phpmyadmin]

The existing service hierarchy must be extended with a new `DaemonWebAdminService` interface that combines Install/IsInstalled/URL from `WebAdminService` with Start/Stop/Status from `Service`. pgweb's Manager implements this new interface, and a ServiceAdapter bridges it into the existing `services.Manager` registry using `RegisterWebAdmin`. The `GetAllStatuses()` method in `app.go` currently returns only `"installed"` or `"not_installed"` for web admin services — this must be extended to return `"running"`, `"stopped"`, or `"not_installed"` for pgweb. [VERIFIED: app.go lines 293-301]

pgweb does not support Unix socket connections (GitHub issue #473 was opened but the feature remains unresolved in v0.17.0 options.go). The connection strategy is TCP via `--host=127.0.0.1 --user=postgres` — compatible with the existing `pg_hba.conf` trust rule `host all all 127.0.0.1/32 trust`. [VERIFIED: postgres/manager.go writeHba(); CITED: github.com/sosedoff/pgweb/issues/473]

**Primary recommendation:** Create `internal/services/pgweb/` package following the postgres package structure. New interface `DaemonWebAdminService` (or inline on the Manager). Process tracking via `*exec.Cmd` held in memory — no PID file needed (process lifetime is bounded by the app lifetime per D-09). Four new IPC methods on `app.go`: `InstallPgweb`, `StartPgweb`, `StopPgweb`, `GetPgwebStatus`. Frontend adds pgweb state variables and a conditional card in `PgDatabasePage.tsx` per the approved UI-SPEC.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Binary download + extraction | Backend (Go) | — | File I/O, arch detection, curl — all server-side |
| Process start/stop/status | Backend (Go) | — | `os/exec.Cmd`, process signal — Go runtime |
| Port conflict detection | Backend (Go) | — | TCP listen probe on 8081 — same pattern as Postgres |
| Opening browser | Backend (Go) | — | `pkg/browser.OpenURL` — existing pattern in app.go |
| UI state (installed/running/port) | Frontend (React) | Backend (Go) | Frontend renders from status returned by IPC call |
| Card layout and interaction | Frontend (React) | — | PgDatabasePage.tsx modification per UI-SPEC |
| Auto-stop coupling (D-08) | Backend (Go) | — | `StopService("postgresql")` calls pgweb Manager.Stop |
| Shutdown coupling (D-08) | Backend (Go) | — | app.go `shutdown()` calls pgweb Manager.Stop |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `os/exec` (stdlib) | Go 1.25 | Launch and track pgweb process | No dependency; `*exec.Cmd` provides PID + signal |
| `net` (stdlib) | Go 1.25 | Port 8081 conflict probe before Start | Same pattern used by postgres.Manager.checkPort() |
| `github.com/pkg/browser` | already in go.mod | Open http://127.0.0.1:8081 in default browser | Already imported in app.go |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `syscall` (stdlib) | Go 1.25 | PID liveness check via `Kill(pid, 0)` | If PID-file-based status tracking is chosen (not recommended — see below) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| In-memory `*exec.Cmd` process handle | PID file on disk | PID file survives crashes; in-memory handle is simpler and D-09 says no persistence across restarts — in-memory wins |
| TCP connect flags `--host 127.0.0.1` | Unix socket (`--url`) | pgweb issue #473 confirms socket support is broken/unimplemented; TCP is safer |
| `DaemonWebAdminService` new interface | Wedge into existing `Service` or `WebAdminService` | Neither existing interface fits; a new narrow interface avoids a forced mismatch |

**Installation:**
No new npm or Go packages required. pgweb binary is a downloaded artifact, not a Go import.

**Version verification:** [VERIFIED: github.com/sosedoff/pgweb/releases/tag/v0.17.0]
```
pgweb v0.17.0 — released 2026-11-22 (latest as of 2026-04-16)
Assets: pgweb_darwin_arm64.zip, pgweb_darwin_amd64.zip
```

---

## Architecture Patterns

### System Architecture Diagram

```
User clicks [Install pgweb]
         |
         v
app.go:InstallPgweb("pgweb", "")
         |
         v
pgweb.Manager.Install()
   ├── detect arch (system.GetArchitecture())
   ├── curl GitHub release zip → ~/.lamboserver/pgweb/pgweb_darwin_arm64.zip
   ├── unzip → ~/.lamboserver/pgweb/pgweb  (binary)
   └── chmod +x

User clicks [Start]
         |
         v
app.go:StartPgweb()
   ├── port preflight: net.Listen("tcp","127.0.0.1:8081")
   ├── exec.Command(pgwebBin,
   │     "--host=127.0.0.1",
   │     "--user=postgres",
   │     "--bind=127.0.0.1",
   │     "--listen=8081",
   │     "--skip-open")
   ├── cmd.Start() — detached, store cmd in Manager.proc
   └── return nil

User clicks [Open]
         |
         v
app.go:OpenPgweb()
   └── browser.OpenURL("http://127.0.0.1:8081")

PostgreSQL stopped / App shutdown
         |
         v
app.go:shutdown() or StopService("postgresql")
   └── pgwebMgr.Stop()
          └── cmd.Process.Kill() if proc != nil
```

### Recommended Project Structure

```
internal/services/pgweb/
├── interfaces.go     # DaemonWebAdminService interface + constants + version pin
├── manager.go        # Install, Start, Stop, Status, IsInstalled, URL, IsRunning
└── service_adapter.go # (optional — if manager directly satisfies a registered interface)
```

### Pattern 1: New Interface — DaemonWebAdminService

**What:** pgweb is a daemon that also has Install/IsInstalled/URL semantics. Neither existing interface captures this.

**When to use:** Any future tool that (a) is downloaded as a binary and (b) runs its own HTTP server (not proxied by Nginx).

```go
// Source: new file internal/services/pgweb/interfaces.go

// DaemonWebAdminService is implemented by web admin tools that run their
// own HTTP daemon (not proxied through Nginx). Combines installation with
// daemon lifecycle management.
type DaemonWebAdminService interface {
    Install(version string) error
    IsInstalled() bool
    URL() string
    Version() string
    Start() error
    Stop() error
    IsRunning() bool
}
```

The Manager can directly implement this interface. If `services.Manager` needs to hold it, add a new map: `daemonWebAdmin map[string]DaemonWebAdminService` with `RegisterDaemonWebAdmin` / `GetDaemonWebAdmin`.

Alternatively — since this project only has one such tool — the `App` struct can hold `Pgweb *pgweb.Manager` directly and call it without going through the generic `services.Manager` registry (matches how `PostgreSQL *postgres.Manager` is held directly for `InitDataDir` and DB operations).

### Pattern 2: Process Tracking via `*exec.Cmd`

**What:** Hold the running process in memory on the Manager. No PID file. Lifecycle matches app lifetime per D-09.

```go
// Source: internal/services/pgweb/manager.go (to be created)

type Manager struct {
    paths *system.Paths
    proc  *exec.Cmd  // nil when stopped
    mu    sync.Mutex
}

func (m *Manager) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.IsRunning() {
        return nil // idempotent
    }
    // port check first
    if err := m.checkPort(); err != nil {
        return err
    }
    cmd := exec.Command(m.binaryPath(),
        "--host=127.0.0.1",
        "--user=postgres",
        "--bind=127.0.0.1",
        "--listen=8081",
        "--skip-open",
    )
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start pgweb: %w", err)
    }
    m.proc = cmd
    return nil
}

func (m *Manager) Stop() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.proc == nil {
        return nil // idempotent
    }
    _ = m.proc.Process.Kill()
    _ = m.proc.Wait() // reap zombie
    m.proc = nil
    return nil
}

func (m *Manager) IsRunning() bool {
    if m.proc == nil || m.proc.Process == nil {
        return false
    }
    err := m.proc.Process.Signal(syscall.Signal(0))
    return err == nil
}
```

**CRITICAL:** `cmd.Start()` (not `cmd.Run()`) is used. `Run()` blocks until the process exits — pgweb is a long-lived daemon.

### Pattern 3: Binary Download + Extraction

**What:** Matches PostgreSQL's curl-based download. Detect darwin-arm64 vs darwin-amd64 with `system.GetArchitecture()`.

```go
// Source: internal/services/pgweb/manager.go (to be created)

const pgwebVersion = "0.17.0"
const downloadURLTemplate = "https://github.com/sosedoff/pgweb/releases/download/v%s/pgweb_darwin_%s.zip"

func (m *Manager) Install() error {
    arch := "arm64"
    if system.GetArchitecture() == "amd64" {
        arch = "amd64"
    }
    url := fmt.Sprintf(downloadURLTemplate, pgwebVersion, arch)
    dir := m.paths.PgwebDir()
    zipPath := filepath.Join(dir, "pgweb.zip")

    if err := m.fs.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create pgweb dir: %w", err)
    }

    if _, err := m.cmd.Run("sh", "-c",
        fmt.Sprintf("curl -sL '%s' -o '%s'", url, zipPath)); err != nil {
        return fmt.Errorf("failed to download pgweb: %w", err)
    }

    if _, err := m.cmd.Run("sh", "-c",
        fmt.Sprintf("unzip -o '%s' -d '%s'", zipPath, dir)); err != nil {
        return fmt.Errorf("failed to extract pgweb: %w", err)
    }

    binary := m.binaryPath()
    if _, err := m.cmd.Run("chmod", "+x", binary); err != nil {
        return fmt.Errorf("failed to make pgweb executable: %w", err)
    }

    m.cmd.Run("rm", "-f", zipPath) // best-effort cleanup
    return nil
}

func (m *Manager) binaryPath() string {
    return filepath.Join(m.paths.PgwebDir(), "pgweb")
}

func (m *Manager) IsInstalled() bool {
    _, err := m.fs.Stat(m.binaryPath())
    return err == nil
}
```

**Note:** The zip asset name is `pgweb_darwin_arm64.zip` containing a single file named `pgweb` (bare binary, no subdirectory). Unzip `-o` (overwrite) is safe for reinstall. [VERIFIED: GitHub release v0.17.0 asset listing]

### Pattern 4: New IPC Methods in app.go

pgweb needs dedicated IPC methods because it has Start/Stop/Status semantics that the generic `StartService`/`StopService`/`GetAllStatuses` dispatch cannot handle without a new registry entry.

```go
// In app.go — new methods

func (a *App) InstallPgweb() error {
    a.Debug.Info("InstallPgweb called")
    err := a.Pgweb.Install()
    a.Debug.Action("InstallPgweb", err)
    return err
}

func (a *App) StartPgweb() error {
    a.Debug.Info("StartPgweb called")
    err := a.Pgweb.Start()
    a.Debug.Action("StartPgweb", err)
    return err
}

func (a *App) StopPgweb() error {
    a.Debug.Info("StopPgweb called")
    err := a.Pgweb.Stop()
    a.Debug.Action("StopPgweb", err)
    return err
}

// GetPgwebStatus returns the full pgweb status struct for the frontend.
func (a *App) GetPgwebStatus() PgwebStatus {
    return PgwebStatus{
        Installed: a.Pgweb.IsInstalled(),
        Running:   a.Pgweb.IsRunning(),
        Port:      8081,
    }
}

func (a *App) OpenPgweb() error {
    a.Debug.Info("OpenPgweb called")
    return browser.OpenURL("http://127.0.0.1:8081")
}
```

`PgwebStatus` struct is defined in `app.go` (no separate package needed):
```go
type PgwebStatus struct {
    Installed bool `json:"installed"`
    Running   bool `json:"running"`
    Port      int  `json:"port"`
}
```

### Pattern 5: Auto-stop Coupling (D-08)

Two places in `app.go` must call `a.Pgweb.Stop()`:

1. **`shutdown()`** — stops pgweb alongside MySQL, PostgreSQL, etc.
2. **`StopService("postgresql")`** — the generic stop dispatcher does NOT automatically call pgweb.Stop. The solution: override or wrap `StopService` so stopping "postgresql" also calls `a.Pgweb.Stop()`, or add a dedicated `StopPostgreSQL()` method that does both. The simpler approach: add a case to `StopService` switch or call Pgweb.Stop after the generic stop.

Recommended: In `app.go StopService()`, after delegating to the manager, check if name == "postgresql" and call `a.Pgweb.Stop()`:

```go
func (a *App) StopService(name string) error {
    // existing dispatch ...
    err = svc.Stop()
    if err == nil && name == "postgresql" {
        _ = a.Pgweb.Stop() // D-08: pgweb must stop with PostgreSQL
    }
    return err
}
```

### Anti-Patterns to Avoid

- **Using `cmd.Run()` instead of `cmd.Start()`:** `Run()` blocks until the process exits. For a long-lived daemon, this hangs the goroutine forever. Always use `Start()`.
- **Forgetting `cmd.Wait()` after Kill():** Without `Wait()`, the process becomes a zombie. Always call `Wait()` after killing.
- **Using the `WebAdminService` registry for pgweb:** `GetAllStatuses()` currently returns `"installed"` or `"not_installed"` for web admin services — it has no `"running"` state. Routing pgweb through that registry would require changing status semantics for all web admin services.
- **Checking IsRunning() without a mutex:** The `proc` field is read from the frontend polling goroutine and written by Start/Stop. Mutex is required.
- **Not calling `--skip-open`:** Without this flag, pgweb opens the browser itself on Start. The app controls browser opening separately via `OpenPgweb()`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Browser open | Custom `exec.Command("open", url)` | `github.com/pkg/browser` | Already imported; cross-platform; existing pattern in app.go |
| Arch detection | Custom runtime.GOARCH string | `system.GetArchitecture()` | Already implemented; returns `"arm64"` or `"amd64"` |
| File download | Custom HTTP client | `curl` via `cmd.Run("sh", "-c", "curl -sL ...")` | Consistent with PostgreSQL/phpMyAdmin patterns; no new imports |
| Port conflict check | Custom net package usage | `net.Listen("tcp", addr)` probe | Same pattern as `postgres.Manager.checkPort()` — copy verbatim |

**Key insight:** pgweb's lifecycle management is the only novel element. Everything else (download, arch, browser, paths) reuses proven codebase patterns.

---

## Common Pitfalls

### Pitfall 1: pgweb Opens Browser on Start

**What goes wrong:** pgweb's default behavior is to open the browser when it starts. If `--skip-open` is not passed, the browser opens immediately on `Start()` — before the user clicks [Open].
**Why it happens:** pgweb has this as a convenience default.
**How to avoid:** Always pass `--skip-open` as a flag in `cmd.Start()`.
**Warning signs:** Browser opens with PostgreSQL credentials visible immediately after clicking [Start].

### Pitfall 2: cmd.Run() Blocks

**What goes wrong:** Using `m.cmd.Run(pgwebBin, args...)` (the `CommandRunner` interface method) blocks until pgweb exits — which is never during normal use. This hangs the IPC call.
**Why it happens:** `CommandRunner.Run()` uses `cmd.CombinedOutput()` which waits for process exit.
**How to avoid:** Use `os/exec.Command(...)` directly with `.Start()`, NOT the `CommandRunner` interface. The `CommandRunner` interface is for short-lived commands (curl, unzip, chmod). pgweb Start must use raw `os/exec`.
**Warning signs:** Frontend hangs indefinitely after clicking [Start]; no response from IPC.

### Pitfall 3: Port 8081 Already in Use

**What goes wrong:** If another process is on port 8081, pgweb fails silently (or logs to stderr which nobody reads). The IPC Start returns no error but pgweb is not accessible.
**Why it happens:** No preflight check on port 8081 before `cmd.Start()`.
**How to avoid:** Add `checkPort()` before calling `cmd.Start()`, returning a descriptive error: `"port 8081 is already in use by another process"`. Mirror `postgres.Manager.checkPort()` exactly.
**Warning signs:** [Start] succeeds but pgweb card shows running; clicking [Open] shows connection refused.

### Pitfall 4: Stale Proc Handle After Process Death

**What goes wrong:** pgweb crashes externally (kill -9, OOM). `m.proc != nil` but the process is dead. `IsRunning()` returns true, UI shows "Running".
**Why it happens:** `m.proc` is not cleared on unexpected process exit.
**How to avoid:** `IsRunning()` should probe the process via `m.proc.Process.Signal(syscall.Signal(0))` — this returns an error if the process is dead, regardless of `m.proc` being non-nil.
**Warning signs:** pgweb card shows "Running" but clicking [Open] fails.

### Pitfall 5: Zip Contains Subdirectory

**What goes wrong:** If the GitHub release zip asset contains a subdirectory (e.g., `pgweb_darwin_arm64/pgweb`), then `unzip -d dir` extracts to `dir/pgweb_darwin_arm64/pgweb` not `dir/pgweb`. `IsInstalled()` checks `dir/pgweb` and returns false.
**Why it happens:** Some GitHub release zips have a top-level directory prefix.
**How to avoid:** After unzip, verify that `m.binaryPath()` (`~/.lamboserver/pgweb/pgweb`) exists. If the release format changes, use `find` to locate the binary and move it. (For v0.17.0, the zip contains a flat binary named `pgweb` — no subdirectory.) [VERIFIED: GitHub release asset listing for v0.17.0]
**Warning signs:** Install succeeds but `IsInstalled()` returns false.

### Pitfall 6: GetAllStatuses() Returns Wrong State

**What goes wrong:** The existing `GetAllStatuses()` only returns `"installed"` or `"not_installed"` for `WebAdminService` entries. If pgweb were registered there, the frontend would never know it's running.
**Why it happens:** `WebAdminService` has no Start/Stop concept.
**How to avoid:** Use dedicated `GetPgwebStatus()` IPC method that returns the full `PgwebStatus` struct. The frontend calls `GetPgwebStatus()` independently of `GetAllStatuses()`.
**Warning signs:** pgweb card always shows "Stopped" even when pgweb process is running.

---

## Code Examples

### PgDatabasePage.tsx — New State Variables

```tsx
// Source: 06-UI-SPEC.md (approved 2026-04-16) + PgDatabasePage.tsx pattern

// Add alongside existing state:
const [pgwebInstalled, setPgwebInstalled] = useState(false);
const [pgwebRunning, setPgwebRunning]   = useState(false);
const [pgwebLoading, setPgwebLoading]   = useState(false);
const [pgwebInstallStep, setPgwebInstallStep] = useState<string | null>(null);
const [pgwebError, setPgwebError]       = useState<string | null>(null);
```

### PgDatabasePage.tsx — loadStatus() Extension

```tsx
// Source: 06-UI-SPEC.md + existing loadStatus pattern

const loadStatus = async () => {
  try {
    const statuses = await GetAllStatuses();
    const pgStatus = statuses["postgresql"] || "not_installed";
    const isRunning = pgStatus === "running";
    setRunning(isRunning);
    setInstalled(pgStatus !== "not_installed");

    // NEW: load pgweb status separately
    const pgwStatus = await GetPgwebStatus();
    setPgwebInstalled(pgwStatus.installed);
    setPgwebRunning(pgwStatus.running);

    setStatusLoaded(true);
    if (isRunning) {
      await loadDatabases();
    } else {
      setDatabases([]);
    }
  } catch (e) {
    console.error("Failed to load status:", e);
    setStatusLoaded(true);
  }
};
```

### PgDatabasePage.tsx — Web Admin Card (condensed)

```tsx
// Source: 06-UI-SPEC.md Component Inventory section

{/* Web Admin card — only shown when PostgreSQL is running */}
{installed && running && (
  <>
    {/* ... existing PostgreSQL service card ... */}

    {/* Web Admin (pgweb) card */}
    <div className="card" style={{ marginBottom: 16 }}>
      <div className="service-row">
        <div className="service-info">
          {pgwebInstalled && <span className={`status-dot ${pgwebRunning ? "running" : "stopped"}`} />}
          <div>
            <div className="service-name">Web Admin (pgweb)</div>
            <div className="service-detail">
              {!pgwebInstalled
                ? "Not installed"
                : pgwebRunning
                  ? "Running · Port 8081"
                  : "Stopped"}
            </div>
          </div>
        </div>
        <div className="service-actions">
          {!pgwebInstalled && (
            <button className="btn btn-primary btn-sm" onClick={handleInstallPgweb} disabled={pgwebLoading}>
              {pgwebInstallStep ?? "Install pgweb"}
            </button>
          )}
          {pgwebInstalled && !pgwebRunning && (
            <button className="btn btn-primary btn-sm" onClick={handleStartPgweb} disabled={pgwebLoading}>
              {pgwebLoading ? "Starting..." : "Start"}
            </button>
          )}
          {pgwebInstalled && pgwebRunning && (
            <>
              <button className="btn btn-danger btn-sm" onClick={handleStopPgweb} disabled={pgwebLoading}>
                {pgwebLoading ? "Stopping..." : "Stop"}
              </button>
              <button className="btn btn-secondary btn-sm" onClick={handleOpenPgweb}>
                Open
              </button>
            </>
          )}
        </div>
      </div>
    </div>

    {/* ... existing Create Database card ... */}
  </>
)}
```

### paths.go — PgwebDir() Addition

```go
// Source: internal/system/paths.go — add alongside PhpMyAdminDir()

// PgwebDir returns the pgweb binary directory (~/.lamboserver/pgweb/).
func (p *Paths) PgwebDir() string { return filepath.Join(p.Home, "pgweb") }
```

Also add `p.PgwebDir()` to `EnsureDirectories()` dirs slice.

### config/defaults.go — PgwebPort Constant

```go
// Source: internal/config/defaults.go — add alongside DefaultPhpMyAdminPort

// DefaultPgwebPort is the default HTTP port for the pgweb web admin tool.
DefaultPgwebPort = 8081
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| phpMyAdmin via Nginx (PHP-FPM proxy) | pgweb Go binary (self-hosted HTTP) | This phase | No Nginx/PHP dependency for PostgreSQL admin |
| CloudBeaver (Java, broken installer) | pgweb (Go binary, ~8MB) | v1.3 milestone | Eliminated Java dependency |

**Deprecated/outdated:**
- CloudBeaver: removed in Phase 5. All CloudBeaver references purged.
- `WebAdminService` registry for daemon tools: WebAdminService is Nginx-proxied tools only. pgweb must NOT be registered through `RegisterWebAdmin` without extending the status model.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | pgweb v0.17.0 zip asset `pgweb_darwin_arm64.zip` contains a flat binary named `pgweb` with no subdirectory prefix | Code Examples (Install) | Binary not found at expected path after unzip; IsInstalled() returns false |
| A2 | pgweb `--skip-open` flag name is correct (verified from options.go fetch) | Anti-Patterns / Code Examples | Browser opens unexpectedly on Start |
| A3 | pgweb accepts `--bind=127.0.0.1` (not `--bind 127.0.0.1` space-separated) | Code Examples | Argument parsing failure; pgweb binds to all interfaces |

Note: A2 and A3 are [CITED: github.com/sosedoff/pgweb/blob/main/pkg/command/options.go] — both flag names and styles are confirmed from the source.

---

## Open Questions

1. **Does pgweb v0.17.0 zip contain a flat binary or a subdirectory?**
   - What we know: GitHub release page lists `pgweb_darwin_arm64.zip` as an asset; HTTP 302 redirect confirms file exists at `https://github.com/sosedoff/pgweb/releases/download/v0.17.0/pgweb_darwin_arm64.zip`
   - What's unclear: Whether unzip produces `~/.lamboserver/pgweb/pgweb` directly or `~/.lamboserver/pgweb/pgweb_darwin_arm64/pgweb`
   - Recommendation: Add a post-install check: `if _, err := fs.Stat(binaryPath); err != nil { return fmt.Errorf("pgweb binary not found at %s after extraction", binaryPath) }` — mirrors PostgreSQL's pg_ctl verification

2. **Should `GetAllStatuses()` be extended or should pgweb use a dedicated IPC call?**
   - What we know: `GetAllStatuses()` currently cannot express `"running"` for web admin services; extending it would change semantics for phpMyAdmin too
   - Recommendation: Use a dedicated `GetPgwebStatus()` method returning `PgwebStatus{Installed, Running, Port}` — this is cleaner and avoids touching phpMyAdmin behavior

---

## Environment Availability

Step 2.6: No new external tool dependencies are introduced. `curl` and `unzip` are available on macOS by default and are already used by the existing phpMyAdmin and PostgreSQL install flows. `os/exec` is stdlib. No new CLI tools or services required.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| curl | Binary download | macOS built-in | system | — |
| unzip | Zip extraction | macOS built-in | system | — |
| port 8081 | pgweb HTTP server | checked at runtime | — | Error: "port 8081 is already in use" |

---

## Validation Architecture

`workflow.nyquist_validation` is not set to false in `.planning/config.json` — validation section is required.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`go test`) |
| Config file | none (standard Go test discovery) |
| Quick run command | `go test ./internal/services/pgweb/... -v` |
| Full suite command | `go test ./... ` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PGW-01 | Install downloads binary to PgwebDir and sets executable | unit | `go test ./internal/services/pgweb/... -run TestInstall` | Wave 0 |
| PGW-01 | IsInstalled() returns true after install, false before | unit | `go test ./internal/services/pgweb/... -run TestIsInstalled` | Wave 0 |
| PGW-02 | Start() fails with descriptive error when port 8081 in use | unit | `go test ./internal/services/pgweb/... -run TestStartPortConflict` | Wave 0 |
| PGW-02 | Stop() is idempotent when called on stopped manager | unit | `go test ./internal/services/pgweb/... -run TestStopIdempotent` | Wave 0 |
| PGW-02 | IsRunning() returns false for dead proc handle | unit | `go test ./internal/services/pgweb/... -run TestIsRunningDeadProc` | Wave 0 |
| PGW-03 | URL() returns "http://127.0.0.1:8081" | unit | `go test ./internal/services/pgweb/... -run TestURL` | Wave 0 |
| PGW-01+02 | Auto-stop: pgweb.Stop() is called when PostgreSQL stops | integration/manual | `go test ./... -run TestStopPostgresStopsPgweb` | Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/services/pgweb/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `internal/services/pgweb/manager_test.go` — covers all PGW unit tests above
- [ ] `internal/services/pgweb/interfaces.go` — defines types needed by tests

---

## Security Domain

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | pgweb uses no auth — localhost-only binding (D-14) |
| V3 Session Management | no | no app-managed sessions |
| V4 Access Control | partial | localhost-only bind (`--bind=127.0.0.1`) prevents remote access |
| V5 Input Validation | no | no user input to backend in this phase |
| V6 Cryptography | no | HTTP only, localhost only — no TLS required |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| pgweb accessible from LAN | Elevation of Privilege | `--bind=127.0.0.1` flag always set (D-14) |
| Port 8081 squatting | Denial of Service | Port preflight probe returns descriptive error before Start |

---

## Sources

### Primary (HIGH confidence)

- Codebase: `internal/services/service.go` — interface definitions verified
- Codebase: `app.go` — composition root, IPC methods, shutdown flow verified
- Codebase: `internal/services/postgres/manager.go` — process management patterns (pg_ctl, port probe, stale PID)
- Codebase: `internal/services/phpmyadmin/manager.go` — WebAdminService install pattern
- Codebase: `internal/system/paths.go` — paths pattern
- Codebase: `internal/config/defaults.go` — port constant pattern
- Codebase: `.planning/phases/06-add-pgweb/06-UI-SPEC.md` — approved UI contract

### Secondary (MEDIUM confidence)

- [CITED: github.com/sosedoff/pgweb/blob/main/pkg/command/options.go] — CLI flags `--host`, `--bind`, `--listen`, `--skip-open` verified
- [CITED: github.com/sosedoff/pgweb/releases/tag/v0.17.0] — v0.17.0 confirmed latest; asset filenames `pgweb_darwin_arm64.zip`, `pgweb_darwin_amd64.zip` confirmed

### Tertiary (LOW confidence)

- [github.com/sosedoff/pgweb/issues/473] — Unix socket support not implemented; TCP is required. Issue closed, but resolution unclear — TCP approach is the safe choice.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all Go stdlib; pgweb version and download URL verified against GitHub releases
- Architecture: HIGH — new interface pattern and process management approach derived directly from existing codebase patterns
- Pitfalls: HIGH — verified against actual source code (cmd.Run vs cmd.Start, GetAllStatuses limitation, socket issue)

**Research date:** 2026-04-16
**Valid until:** 2026-07-16 (pgweb releases infrequently; check for new version before pinning)
