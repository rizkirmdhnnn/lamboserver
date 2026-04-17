# Phase 6: Add pgweb - Pattern Map

**Mapped:** 2026-04-16
**Files analyzed:** 8 new/modified files
**Analogs found:** 8 / 8

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/services/pgweb/interfaces.go` | config/types | — | `internal/services/postgres/interfaces.go` | exact |
| `internal/services/pgweb/manager.go` | service | event-driven (daemon process) | `internal/services/postgres/manager.go` | role-match (process mgmt differs) |
| `internal/services/pgweb/manager_test.go` | test | — | `internal/services/postgres/manager_test.go` | exact |
| `internal/system/paths.go` | config/utility | — | self (modification) | self |
| `internal/config/defaults.go` | config | — | self (modification) | self |
| `app.go` | provider/IPC | request-response | self (modification) | self |
| `frontend/src/pages/PgDatabasePage.tsx` | component | request-response | self (modification) | self |

---

## Pattern Assignments

### `internal/services/pgweb/interfaces.go` (config, types)

**Analog:** `internal/services/postgres/interfaces.go`

**Full file pattern** (lines 1–75):
```go
// Package pgweb provides pgweb daemon lifecycle management for LamboServer.
// pgweb is a self-contained Go binary that runs its own HTTP server — it is
// NOT proxied through Nginx and does NOT use launchd.
package pgweb

import "io/fs"

// pgwebVersion is the pinned pgweb version downloaded from github.com/sosedoff/pgweb.
const pgwebVersion = "0.17.0"

// downloadURLTemplate is the template for the pgweb zip archive URL.
// Assets: pgweb_darwin_arm64.zip, pgweb_darwin_amd64.zip
const downloadURLTemplate = "https://github.com/sosedoff/pgweb/releases/download/v%s/pgweb_darwin_%s.zip"

// defaultPort is the pgweb HTTP server port (matches DefaultPgwebPort in config/defaults.go).
const defaultPort = 8081

// DaemonWebAdminService is implemented by web admin tools that run their own
// HTTP daemon (not proxied through Nginx). Combines installation semantics
// from WebAdminService with daemon lifecycle management from Service.
type DaemonWebAdminService interface {
    Install() error
    IsInstalled() bool
    URL() string
    Version() string
    Start() error
    Stop() error
    IsRunning() bool
}

// ServiceStatus reports whether the pgweb binary is present and the process is running.
// Matches the pattern of postgres.ServiceStatus.
type ServiceStatus struct {
    Installed bool `json:"installed"`
    Running   bool `json:"running"`
    Port      int  `json:"port"`
}

// FileSystem is the filesystem subset that pgweb.Manager needs.
// Mirrors phpmyadmin.FileSystem (no ReadDir/ReadFile needed).
type FileSystem interface {
    Stat(name string) (fs.FileInfo, error)
    MkdirAll(path string, perm fs.FileMode) error
}

// CommandRunner is the exec subset for short-lived commands (curl, unzip, chmod, rm).
// Mirrors postgres.CommandRunner exactly.
type CommandRunner interface {
    Run(name string, args ...string) (string, error)
}
```

**Key notes:**
- `DaemonWebAdminService` is a new interface — no existing analog. pgweb is held directly on `App` struct (like `PostgreSQL *postgres.Manager`), not registered in `services.Manager`.
- `ServiceStatus` struct: copy from `postgres/interfaces.go` lines 22–26, substituting `Port: 8081`.
- `FileSystem` interface: subset of `postgres.FileSystem` — only `Stat` and `MkdirAll` needed (no `ReadFile`, `WriteFile`, `ReadDir`, `RemoveAll`).
- `CommandRunner` interface: copy verbatim from `postgres/interfaces.go` lines 40–42.

---

### `internal/services/pgweb/manager.go` (service, event-driven daemon)

**Analog:** `internal/services/postgres/manager.go`

**Imports pattern** (from `postgres/manager.go` lines 1–19, adapted):
```go
package pgweb

import (
    "fmt"
    "net"
    "os/exec"
    "path/filepath"
    "sync"
    "syscall"

    "github.com/rizkirmdhnnn/lamboserver/internal/system"
)
```

**Manager struct + constructor pattern** (from `postgres/manager.go` lines 48–74):
```go
// osFileSystem is the production implementation of FileSystem using the os package.
type osFileSystem struct{}

func (osFileSystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

// systemCommandRunner is the production implementation of CommandRunner.
type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) (string, error) {
    return system.RunCommand(name, args...)
}

// Manager handles pgweb binary download, process start/stop, and status reporting.
type Manager struct {
    paths *system.Paths
    fs    FileSystem
    cmd   CommandRunner
    proc  *exec.Cmd   // nil when stopped; non-nil when process is running
    mu    sync.Mutex  // guards proc field
}

// NewManager creates a Manager with injected dependencies.
func NewManager(paths *system.Paths, fs FileSystem, cmd CommandRunner) *Manager {
    return &Manager{paths: paths, fs: fs, cmd: cmd}
}
```

**IsInstalled pattern** (from `postgres/manager.go` lines 88–88 and `phpmyadmin/manager.go` lines 264–267):
```go
// IsInstalled reports whether the pgweb binary is present at the expected path.
func (m *Manager) IsInstalled() bool {
    _, err := m.fs.Stat(m.binaryPath())
    return err == nil
}

func (m *Manager) binaryPath() string {
    return filepath.Join(m.paths.PgwebDir(), "pgweb")
}
```

**Install (binary download) pattern** (from `phpmyadmin/manager.go` lines 67–120, adapted for zip):
```go
// Install downloads pgweb from GitHub releases to ~/.lamboserver/pgweb/.
// Arch detection mirrors postgres.Manager.Install() (lines 94–96).
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

    // Step 1: Download zip (mirrors phpmyadmin/manager.go lines 75-79).
    if _, err := m.cmd.Run("sh", "-c",
        fmt.Sprintf("curl -sL '%s' -o '%s'", url, zipPath)); err != nil {
        return fmt.Errorf("failed to download pgweb: %w", err)
    }

    // Step 2: Extract zip (mirrors phpmyadmin/manager.go lines 82-86).
    if _, err := m.cmd.Run("sh", "-c",
        fmt.Sprintf("unzip -o '%s' -d '%s'", zipPath, dir)); err != nil {
        return fmt.Errorf("failed to extract pgweb: %w", err)
    }

    // Step 3: Make executable.
    if _, err := m.cmd.Run("chmod", "+x", m.binaryPath()); err != nil {
        return fmt.Errorf("failed to make pgweb executable: %w", err)
    }

    // Step 4: Clean up zip (best-effort, mirrors phpmyadmin/manager.go line 96).
    m.cmd.Run("rm", "-f", zipPath) //nolint:errcheck

    // Step 5: Post-install verification (mirrors postgres/manager.go lines 121-124).
    if _, err := m.fs.Stat(m.binaryPath()); err != nil {
        return fmt.Errorf("pgweb binary not found at %s after extraction", m.binaryPath())
    }
    return nil
}
```

**Start/Stop/IsRunning pattern** (novel — based on RESEARCH.md Pattern 2; port check from `postgres/manager.go` lines 279–286):
```go
// checkPort probes port 8081. Returns an error with a descriptive message if occupied.
// Copy verbatim from postgres/manager.go lines 279-286, substituting port.
func (m *Manager) checkPort() error {
    ln, err := net.Listen("tcp", "127.0.0.1:8081")
    if err != nil {
        return fmt.Errorf("port 8081 is already in use by another process")
    }
    ln.Close()
    return nil
}

// Start performs a port preflight then launches pgweb as a background daemon.
// CRITICAL: uses exec.Command + cmd.Start() (NOT CommandRunner.Run which blocks).
func (m *Manager) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.IsRunning() {
        return nil // idempotent
    }
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

// Stop kills the pgweb process. Idempotent when already stopped.
func (m *Manager) Stop() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.proc == nil {
        return nil // idempotent
    }
    _ = m.proc.Process.Kill()
    _ = m.proc.Wait() // reap zombie — REQUIRED after Kill()
    m.proc = nil
    return nil
}

// IsRunning probes the process with signal 0 — returns false if proc is dead
// even when m.proc != nil (handles external process death / Pitfall 4).
// Pattern: syscall.Kill(pid,0) from postgres/manager.go lines 271-276.
func (m *Manager) IsRunning() bool {
    if m.proc == nil || m.proc.Process == nil {
        return false
    }
    err := m.proc.Process.Signal(syscall.Signal(0))
    return err == nil
}

// URL returns the pgweb HTTP URL.
func (m *Manager) URL() string { return "http://127.0.0.1:8081" }

// Version returns the pinned pgweb version.
func (m *Manager) Version() string { return pgwebVersion }

// Status returns a snapshot of the current pgweb state.
// Pattern: postgres/manager.go lines 254-260.
func (m *Manager) Status() ServiceStatus {
    return ServiceStatus{
        Installed: m.IsInstalled(),
        Running:   m.IsRunning(),
        Port:      defaultPort,
    }
}
```

**Error handling pattern** (from `postgres/manager.go` lines 122–125, `phpmyadmin/manager.go` lines 77–79):
```go
// Always wrap errors with fmt.Errorf("descriptive action: %w", err)
// Examples from analogs:
return fmt.Errorf("failed to download pgweb: %w", err)
return fmt.Errorf("failed to extract pgweb: %w", err)
return fmt.Errorf("failed to start pgweb: %w", err)
```

---

### `internal/services/pgweb/manager_test.go` (test)

**Analog:** `internal/services/postgres/manager_test.go`

**Package + imports pattern** (from `postgres/manager_test.go` lines 1–19):
```go
// Package pgweb_test contains unit tests for the pgweb package.
package pgweb_test

import (
    "io/fs"
    "net"
    "os"
    "testing"

    "github.com/rizkirmdhnnn/lamboserver/internal/services/pgweb"
    "github.com/rizkirmdhnnn/lamboserver/internal/system"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)
```

**Mock pattern** (from `postgres/manager_test.go` lines 23–76 — testify/mock style):
```go
// mockFS is a testify mock implementing pgweb.FileSystem.
type mockFS struct{ mock.Mock }

func (m *mockFS) Stat(name string) (fs.FileInfo, error) {
    args := m.Called(name)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(fs.FileInfo), args.Error(1)
}
func (m *mockFS) MkdirAll(path string, perm fs.FileMode) error {
    return m.Called(path, perm).Error(0)
}

// mockCmd is a testify mock implementing pgweb.CommandRunner.
type mockCmd struct{ mock.Mock }

func (m *mockCmd) Run(name string, args ...string) (string, error) {
    allArgs := append([]string{name}, args...)
    callArgs := m.Called(allArgs)
    return callArgs.String(0), callArgs.Error(1)
}
```

**Test helper pattern** (from `postgres/manager_test.go` lines 100–113):
```go
func newTestPaths(t *testing.T) *system.Paths {
    t.Helper()
    return &system.Paths{Home: t.TempDir()}
}

func newTestManager(t *testing.T) (*pgweb.Manager, *mockFS, *mockCmd, *system.Paths) {
    t.Helper()
    paths := newTestPaths(t)
    mfs := &mockFS{}
    mc := &mockCmd{}
    mgr := pgweb.NewManager(paths, mfs, mc)
    return mgr, mfs, mc, paths
}
```

**Port-free helper pattern** (from `postgres/manager_test.go` lines 116–123, adapted for port 8081):
```go
func port8081IsFree() bool {
    ln, err := net.Listen("tcp", "127.0.0.1:8081")
    if err != nil { return false }
    ln.Close()
    return true
}
```

**Test case pattern** (from `postgres/manager_test.go` lines 138–162 — assert.True/False, require.Error/NoError, AssertExpectations):
```go
func TestIsInstalled_FalseWhenBinaryAbsent(t *testing.T) {
    mgr, mfs, _, paths := newTestManager(t)
    mfs.On("Stat", filepath.Join(paths.PgwebDir(), "pgweb")).
        Return(nil, os.ErrNotExist)
    assert.False(t, mgr.IsInstalled())
    mfs.AssertExpectations(t)
}
```

---

### `internal/system/paths.go` (config, modification)

**Analog:** self — add alongside existing `PhpMyAdminDir()` pattern

**Pattern to add** (after `PostgreSQLHbaFile()` at line 190, following `PhpMyAdminDir()` at line 160):
```go
// PgwebDir returns the pgweb binary directory (~/.lamboserver/pgweb/).
func (p *Paths) PgwebDir() string { return filepath.Join(p.Home, "pgweb") }
```

**EnsureDirectories pattern** (lines 206–231 — add `p.PgwebDir()` to dirs slice, matching the `p.PhpMyAdminDir()` entry at line 223):
```go
// In EnsureDirectories() dirs slice, add after p.PostgreSQLSocketDir():
p.PgwebDir(),   // new: pgweb binary dir
```

---

### `internal/config/defaults.go` (config, modification)

**Analog:** self — add alongside `DefaultPhpMyAdminPort`

**Pattern to add** (after `DefaultPhpMyAdminPort = 8088` at line 27):
```go
// DefaultPgwebPort is the default HTTP port for the pgweb web admin tool.
DefaultPgwebPort = 8081
```

---

### `app.go` — struct, NewApp, shutdown, StopService, new IPC methods (provider/IPC, modification)

**Analog:** self — modifications follow exact patterns of adjacent code

**App struct addition** (after `PhpMyAdmin *phpmyadmin.Manager` at line 48):
```go
// In App struct, add after PhpMyAdmin field:
Pgweb *pgweb.Manager
```

**NewApp() — manager creation pattern** (from `app.go` lines 71–72, following postgres pattern):
```go
// In NewApp(), after postgresMgr creation (line 71):
pgwebMgr := pgweb.NewManager(paths, pgweb.RealFS{}, pgweb.RealCmdRunner{})

// In App struct literal (after PhpMyAdmin: phpMyAdminMgr, line 99):
Pgweb: pgwebMgr,
```

**shutdown() pattern** (from `app.go` lines 167–178 — add Pgweb.Stop() before PostgreSQL.Stop()):
```go
// In shutdown(), add before a.PostgreSQL.Stop() (line 170):
a.Pgweb.Stop()  // D-08: must stop before PostgreSQL (pgweb depends on it)
```

**StopService() auto-stop coupling** (from `app.go` lines 265–274 — add pgweb coupling after err assignment):
```go
func (a *App) StopService(name string) error {
    a.Debug.Info("StopService called: %s", name)
    svc, err := a.Manager.Get(name)
    if err != nil {
        return err
    }
    err = svc.Stop()
    // D-08: pgweb must stop when PostgreSQL stops (it cannot function without PG).
    if err == nil && name == "postgresql" {
        _ = a.Pgweb.Stop()
    }
    a.Debug.Action(fmt.Sprintf("StopService(%s)", name), err)
    return err
}
```

**New IPC methods pattern** (from `app.go` lines 449–473 — Web Admin section, debug logging style):
```go
// ── pgweb ─────────────────────────────────────────────────────────────

// PgwebStatus is the frontend-facing status snapshot for pgweb.
type PgwebStatus struct {
    Installed bool `json:"installed"`
    Running   bool `json:"running"`
    Port      int  `json:"port"`
}

func (a *App) GetPgwebStatus() PgwebStatus {
    s := a.Pgweb.Status()
    return PgwebStatus{Installed: s.Installed, Running: s.Running, Port: s.Port}
}

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

func (a *App) OpenPgweb() error {
    a.Debug.Info("OpenPgweb called")
    return browser.OpenURL("http://127.0.0.1:8081")
}
```

**Debug.Action call pattern** (from `app.go` lines 260–262 — single-arg string, not fmt.Sprintf):
```go
// Simple name: a.Debug.Action("InstallPgweb", err)
// Compound name: a.Debug.Action(fmt.Sprintf("StopService(%s)", name), err)
```

---

### `frontend/src/pages/PgDatabasePage.tsx` (component, request-response, modification)

**Analog:** self — modifications follow existing state/handler/JSX patterns

**New IPC import additions** (from `PgDatabasePage.tsx` lines 5–13 — import block):
```tsx
import {
  StartService,
  StopService,
  GetAllStatuses,
  InstallVersion,
  InitService,
  ListServiceDatabases,
  CreateServiceDatabase,
  DropServiceDatabase,
  // ADD:
  GetPgwebStatus,
  InstallPgweb,
  StartPgweb,
  StopPgweb,
  OpenPgweb,
} from "../../wailsjs/go/main/App";
```

**New state variables pattern** (from `PgDatabasePage.tsx` lines 21–32 — useState declarations):
```tsx
// Add after existing state declarations (line 32):
const [pgwebInstalled, setPgwebInstalled] = useState(false);
const [pgwebRunning, setPgwebRunning]     = useState(false);
const [pgwebLoading, setPgwebLoading]     = useState(false);
const [pgwebInstallStep, setPgwebInstallStep] = useState<string | null>(null);
const [pgwebError, setPgwebError]         = useState<string | null>(null);
```

**loadStatus() extension pattern** (from `PgDatabasePage.tsx` lines 38–56 — extend existing try block):
```tsx
const loadStatus = async () => {
  try {
    const statuses = await GetAllStatuses();
    const pgStatus = statuses["postgresql"] || "not_installed";
    const isRunning = pgStatus === "running";
    setRunning(isRunning);
    setInstalled(pgStatus !== "not_installed");

    // NEW: load pgweb status independently (dedicated IPC call per RESEARCH.md Pattern 4)
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

**Handler pattern** (from `PgDatabasePage.tsx` lines 74–95 / 97–108 — install+start/stop handler shape):
```tsx
// Install handler — mirrors handleInstall() pattern (lines 74-95)
const handleInstallPgweb = async () => {
  setPgwebLoading(true);
  setPgwebError(null);
  try {
    setPgwebInstallStep("Downloading...");
    await InstallPgweb();
    setPgwebInstallStep("\u2713 Installed");
    setTimeout(async () => {
      setPgwebInstallStep(null);
      setPgwebLoading(false);
      await loadStatus();
    }, 1200);
    return;
  } catch (e: any) {
    console.error(e);
    setPgwebInstallStep(null);
    setPgwebError(String(e));
  }
  setPgwebLoading(false);
};

// Start handler — mirrors handleStart() pattern (lines 97-108)
const handleStartPgweb = async () => {
  setPgwebLoading(true);
  setPgwebError(null);
  try {
    await StartPgweb();
    await loadStatus();
  } catch (e: any) {
    console.error(e);
    setPgwebError(String(e));
  }
  setPgwebLoading(false);
};

// Stop handler — mirrors handleStop() pattern (lines 110-122)
const handleStopPgweb = async () => {
  setPgwebLoading(true);
  setPgwebError(null);
  try {
    await StopPgweb();
    await loadStatus();
  } catch (e: any) {
    console.error(e);
    setPgwebError(String(e));
  }
  setPgwebLoading(false);
};

const handleOpenPgweb = () => { OpenPgweb(); };
```

**Web Admin card JSX pattern** (from `PgDatabasePage.tsx` lines 232–254 — service-row card shape; insert between PostgreSQL service card and Create Database card):
```tsx
{/* Web Admin (pgweb) card — only shown when PostgreSQL is running */}
{installed && running && (
  <>
    {/* ... existing PostgreSQL service card ends ... */}

    {/* Web Admin card — position: after PG service card, before Create Database */}
    <div className="card" style={{ marginBottom: 16 }}>
      <div className="service-row">
        <div className="service-info">
          {pgwebInstalled && (
            <span className={`status-dot ${pgwebRunning ? "running" : "stopped"}`} />
          )}
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
            <button
              className="btn btn-primary btn-sm"
              onClick={handleInstallPgweb}
              disabled={pgwebLoading}
            >
              {pgwebInstallStep ?? "Install pgweb"}
            </button>
          )}
          {pgwebInstalled && !pgwebRunning && (
            <button
              className="btn btn-primary btn-sm"
              onClick={handleStartPgweb}
              disabled={pgwebLoading}
            >
              {pgwebLoading ? "Starting..." : "Start"}
            </button>
          )}
          {pgwebInstalled && pgwebRunning && (
            <>
              <button
                className="btn btn-danger btn-sm"
                onClick={handleStopPgweb}
                disabled={pgwebLoading}
              >
                {pgwebLoading ? "Stopping..." : "Stop"}
              </button>
              <button
                className="btn btn-secondary btn-sm"
                onClick={handleOpenPgweb}
              >
                Open
              </button>
            </>
          )}
        </div>
      </div>
    </div>

    {/* ... existing Create Database card continues ... */}
  </>
)}
```

---

## Shared Patterns

### osFileSystem / systemCommandRunner Production Implementations
**Source:** `internal/services/postgres/manager.go` lines 22–39 and `internal/services/phpmyadmin/manager.go` lines 24–40
**Apply to:** `internal/services/pgweb/manager.go`
```go
type osFileSystem struct{}
func (osFileSystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

type systemCommandRunner struct{}
func (systemCommandRunner) Run(name string, args ...string) (string, error) {
    return system.RunCommand(name, args...)
}
```
Rename as `RealFS` / `RealCmdRunner` inside the pgweb package (or use the same private type convention as postgres — either is acceptable).

### Error Wrapping
**Source:** `internal/services/postgres/manager.go` lines 122–125, 198–200, 238–240
**Apply to:** All `internal/services/pgweb/manager.go` methods
```go
return fmt.Errorf("<short action description>: %w", err)
```
Convention: always use `%w` (not `%v`) for wrapping so callers can `errors.Is()` on the root cause.

### Port Preflight Check
**Source:** `internal/services/postgres/manager.go` lines 279–286
**Apply to:** `pgweb.Manager.checkPort()` — copy verbatim, substitute port 8081 for 5432
```go
func (m *Manager) checkPort() error {
    ln, err := net.Listen("tcp", "127.0.0.1:8081")
    if err != nil {
        return fmt.Errorf("port 8081 is already in use by another process")
    }
    ln.Close()
    return nil
}
```

### Debug Logging in IPC Methods
**Source:** `app.go` lines 253–262 (StartService), 449–459 (InstallWebAdmin)
**Apply to:** All new IPC methods in `app.go` (InstallPgweb, StartPgweb, StopPgweb, OpenPgweb, GetPgwebStatus)
```go
func (a *App) <Method>() error {
    a.Debug.Info("<Method> called")
    err := a.<Manager>.<Action>()
    a.Debug.Action("<Method>", err)
    return err
}
```

### ServiceAdapter Compile-Time Interface Check
**Source:** `internal/services/postgres/service_adapter.go` line 19, `internal/services/phpmyadmin/service_adapter.go` line 19
**Apply to:** `internal/services/pgweb/manager.go` (or a separate `service_adapter.go` if one is created)
```go
// Compile-time interface check.
var _ DaemonWebAdminService = (*Manager)(nil)
```

### Toast Error Parsing in Frontend
**Source:** `frontend/src/pages/PgDatabasePage.tsx` lines 68–72
**Apply to:** `handleInstallPgweb` in `PgDatabasePage.tsx`
```tsx
const parseInstallError = (raw: string): ToastError => {
    const cleaned = raw.replace(/^Error invoking method "[^"]+": /, "");
    const firstLine = cleaned.split("\n")[0].slice(0, 120);
    return { summary: firstLine, detail: cleaned };
};
// Use setToastError(parseInstallError(String(e))) on install failure.
```

---

## No Analog Found

No files are without analog. All new files have close matches in the existing codebase.

The one novel element — in-memory `*exec.Cmd` process tracking with `sync.Mutex` — has no codebase analog but is a direct application of Go stdlib patterns (`os/exec`). The RESEARCH.md Pattern 2 provides the complete code template for this.

---

## Critical Anti-Patterns (from RESEARCH.md)

| Anti-Pattern | Source in Codebase | Correct Pattern |
|---|---|---|
| `cmd.Run(pgwebBin, ...)` for Start | `CommandRunner.Run()` uses `CombinedOutput()` — blocks forever | Use raw `os/exec.Command(...).Start()` |
| Register pgweb via `mgr.RegisterWebAdmin` | `GetAllStatuses()` returns only `"installed"`/`"not_installed"` for web admin — no `"running"` state | Hold `*pgweb.Manager` directly on `App` struct; use `GetPgwebStatus()` |
| Omit `cmd.Wait()` after `Kill()` | Process becomes zombie | Always `_ = m.proc.Wait()` after `m.proc.Process.Kill()` |
| Omit `--skip-open` flag | pgweb opens browser itself on Start | Always pass `--skip-open` in exec args |

---

## Metadata

**Analog search scope:** `internal/services/`, `internal/system/`, `internal/config/`, `app.go`, `frontend/src/pages/`
**Files scanned:** 11 source files read
**Pattern extraction date:** 2026-04-16
