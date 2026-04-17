# Architecture

**Analysis Date:** 2026-04-17

## Pattern Overview

**Overall:** Desktop monolith using Wails (Go backend + React frontend) with dependency injection and service adapter pattern.

**Key Characteristics:**
- Composition root pattern: all dependency wiring happens in `NewApp()` in `app.go`
- Go backend exposes public methods to React frontend via Wails IPC bindings (no REST API)
- Interface-based dependency injection with consumer-defined interfaces (Go convention D-02)
- Service adapter pattern: each service has a domain-specific Manager + a ServiceAdapter that implements unified `services.Service` or `services.VersionedService`
- macOS launchd for daemon lifecycle (nginx, dnsmasq, mysql run as system LaunchDaemons via a privileged helper)
- All application state stored under `~/.lamboserver/` (config, binaries, certs, logs)

## Layers

**Presentation Layer (Frontend):**
- Purpose: React SPA providing the desktop GUI
- Location: `frontend/src/`
- Contains: Pages, components, CSS, Wails-generated TypeScript bindings
- Depends on: Wails IPC bindings (`frontend/wailsjs/go/main/App.js`)
- Used by: End user via native macOS window

**IPC Binding Layer (App struct):**
- Purpose: Thin glue between frontend and backend; validates input, delegates to managers, returns results
- Location: `app.go`
- Contains: All public methods on the `App` struct (auto-exposed to frontend by Wails)
- Depends on: All internal service managers
- Used by: Frontend via Wails-generated TypeScript functions

**Service Manager Layer:**
- Purpose: Domain-specific business logic for each managed service
- Location: `internal/services/*/manager.go`
- Contains: Installation, configuration generation, start/stop/restart, version management
- Depends on: `internal/system` (paths, launchd, fs, cmd), `internal/config`, `internal/cert`
- Used by: `App` struct methods in `app.go`

**Unified Service Registry:**
- Purpose: Generic service dispatch by name (e.g., `StartService("nginx")`)
- Location: `internal/services/service.go`
- Contains: `Service`, `VersionedService`, `WebAdminService` interfaces; `Manager` registry
- Depends on: Service adapters
- Used by: `App.StartService()`, `App.StopService()`, etc.

**System Abstraction Layer:**
- Purpose: OS-specific operations (launchd, shell commands, privilege elevation, filesystem)
- Location: `internal/system/`
- Contains: `Paths`, `LaunchdManager`, `Helper`, `RealFS`, `RealCmdRunner`, `RealAdminRunner`, `Integration`
- Depends on: macOS system APIs (launchctl, osascript, pgrep)
- Used by: All service managers

**Configuration Layer:**
- Purpose: Persistent, thread-safe application config
- Location: `internal/config/store.go`
- Contains: `AppConfig` struct, `Store` with `sync.RWMutex`
- Depends on: Filesystem (JSON file at `~/.lamboserver/config.json`)
- Used by: All service managers, `App` struct

**Supporting Packages:**
- Purpose: Cross-cutting utilities
- Location: `pkg/logger/`, `pkg/notify/`, `internal/binaries/`, `internal/cert/`, `internal/process/`
- Contains: Debug logging, notifications, binary download registry, SSL cert generation, process abstractions

## Data Flow

**Frontend calls Go backend (e.g., starting a service):**

1. React component calls `StartService("nginx")` from `frontend/wailsjs/go/main/App.js`
2. Wails IPC deserializes args and invokes `App.StartService("nginx")` in `app.go`
3. `App.StartService` looks up the service adapter via `a.Manager.Get("nginx")`
4. The adapter's `Start()` method delegates to `nginx.Manager.Start()`
5. `nginx.Manager` generates config, writes plist, and calls `LaunchdManager.Install()` 
6. `LaunchdManager` uses `Helper.Run("install-daemon", ...)` which runs `sudo lambo-helper install-daemon`
7. Result (error or nil) propagates back through the chain to the frontend as a Promise resolution/rejection

**Service status polling:**

1. Frontend calls `GetDashboardStatus()` via Wails binding
2. `App.GetDashboardStatus()` queries each manager's `Status()` method
3. Each manager checks launchd/process state (e.g., `launchctl list`, `pgrep`)
4. Aggregated `DashboardStatus` struct returned to frontend as JSON

**Version installation (PHP/Node.js):**

1. Frontend calls `InstallVersion("php", "8.3")`
2. `App.InstallVersion` routes to `VersionedService.Install()` via service registry
3. PHP adapter delegates to `php.Manager.Install()` which downloads from herdphp.com
4. Binary extracted to `~/.lamboserver/php/8.3/`
5. Shell symlinks updated, PHP-FPM restarted

**State Management:**
- Backend state: `config.Store` (in-memory cache + JSON file, protected by `sync.RWMutex`)
- Frontend state: React `useState` hooks per page (no global state manager like Redux)
- Service state: Derived at query time from launchd/process inspection (not cached)

## Key Abstractions

**Service Interface Hierarchy:**
- Purpose: Unified API for heterogeneous services
- Examples: `internal/services/service.go`
- Pattern: Three-tier interface hierarchy:
  - `Service` - base (Install, Start, Stop, Restart, Status, Logs, Version)
  - `VersionedService` - extends Service (SwitchVersion, InstalledVersions, ActiveVersion)
  - `WebAdminService` - separate (Install, URL, IsInstalled, Version)

**Service Adapter Pattern:**
- Purpose: Bridges domain-specific Managers to unified Service interfaces
- Examples: `internal/services/php/service_adapter.go`, `internal/services/nodejs/service_adapter.go`, `internal/services/phpmyadmin/service_adapter.go`
- Pattern: Adapter composes one or more Managers and implements the Service/VersionedService interface. Compile-time check via `var _ services.VersionedService = (*ServiceAdapter)(nil)`

**Consumer-Defined Interfaces (D-02):**
- Purpose: Each package defines only the interface methods it needs, not a god interface
- Examples: `internal/sites/interfaces.go` (FileSystem with 3 methods), `internal/cert/interfaces.go` (FileSystem with 5 methods), `internal/services/nginx/` (separate LaunchdService, FileSystem, HelperRunner)
- Pattern: Real implementations (`RealFS`, `RealCmdRunner`, `RealAdminRunner`) in `internal/system/` satisfy all consumer interfaces via structural typing

**Paths (Single Source of Truth for Locations):**
- Purpose: Centralizes all file/directory path resolution
- Examples: `internal/system/paths.go`
- Pattern: All paths derive from `~/.lamboserver/`. Any path change only requires editing `paths.go`.

**Privileged Helper:**
- Purpose: Avoid repeated password prompts for root operations
- Examples: `internal/system/helper.go`
- Pattern: One-time sudoers setup, then `sudo -n lambo-helper <command>` for daemon management

## Entry Points

**Application Entry Point:**
- Location: `main.go`
- Triggers: macOS app launch
- Responsibilities: Embeds frontend assets, creates App via `NewApp()`, configures Wails window options, calls `wails.Run()`

**App Lifecycle Hooks:**
- Location: `app.go` lines 112-183
- `startup()`: Ensures directories, installs helper, sets up CA, restores symlinks, starts services
- `beforeClose()`: Shows confirmation dialog
- `shutdown()`: Stops all services in dependency order, kills stale processes

**Frontend Entry Point:**
- Location: `frontend/src/main.tsx`
- Triggers: Wails WebView load
- Responsibilities: Renders React app root

## Error Handling

**Strategy:** Errors propagate from service managers through App methods to frontend as rejected Promises. No global error handler.

**Patterns:**
- Go errors returned directly from all public `App` methods; Wails converts to Promise rejections
- Startup errors logged via `a.Debug.Error()` but do not prevent app launch (graceful degradation)
- `a.Debug.Action(name, err)` pattern used consistently to log operation outcomes
- Frontend uses try/catch on async calls with `console.error` logging

## Cross-Cutting Concerns

**Logging:** Optional file-based debug logger (`pkg/logger/logger.go`). Toggled at runtime. Writes to `~/.lamboserver/logs/debug.log`. Service-specific logs in `~/.lamboserver/logs/`.

**Validation:** Minimal - input validation in App methods is thin. Service managers validate internally.

**Authentication:** Not applicable (local desktop app). Privilege elevation via macOS sudoers + osascript admin dialog.

**Process Management:** macOS launchd for nginx, dnsmasq, mysql. Direct process management for PHP-FPM, PostgreSQL, pgweb. Process abstraction defined in `internal/process/runner.go`.

**SSL/TLS:** Local CA generated and trusted in system keychain (`internal/cert/ca.go`). Per-site certificates auto-generated during site creation.

**DNS:** dnsmasq resolves `.test` TLD to 127.0.0.1 via `/etc/resolver/test`.

---

*Architecture analysis: 2026-04-17*
