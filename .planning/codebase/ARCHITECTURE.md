# Architecture

**Analysis Date:** 2026-04-16

## Pattern Overview

**Overall:** Layered Wails desktop application with unified service management

**Key Characteristics:**
- Composition-rooted dependency injection (App struct as wiring point)
- Adapter pattern for service abstraction (ServiceAdapter wraps managers to implement unified interfaces)
- Unidirectional IPC: Go backend methods exposed to React frontend via Wails binding
- Interface-driven design with testable dependencies (FileSystem, CmdRunner, AdminRunner interfaces)
- Centralized service registry (services.Manager) for dynamic service lookup and control

## Layers

**Presentation (Frontend):**
- Purpose: React-based desktop UI for managing services, sites, and infrastructure
- Location: `frontend/src/`
- Contains: React components (pages, components), styles, assets
- Depends on: Wails runtime bindings to App methods
- Used by: Desktop users via the Wails window

**IPC Bridge (Wails Binding):**
- Purpose: Expose Go backend methods to JavaScript frontend
- Location: `app.go` (public methods on App struct)
- Contains: All public methods on App are automatically bindable (StartService, StopService, InstallPhp, etc.)
- Depends on: Go backend (internal packages)
- Used by: Frontend TypeScript calls via Wails.Call()

**Application (Composition Root):**
- Purpose: Orchestrate lifecycle and service coordination
- Location: `app.go` (App struct and lifecycle methods), `main.go` (Wails.Run entry)
- Contains: Dependency injection (NewApp), lifecycle hooks (startup, beforeClose, shutdown)
- Depends on: All service managers, system utilities, logger
- Used by: Wails framework for lifecycle callbacks and IPC binding

**Service Management:**
- Purpose: Unified interface for daemon and web-admin services
- Location: `internal/services/` (manager.go, service.go, and per-service subdirectories)
- Contains: 
  - Manager: Registry mapping service names to implementations
  - Service interface: start/stop/restart/status/logs/version
  - VersionedService: extends Service with version switching
  - WebAdminService: install/URL/isInstalled/version (for phpMyAdmin, pgAdmin, CloudBeaver)
  - ServiceAdapter: Wraps individual managers to implement unified interfaces
- Depends on: Individual service managers
- Used by: App struct (for generic StartService/StopService), frontend (for service control)

**Domain Services (Service Managers):**
- Purpose: Implement specific service lifecycle and configuration
- Location: `internal/services/nginx/`, `internal/services/php/`, `internal/services/mysql/`, `internal/services/postgres/`, `internal/services/nodejs/`, `internal/services/dnsmasq/`, `internal/services/phpmyadmin/`, `internal/services/pgadmin/`, `internal/services/cloudbeaver/`
- Contains: Manager types (e.g., nginx.Manager) with business logic for each service
- Examples:
  - `internal/services/nginx/manager.go`: Nginx lifecycle and virtual host config generation
  - `internal/services/php/manager.go`: PHP version detection and PHP-FPM coordination
  - `internal/services/mysql/manager.go`: MySQL binary installation and daemon control
  - `internal/services/postgres/manager.go`: PostgreSQL binary installation and daemon control
  - `internal/services/nodejs/manager.go`: Node.js version switching
- Depends on: System abstractions (Paths, LaunchdManager, FileSystem, CmdRunner), config store
- Used by: ServiceAdapter (wrapped by adapter), App struct (direct access for manager-specific methods)

**Domain Features:**
- **Sites Management:** `internal/sites/manager.go` - Creates/destroys sites, generates Nginx virtual host configs, manages SSL certificates
- **Certificates:** `internal/cert/manager.go` - CA management, per-site certificate signing, Keychain trust
- **Configuration:** `internal/config/store.go` - Thread-safe persistent store for active PHP version, active Node version, site mappings, debug mode
- **System Abstraction:** `internal/system/` - Paths (single source of truth for ~/.lamboserver/), LaunchdManager (macOS launchctl), FileSystem, CmdRunner, AdminRunner abstractions
- **Logging:** `pkg/logger/` - Debug logging (disable by default, enable via config), service log file reader

## Data Flow

**Service Start (Frontend Initiation):**

1. User clicks "Start Nginx" in frontend UI
2. Frontend calls Wails-bound `App.StartService("nginx")`
3. App retrieves service from registry: `a.Manager.Get("nginx")` → returns `Service` interface
4. Calls `service.Start()` on the underlying adapter/manager
5. Manager implements platform-specific start logic (e.g., launchctl load for macOS daemons)
6. Response returns to frontend with success/error

**Site Creation:**

1. Frontend submits site creation form (domain, path)
2. Calls `App.CreateSite(domain, path)` (custom method in App)
3. Sites manager generates Nginx virtual host config from template
4. Cert manager signs SSL certificate for domain
5. Config store persists site record
6. Nginx is reloaded to pick up new virtual host
7. Resolver entry is registered with dnsmasq

**Service Startup Sequence (App Lifecycle):**

1. Wails framework calls `app.startup(ctx context.Context)`
2. Paths ensures ~/.lamboserver/ directories exist
3. Privileged helper (mxcl.PrivilegedHelper) is installed if needed
4. Local CA is setup and trusted via macOS Keychain
5. Shell integration (PATH) is added to ~/.bashrc/.zshrc
6. Stale LaunchAgent plist files are cleaned up
7. Services are restored (PHP/Node symlinks, MySQL/PostgreSQL services if config enables)
8. Critical services (Nginx, DNSMasq) are auto-started if not running

**State Management:**

- **In-Memory:** Cached in App struct (Config, Manager, all individual managers)
- **Persistent:** Config store in `~/.lamboserver/config.json` (JSON file, sync.RWMutex-protected reads/writes)
- **Process-Level:** Service daemons managed via macOS launchctl (plist files in ~/Library/LaunchDaemons or ~/Library/LaunchAgents)
- **File-System:** Binary installation in ~/.lamboserver/{php,nodejs,nginx}, configs in ~/.lamboserver/{nginx,dnsmasq,certs}

## Key Abstractions

**Service Interface:**
- Purpose: Unified contract for daemon services (Nginx, MySQL, PostgreSQL, CloudBeaver, DNSMasq)
- Examples: `internal/services/nginx/service_adapter.go`, `internal/services/php/service_adapter.go`
- Pattern: ServiceAdapter wraps manager-specific implementations to implement the unified interface
- Benefits: Generic service methods in App (StartService, StopService, RestartService) work for all services without conditional logic

**VersionedService Interface:**
- Purpose: Extends Service for runtimes supporting multiple installed versions (PHP, Node.js)
- Adds: SwitchVersion(), InstalledVersions(), ActiveVersion()
- Examples: PHP Manager implements VersionedService; Node Manager implements VersionedService
- Pattern: App.SwitchPhpVersion() looks up service via Manager.GetVersioned() and calls SwitchVersion()

**WebAdminService Interface:**
- Purpose: Web-based admin tools served through existing infrastructure (phpMyAdmin via Nginx+PHP-FPM)
- Adds: URL(), IsInstalled(), Install(version)
- Examples: `internal/services/phpmyadmin/`, `internal/services/pgadmin/`, `internal/services/cloudbeaver/`
- Pattern: RegisterWebAdmin() adds to separate webAdmin map in Manager; frontend can fetch URLs from App.GetWebAdminURL()

**FileSystem Interface:**
- Purpose: Testable file I/O abstraction
- Implementations: RealFS (actual filesystem), mock implementations in tests
- Pattern: Service managers depend on FileSystem interface; tests inject mock; real code injects RealFS{}

**CmdRunner Interface:**
- Purpose: Testable command execution abstraction
- Implementations: RealCmdRunner (actual execution), mock implementations in tests
- Pattern: Service managers depend on CmdRunner; tests mock; production uses RealCmdRunner{}

**AdminRunner Interface:**
- Purpose: Testable privilege elevation (sudo, osascript) abstraction
- Implementations: RealAdminRunner, mock implementations in tests
- Pattern: Cert manager and LaunchdManager depend on AdminRunner for root operations

**LaunchdService Interface:**
- Purpose: Testable macOS launchctl abstraction
- Usage: Service managers (Nginx, MySQL, PostgreSQL, PHP-FPM, DNSMasq) depend on LaunchdService to load/unload/start/stop daemons

## Entry Points

**Main Entry:**
- Location: `main.go`
- Triggers: OS launches executable (macOS app bundle entry point)
- Responsibilities: Create App instance via NewApp(), pass to Wails.Run() with asset embedding and lifecycle bindings

**Wails Startup Lifecycle:**
- Location: `app.go`, func (a *App) startup(ctx context.Context)
- Triggers: Wails framework calls on window creation after assets loaded
- Responsibilities: Ensure directories, install privileged helper, setup local CA, install shell integration, restore symlinks, auto-start critical services

**Frontend IPC Endpoints:**
- Location: Public methods on App struct in `app.go`
- Triggers: Frontend JavaScript calls via Wails.Call(methodName, ...args)
- Examples:
  - `StartService(name)`, `StopService(name)`, `RestartService(name)` — generic service control
  - `InstallPhp(version)`, `SwitchPhpVersion(version)`, `GetPhpVersions()` — PHP version management
  - `InstallNode(version)`, `SwitchNodeVersion(version)`, `GetNodeVersions()` — Node.js version management
  - `CreateSite(domain, path)`, `DeleteSite(domain)`, `GetSites()` — site management
  - `GetAllStatuses()` — poll service statuses for dashboard

## Error Handling

**Strategy:** Explicit error returns; App methods return error to frontend for user display

**Patterns:**
- Manager methods return `error` — caller decides whether to log, retry, or propagate
- App public methods return error — passed back to frontend as JSON response
- Debug logger tracks failures (Info, Error, Action levels)
- No global error handler; each layer handles errors at its abstraction level

**Examples:**
- `App.StartService(name)` returns error if service not found or start fails
- `sites.Manager.Create(domain, path)` returns error if cert signing or Nginx config generation fails
- `config.Store.Set()` returns error if JSON encoding or file write fails

## Cross-Cutting Concerns

**Logging:** 
- Framework: `pkg/logger/` (custom Info/Error/Action levels with millisecond timestamps)
- Disabled by default; enabled via config.DebugMode or App.EnableDebug()
- Writes to `~/.lamboserver/logs/debug.log`
- Service log readers: Logger.Reader type reads service-specific log files (Nginx error log, PHP-FPM error log, etc.)

**Validation:** 
- Version strings validated by service managers (e.g., PHP.ListInstalled() only returns valid versions)
- Domain names validated by sites manager before Nginx config generation
- Path existence checked before site creation
- No global validation layer; each domain validates its inputs

**Authentication:** 
- No user authentication (local desktop app)
- Privilege elevation handled via macOS privilege helper (mxcl.PrivilegedHelper) for root operations
- Keychain access for CA trust (via Cert manager)

**Concurrency:**
- Config store uses sync.RWMutex for thread-safe concurrent access
- Service managers use LaunchdManager/LaunchdService for safe daemon lifecycle coordination
- No explicit goroutines in service managers; launchctl and system commands are synchronous

