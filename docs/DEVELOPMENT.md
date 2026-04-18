# LamboServer — Developer Guide

LamboServer is a macOS desktop application built with [Wails v2](https://wails.io). It
manages local development infrastructure — PHP, Node.js, Nginx, and DNS — from a single
native window. The backend is Go; the UI is React.

This guide covers: prerequisites, project structure, how to build and run the app, how to
run tests, core architectural concepts, and a step-by-step walkthrough for adding a new
service manager.

For architectural decisions, see the ADR files:
- [ADR-001: Interface Adoption for Testability](adr/ADR-001-interface-adoption.md)
- [ADR-002: Bottom-Up Test Ordering](adr/ADR-002-bottom-up-test-ordering.md)
- [ADR-003: App Orchestration Layer Split](adr/ADR-003-app-layer-split.md)

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.23+ | Backend compilation |
| Wails CLI | v2.x | Build and dev server (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) |
| Node.js | 18+ | Frontend development only (Vite bundler) |
| npm | 9+ | Frontend dependency management |
| Xcode CLI Tools | Latest | macOS compilation (`xcode-select --install`) |
| macOS | 12+ | Runtime — macOS only (launchd, osascript) |

Install the Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:

```bash
wails doctor
```

---

## Project Structure

```
lamboserver/
├── main.go              # Wails bootstrap: embed frontend/dist, create App, configure runtime
├── app.go               # App struct definition, NewApp() composition root, lifecycle hooks
├── app_compat.go        # Legacy Wails bindings (backward-compatible, to be removed after frontend migration)
├── doc.go               # Package documentation
├── wails.json           # Wails app name, output filename, frontend dev/build commands
├── go.mod, go.sum       # Go module and dependency lock
│
├── internal/
│   ├── system/          # Platform abstractions (macOS-specific code lives here)
│   │   ├── paths.go     # All file/directory locations (single source of truth)
│   │   ├── launchd.go   # LaunchDaemon lifecycle via launchctl
│   │   ├── darwin.go    # osascript, admin elevation (Darwin-only)
│   │   ├── shell.go     # PATH injection into .bashrc/.zshrc
│   │   ├── binary.go    # Binary discovery in PATH or local dirs
│   │   ├── realfs.go    # RealFS: FileSystem interface backed by os package
│   │   ├── realcmd.go   # RealCmdRunner, RealAdminRunner: exec interface implementations
│   │   ├── download.go  # HTTP download and archive extraction
│   │   ├── helper.go    # Privileged helper script management
│   │   └── embed.go     # Access bundled binary resources
│   │
│   ├── config/          # Thread-safe config store
│   │   ├── store.go     # AppConfig struct, JSON persistence, sync.RWMutex
│   │   └── defaults.go  # Default port assignments for all services
│   │
│   ├── process/         # Process and launchd management utilities
│   │   ├── runner.go    # Process runner abstraction
│   │   ├── pidfile.go   # PID file management
│   │   ├── port.go      # Port availability checking
│   │   ├── launchd.go   # Launchd helpers
│   │   └── plist.go     # Plist template generation
│   │
│   ├── binaries/        # Binary registry and download management
│   │   ├── registry.go  # Binary version registry
│   │   └── downloader.go # HTTP download utilities
│   │
│   ├── services/        # Unified service interfaces and managers
│   │   ├── service.go   # Service, VersionedService, WebAdminService interfaces
│   │   ├── manager.go   # Central orchestrator — lookup services by name
│   │   │
│   │   ├── php/         # PHP version management
│   │   │   ├── manager.go           # ListInstalled, SetActive, Install, Uninstall
│   │   │   ├── detector.go          # Detect PHP from Homebrew, System, Herd, Manual
│   │   │   ├── fpm.go               # PHP-FPM daemon lifecycle
│   │   │   ├── service_adapter.go   # Adapts Manager to services.VersionedService
│   │   │   ├── interfaces.go        # FileSystem, CommandRunner, LaunchdService
│   │   │   └── errors.go            # Domain error types
│   │   │
│   │   ├── nginx/       # Nginx web server
│   │   │   ├── manager.go           # Install, Uninstall, Start, Stop, Restart, Status
│   │   │   ├── config.go            # Nginx configuration template generation
│   │   │   ├── service_adapter.go   # Adapts Manager to services.Service
│   │   │   ├── interfaces.go        # FileSystem, LaunchdService, HelperRunner
│   │   │   └── errors.go            # Domain error types
│   │   │
│   │   ├── dnsmasq/     # DNS resolver
│   │   │   ├── manager.go           # Install, configure /etc/resolver/test, lifecycle
│   │   │   ├── service_adapter.go   # Adapts Manager to services.Service
│   │   │   ├── interfaces.go        # FileSystem, LaunchdService, AdminRunner
│   │   │   └── errors.go            # Domain error types
│   │   │
│   │   ├── nodejs/      # Node.js version management
│   │   │   ├── manager.go           # ListInstalled, Install, SetActive
│   │   │   ├── service_adapter.go   # Adapts Manager to services.VersionedService
│   │   │   └── interfaces.go        # FileSystem, CommandRunner
│   │   │
│   │   ├── mysql/       # MySQL database
│   │   │   ├── manager.go           # Install, Start, Stop, Status, database ops
│   │   │   ├── service_adapter.go   # Adapts Manager to services.Service
│   │   │   └── interfaces.go        # FileSystem, CommandRunner
│   │   │
│   │   ├── postgres/    # PostgreSQL database
│   │   │   ├── manager.go           # Install, Start, Stop, Status, database ops
│   │   │   ├── service_adapter.go   # Adapts Manager to services.Service
│   │   │   └── interfaces.go        # FileSystem, CommandRunner
│   │   │
│   │   ├── phpmyadmin/  # phpMyAdmin web admin
│   │   │   ├── manager.go           # Install, Uninstall, Status, URL
│   │   │   ├── service_adapter.go   # Adapts Manager to services.WebAdminService
│   │   │   └── interfaces.go        # FileSystem
│   │   │
│   │       ├── config.go            # Connection preset configuration
│   │       ├── installer.go         # Binary download and setup
│   │       └── service_adapter.go   # Adapts Manager to services.Service
│   │
│   ├── sites/           # Site (domain) management
│   │   ├── interfaces.go    # FileSystem interface
│   │   └── manager.go       # Create/destroy Nginx site configs, SSL cert linking
│   │
│   ├── cert/            # SSL certificate management
│   │   ├── interfaces.go    # FileSystem, AdminRunner interfaces
│   │   └── ca.go            # Generate local CA, per-site certificates
│   │
│   └── integration/     # Integration tests
│       └── *_test.go    # End-to-end workflow tests
│
├── pkg/
│   └── logger/          # Debug logger and log reader
│       ├── logger.go    # File-based logger: Info, Error, Action methods
│       └── reader.go    # List log files, read last N lines
│
└── frontend/            # React UI (out of scope for backend work)
    ├── src/
    │   ├── main.tsx     # React entry point
    │   ├── App.tsx      # Main layout with sidebar navigation
    │   └── pages/       # Six page components
    └── dist/            # Built assets (generated by Vite, embedded in binary)
```

---

## Build and Run

### Development Mode

```bash
wails dev
```

This starts:
- The Go backend with hot-reload on `.go` file changes
- The Vite dev server for the frontend with HMR

The app window opens automatically. Backend logs go to the terminal.

### Production Build

```bash
wails build
```

Output: `build/bin/LamboServer.app` (macOS app bundle)

The frontend is compiled by Vite (`npm run build`) and embedded in the Go binary via
`//go:embed all:frontend/dist` in `main.go`. No external files are needed at runtime.

**Note:** macOS only. The app uses launchd, osascript, and macOS-specific paths.
Cross-platform builds are out of scope.

---

## Running Tests

Run all tests:

```bash
go test ./...
```

Run with race detection (required for config store tests):

```bash
go test -race ./...
```

Run a specific package:

```bash
go test ./internal/config/...
go test ./internal/services/php/...
```

Run with coverage:

```bash
go test -cover ./...
```

Run a specific test by name:

```bash
go test -run TestStore_ConcurrentReadWrite ./internal/config/...
```

**Note:** Tests use interface fakes, not real system services. No Nginx, PHP, or launchd
needed to run tests. Tests pass on any macOS machine.

---

## Architecture Concepts

### Interface-Based Dependency Injection

Every service manager receives its external dependencies (filesystem, exec, launchd) through
constructor arguments, not globals or direct `os` calls.

Each package defines its own interfaces in `interfaces.go` — only the methods that package
needs. This is the Go convention: define interfaces at the consumer site.

```go
// internal/services/php/interfaces.go
type FileSystem interface {
    Stat(name string) (fs.FileInfo, error)
    ReadDir(name string) ([]fs.DirEntry, error)
    MkdirAll(path string, perm fs.FileMode) error
    WriteFile(name string, data []byte, perm fs.FileMode) error
    // ...
}
```

In production, `system.RealFS{}` satisfies this interface via the `os` package.
In tests, a fake struct is defined in `testutil_test.go`.

See [ADR-001](adr/ADR-001-interface-adoption.md) for the full rationale.

### Manager Pattern

Each service has a `*Manager` struct with a `NewManager()` constructor. The pattern:

```go
// Constructor accepts dependencies as interfaces
func NewManager(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager {
    return &Manager{paths: paths, store: store, fs: fs, cmd: cmd}
}

// Lifecycle methods
func (m *Manager) IsInstalled() bool { ... }
func (m *Manager) Install() error { ... }
func (m *Manager) Uninstall() error { ... }
func (m *Manager) Start() error { ... }
func (m *Manager) Stop() error { ... }
func (m *Manager) Status() ServiceStatus { ... }
```

Not all managers implement all methods — only what the domain requires.

### Composition Root in NewApp()

`NewApp()` in `app.go` is the single place that instantiates real implementations and wires
them into all managers. No other code knows which concrete type backs a `FileSystem` interface.

```go
func NewApp() *App {
    paths := system.NewPaths()
    store := config.NewStore(paths.ConfigFile())
    launchd := system.NewLaunchdManager(paths)
    fs   := system.RealFS{}
    cmd  := system.RealCmdRunner{}
    admin := system.RealAdminRunner{}
    certMgr := cert.NewManager(paths, fs, admin)

    return &App{
        Php:    php.NewManager(paths, store, fs, cmd),
        Nginx:  nginx.NewManager(paths, launchd, fs, launchd.Helper()),
        // ...
    }
}
```

If you need to add a new dependency or swap an implementation, this is the only file to change.

### Wails Binding

Wails auto-generates TypeScript bindings from all public methods on the `App` struct. A Go
method:

```go
func (a *App) GetPhpVersions() ([]php.PhpVersion, error)
```

becomes callable from React as:

```typescript
import { GetPhpVersions } from "../../wailsjs/go/main/App"
const versions = await GetPhpVersions()
```

**Constraint:** Method signatures on `App` must not change. Changing a signature breaks the
auto-generated bindings and the frontend.

See [ADR-003](adr/ADR-003-app-layer-split.md) for how App methods are organized.

---

## Adding a New Service Manager

This walkthrough adds a hypothetical `myservice` manager. Use `internal/services/php/` as the
canonical reference for each step.

### Step 1: Create the package directory

```bash
mkdir internal/services/myservice
```

### Step 2: Define interfaces in `interfaces.go`

Only include the methods your manager actually needs. Keep interfaces small.

```go
// internal/services/myservice/interfaces.go
package myservice

import (
    "io/fs"
    "github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// FileSystem is the filesystem subset that myservice.Manager needs.
type FileSystem interface {
    Stat(name string) (fs.FileInfo, error)
    MkdirAll(path string, perm fs.FileMode) error
    WriteFile(name string, data []byte, perm fs.FileMode) error
    Remove(name string) error
}

// LaunchdService is the launchd subset that myservice.Manager needs.
type LaunchdService interface {
    Install(cfg system.ServiceConfig) error
    Uninstall(cfg system.ServiceConfig) error
    IsRunning(label string) bool
}
```

### Step 3: Define error types in `errors.go`

```go
// internal/services/myservice/errors.go
package myservice

import "errors"

var (
    ErrNotInstalled = errors.New("myservice: not installed")
    ErrAlreadyRunning = errors.New("myservice: already running")
)
```

### Step 4: Create `manager.go` with `Manager` struct and `NewManager()`

```go
// internal/services/myservice/manager.go
package myservice

import "github.com/rizkirmdhnnn/lamboserver/internal/system"

const ServiceLabel = "com.lamboserver.myservice"

// Manager manages the myservice lifecycle.
type Manager struct {
    paths   *system.Paths
    launchd LaunchdService
    fs      FileSystem
}

// NewManager creates a new Manager with the given dependencies.
func NewManager(paths *system.Paths, launchd LaunchdService, fs FileSystem) *Manager {
    return &Manager{paths: paths, launchd: launchd, fs: fs}
}

// IsInstalled reports whether myservice binary is present.
func (m *Manager) IsInstalled() bool {
    _, err := m.fs.Stat(m.paths.MyServiceBin())
    return err == nil
}

// Install downloads and installs myservice.
func (m *Manager) Install() error {
    // ... implementation
    return nil
}

// Start launches myservice via launchd.
func (m *Manager) Start() error {
    // ... implementation
    return nil
}

// Stop halts the running myservice daemon.
func (m *Manager) Stop() {
    // ... implementation
}

// Status reports whether myservice is running.
func (m *Manager) Status() system.ServiceStatus {
    return system.ServiceStatus{Running: m.launchd.IsRunning(ServiceLabel)}
}
```

### Step 5: Add a package doc comment in `doc.go`

```go
// Package myservice manages the myservice daemon lifecycle.
package myservice
```

### Step 6: Add a path helper in `internal/system/paths.go`

Add any path accessors your manager needs:

```go
func (p *Paths) MyServiceBin() string {
    return filepath.Join(p.BinDir(), "myservice")
}
```

### Step 7: Wire into `app.go`

Add a field to the `App` struct:

```go
type App struct {
    // existing fields ...
    MyService *myservice.Manager
}
```

Instantiate in `NewApp()`:

```go
func NewApp() *App {
    // existing wiring ...
    return &App{
        // existing managers ...
        MyService: myservice.NewManager(paths, launchd, fs),
    }
}
```

### Step 8: Create a service adapter and register in `app.go`

Create `internal/services/myservice/service_adapter.go` to adapt your Manager to the
unified `services.Service` interface:

```go
// internal/services/myservice/service_adapter.go
package myservice

import "github.com/rizkirmdhnnn/lamboserver/internal/services"

type serviceAdapter struct{ mgr *Manager }

func NewServiceAdapter(mgr *Manager) services.Service { return &serviceAdapter{mgr: mgr} }

func (a *serviceAdapter) Install(version string) error   { return a.mgr.Install() }
func (a *serviceAdapter) Start() error                   { return a.mgr.Start() }
func (a *serviceAdapter) Stop() error                    { a.mgr.Stop(); return nil }
func (a *serviceAdapter) Restart() error                 { a.mgr.Stop(); return a.mgr.Start() }
func (a *serviceAdapter) Status() services.ServiceStatus { /* map from mgr.Status() */ }
func (a *serviceAdapter) Logs() ([]string, error)        { return nil, nil }
func (a *serviceAdapter) Version() string                { return "" }
```

Then register it in `NewApp()` in `app.go`:

```go
mgr.Register("myservice", myservice.NewServiceAdapter(myServiceMgr))
```

For backward-compatible legacy methods, add them in `app_compat.go`.

### Step 9: Write tests in `manager_test.go`

Create `internal/services/myservice/testutil_test.go` with fake implementations:

```go
// internal/services/myservice/testutil_test.go
package myservice

import (
    "io/fs"
    "github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// fakeFS is a test fake for the FileSystem interface.
type fakeFS struct {
    statErr error
}

func (f *fakeFS) Stat(name string) (fs.FileInfo, error) {
    return nil, f.statErr
}

func (f *fakeFS) MkdirAll(path string, perm fs.FileMode) error { return nil }
func (f *fakeFS) WriteFile(name string, data []byte, perm fs.FileMode) error { return nil }
func (f *fakeFS) Remove(name string) error { return nil }

// fakeLaunchd is a test fake for the LaunchdService interface.
type fakeLaunchd struct {
    running bool
}

func (f *fakeLaunchd) Install(cfg system.ServiceConfig) error  { return nil }
func (f *fakeLaunchd) Uninstall(cfg system.ServiceConfig) error { return nil }
func (f *fakeLaunchd) IsRunning(label string) bool             { return f.running }
```

Write table-driven tests in `internal/services/myservice/manager_test.go`:

```go
// internal/services/myservice/manager_test.go
package myservice

import (
    "errors"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/rizkirmdhnnn/lamboserver/internal/system"
)

func TestManager_IsInstalled(t *testing.T) {
    paths := system.NewPathsWithRoot(t.TempDir())
    tests := []struct {
        name        string
        statErr     error
        wantInstalled bool
    }{
        {"binary present", nil, true},
        {"binary missing", errors.New("not found"), false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := NewManager(paths, &fakeLaunchd{}, &fakeFS{statErr: tt.statErr})
            assert.Equal(t, tt.wantInstalled, m.IsInstalled())
        })
    }
}
```

### Step 10: Verify

```bash
go build ./...      # must compile
go vet ./...        # must pass
go test ./internal/services/myservice/...  # must pass
```

---

## Conventions

Follow the patterns in `internal/services/php/` as the canonical example. Key rules:

**Naming:**
- Exported identifiers: PascalCase (`NewManager`, `IsInstalled`, `ServiceStatus`)
- Unexported: camelCase (`mu`, `paths`, `ensureConfig`)
- Constants: UPPER_CASE (`ServiceLabel`, `TestTLD`)

**Errors:**
- Wrap with `%w` for chain preservation: `fmt.Errorf("start myservice: %w", err)`
- Define domain sentinel errors in `errors.go`

**Logging:**
- Entry: `a.Debug.Info("MethodName called")`
- Completion: `a.Debug.Action("MethodName", err)` (logs OK or FAILED)
- Infrastructure errors: `a.Debug.Error("detail: %v", err)`

**Functions:**
- Keep compact (10–50 lines preferred)
- Return errors explicitly — no panic except in `init()` or unreachable paths
- Pointer receivers for all methods: `(m *Manager)`

For the full conventions reference, see `CLAUDE.md` at the project root.

---

## Useful Commands

| Command | Description |
|---------|-------------|
| `wails dev` | Start dev server with hot-reload |
| `wails build` | Build production app bundle |
| `wails doctor` | Verify Wails/Go/Node install |
| `go build ./...` | Compile all Go packages |
| `go vet ./...` | Static analysis |
| `go test ./...` | Run all tests |
| `go test -race ./...` | Run tests with race detector |
| `go test -cover ./...` | Run tests with coverage report |
| `go test -run TestName ./internal/pkg/...` | Run a specific test |
| `go mod tidy` | Clean up go.mod/go.sum |
| `cat ~/.lamboserver/logs/debug.log` | Tail the debug log |
