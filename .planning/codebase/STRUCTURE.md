# Codebase Structure

**Analysis Date:** 2026-04-16

## Directory Layout

```
lamboserver/
├── main.go                          # Wails entry point - creates App, runs Wails.Run()
├── app.go                           # Composition root, lifecycle hooks, generic service methods
├── doc.go                           # Package documentation for main
│
├── frontend/                        # React TypeScript frontend
│   ├── src/
│   │   ├── App.tsx                 # Main navigation and layout component
│   │   ├── main.tsx                # React entry point
│   │   ├── pages/                  # Page components (Dashboard, Services, PHP, Node, Sites, Logs, Databases)
│   │   ├── components/             # Reusable UI components (VersionList, etc.)
│   │   ├── assets/                 # Static assets (images, fonts)
│   │   └── style.css               # Global styles
│   ├── index.html                  # HTML entry point
│   ├── package.json                # Frontend dependencies (React, Vite, etc.)
│   ├── tsconfig.json               # TypeScript configuration
│   ├── vite.config.ts              # Vite build configuration
│   ├── wailsjs/                    # Auto-generated Wails type bindings for Go methods
│   └── dist/                       # Compiled frontend (output directory, git-ignored)
│
├── internal/                       # Private Go packages (not importable from other projects)
│   │
│   ├── services/                   # Unified service management and abstraction
│   │   ├── manager.go              # Service registry - maps service names to implementations
│   │   ├── service.go              # Interface definitions (Service, VersionedService, WebAdminService)
│   │   │
│   │   ├── nginx/                  # HTTP server management
│   │   │   ├── manager.go
│   │   │   ├── service_adapter.go  # Adapts manager to Service interface
│   │   │   ├── config.go           # Virtual host config generation
│   │   │   ├── interfaces.go       # Filesystem, Launchd, CmdRunner abstractions
│   │   │   ├── doc.go
│   │   │   ├── errors.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── php/                    # PHP runtime and FPM
│   │   │   ├── manager.go          # PHP version detection, installation
│   │   │   ├── fpm_manager.go      # PHP-FPM daemon lifecycle
│   │   │   ├── service_adapter.go  # Adapts to VersionedService
│   │   │   ├── doc.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── mysql/                  # MySQL database server
│   │   │   ├── manager.go          # MySQL installation, daemon control
│   │   │   ├── service_adapter.go
│   │   │   ├── doc.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── postgres/               # PostgreSQL database server
│   │   │   ├── manager.go
│   │   │   ├── service_adapter.go
│   │   │   ├── doc.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── nodejs/                 # Node.js runtime
│   │   │   ├── manager.go          # Node version switching
│   │   │   ├── service_adapter.go  # Adapts to VersionedService
│   │   │   ├── doc.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── dnsmasq/                # Local DNS resolver
│   │   │   ├── manager.go
│   │   │   ├── service_adapter.go
│   │   │   ├── doc.go
│   │   │   ├── *_test.go
│   │   │   └── testutil_test.go
│   │   │
│   │   ├── phpmyadmin/             # MySQL web admin UI
│   │   │   ├── manager.go
│   │   │   ├── service_adapter.go  # Adapts to WebAdminService
│   │   │   ├── doc.go
│   │   │   └── *_test.go
│   │   │
│   │   ├── pgadmin/                # PostgreSQL web admin UI
│   │   │   ├── manager.go
│   │   │   ├── service_adapter.go  # Adapts to WebAdminService
│   │   │   ├── doc.go
│   │   │   └── *_test.go
│   │   │
│   │   └── cloudbeaver/            # Multi-database web admin UI
│   │       ├── manager.go
│   │       ├── service_adapter.go
│   │       ├── doc.go
│   │       └── *_test.go
│   │
│   ├── sites/                      # Site management (domain -> path mapping)
│   │   ├── manager.go              # Create/delete sites, generate Nginx configs
│   │   ├── interfaces.go           # FileSystem, CertManager abstractions
│   │   ├── doc.go
│   │   ├── *_test.go
│   │   └── testutil_test.go
│   │
│   ├── cert/                       # SSL certificate management
│   │   ├── manager.go              # CA setup, per-site cert signing, Keychain trust
│   │   ├── interfaces.go           # FileSystem, AdminRunner abstractions
│   │   ├── doc.go
│   │   ├── *_test.go
│   │   └── testutil_test.go
│   │
│   ├── config/                     # Application configuration (persistent store)
│   │   ├── store.go                # Thread-safe config with sync.RWMutex, JSON persistence
│   │   ├── defaults.go             # Default config values
│   │   ├── doc.go
│   │   ├── *_test.go
│   │   └── testutil_test.go
│   │
│   ├── system/                     # macOS system abstractions and utilities
│   │   ├── paths.go                # Single source of truth: ~/.lamboserver/ directory structure
│   │   ├── launchd.go              # LaunchdManager - macOS launchctl abstraction
│   │   ├── darwin.go               # Darwin-specific operations (osascript, Keychain)
│   │   ├── realfs.go               # RealFS - concrete filesystem implementation
│   │   ├── realcmd.go              # RealCmdRunner - concrete command execution
│   │   ├── helper.go               # Privileged helper for root operations
│   │   ├── download.go             # Binary download utilities
│   │   ├── binary.go               # Binary detection and verification
│   │   ├── embed.go                # Embedded binary resources
│   │   ├── shell.go                # Shell integration (PATH injection to ~/.bashrc/.zshrc)
│   │   ├── embedded/               # Embedded binary archives (darwin-arm64, etc.)
│   │   ├── doc.go
│   │   ├── *_test.go
│   │   └── testutil_test.go
│   │
│   ├── cert/                       # (Listed above - SSL certificate CA management)
│   │
│   ├── binaries/                   # Binary distribution and installation
│   │   ├── (Source files for binary archive generation)
│   │
│   ├── process/                    # Process management utilities
│   │   └── (Process-level abstractions)
│   │
│   └── integration/                # Integration tests
│       ├── startup_test.go         # Test full startup sequence
│       ├── php_switch_test.go      # Test PHP version switching workflow
│       ├── site_creation_test.go   # Test site creation end-to-end
│       ├── service_lifecycle_test.go # Test service start/stop workflows
│       ├── helpers_test.go         # Shared test utilities
│       ├── doc.go
│       └── testutil_test.go
│
├── pkg/                            # Public Go packages (reusable by other projects)
│   │
│   ├── logger/                     # Application logging framework
│   │   ├── logger.go               # Logger struct (Info, Error, Action levels)
│   │   ├── reader.go               # Reader struct (read service log files)
│   │   ├── interfaces.go           # FileReader interface for testability
│   │   ├── doc.go
│   │   ├── *_test.go
│   │   └── testutil_test.go
│   │
│   └── notify/                     # Desktop notification utilities
│       └── notify.go               # Send macOS notifications
│
├── build/                          # Build configuration and assets
│   ├── darwin/                     # macOS build assets
│   │   └── (Build scripts, codesigning configs)
│   ├── windows/                    # Windows build assets
│   └── bin/                        # Build output directory
│
├── go.mod                          # Go module definition
├── go.sum                          # Go dependency checksums
├── wails.json                      # Wails framework configuration
├── main.go                         # (Root entry point - described above)
└── README.md                       # Project overview and setup instructions
```

## Directory Purposes

**`frontend/src/`:**
- Purpose: React/TypeScript UI for the desktop application
- Contains: Page components (Dashboard, Services, PHP, Node, Sites, Logs, Databases), reusable UI components, styles, assets
- Key files: `App.tsx` (main navigation), `main.tsx` (React entry), pages directory for domain-specific pages

**`internal/services/`:**
- Purpose: Unified service abstraction and manager implementations
- Contains: Service registry (Manager), interface definitions (Service, VersionedService, WebAdminService), per-service manager implementations with adapters
- Pattern: Each service has its own subdirectory with manager.go (business logic) and service_adapter.go (interface implementation)

**`internal/services/{nginx,php,mysql,postgres,nodejs,dnsmasq}/`:**
- Purpose: Individual service lifecycle management
- Contains: Manager type with Install/Start/Stop/Restart/Status/Logs methods, ServiceAdapter wrapping manager to implement unified Service interface
- Pattern: Managers depend on FileSystem, CmdRunner, LaunchdService interfaces for testability

**`internal/services/{phpmyadmin,pgadmin,cloudbeaver}/`:**
- Purpose: Web-based admin tools served through existing infrastructure
- Contains: Manager with Install/URL/IsInstalled methods, ServiceAdapter implementing WebAdminService interface
- Pattern: No daemon of their own; served via Nginx + backend runtime

**`internal/sites/`:**
- Purpose: Site lifecycle management (domain -> path mapping)
- Contains: Manager type for creating/deleting sites, generating Nginx virtual host configs, persisting to config store
- Key dependency: CertManager (for per-site SSL certificates)

**`internal/cert/`:**
- Purpose: Local CA and SSL certificate management
- Contains: Manager for CA setup, per-site certificate signing, Keychain trust (macOS-specific)
- Key interface: Depends on FileSystem and AdminRunner (for privilege elevation)

**`internal/config/`:**
- Purpose: Thread-safe persistent application configuration
- Contains: Store type with sync.RWMutex for concurrent access, JSON persistence to ~/.lamboserver/config.json
- Stores: ActivePhpVersion, ActiveNodeVersion, site mappings, MySQLEnabled, PostgreSQLEnabled, DebugMode

**`internal/system/`:**
- Purpose: macOS system abstractions and platform utilities
- Contains: 
  - Paths: Single source of truth for all ~/.lamboserver/ paths
  - LaunchdManager: macOS launchctl abstraction for daemon lifecycle
  - Darwin-specific: osascript, Keychain access
  - Abstractions: FileSystem, CmdRunner, AdminRunner, LaunchdService interfaces
  - Shell integration: PATH injection to ~/.bashrc/.zshrc
- Key file: `paths.go` - every path used in the app is defined here; update this file to change directory structure

**`internal/integration/`:**
- Purpose: End-to-end integration tests
- Contains: Multi-component workflow tests (startup sequence, PHP switching, site creation, service lifecycle)
- Pattern: Uses real implementations but against temporary directories; does not touch host system

**`pkg/logger/`:**
- Purpose: Reusable logging framework
- Contains: Logger (Info/Error/Action levels), Reader (read service log files)
- Public package: Can be imported by other projects

**`pkg/notify/`:**
- Purpose: Desktop notification utilities
- Contains: Helper to send macOS notifications
- Public package: Can be imported by other projects

**`build/`:**
- Purpose: Build configuration, code signing, platform-specific assets
- Contains: macOS (darwin/) and Windows (windows/) build scripts and configuration
- Output: Compiled binaries and app bundles

## Key File Locations

**Entry Points:**
- `main.go`: Wails entry point - creates App instance, starts Wails.Run()
- `app.go`: App struct definition, lifecycle hooks (startup, beforeClose, shutdown), public IPC methods
- `frontend/src/App.tsx`: React UI entry, navigation logic
- `frontend/src/main.tsx`: React DOM root

**Configuration:**
- `wails.json`: Wails framework config (window size, build commands, author)
- `go.mod`: Go module dependencies
- `frontend/package.json`: Node dependencies (React, Vite, TypeScript)
- `frontend/tsconfig.json`: TypeScript configuration

**Core Logic:**
- `internal/services/manager.go`: Service registry (maps service names to implementations)
- `internal/services/service.go`: Service interface definitions
- `internal/system/paths.go`: All ~.lamboserver/ paths defined here
- `internal/config/store.go`: Configuration persistence and access
- `internal/sites/manager.go`: Site creation/deletion logic

**Testing:**
- `internal/integration/`: End-to-end integration tests
- `internal/{services,sites,cert,config,system}/*_test.go`: Unit tests for each package
- `internal/{services,sites,cert,config,system}/testutil_test.go`: Shared test utilities

## Naming Conventions

**Files:**
- `manager.go`: Main type for a package (e.g., nginx.Manager, php.Manager)
- `service_adapter.go`: Adapter implementing a unified interface (e.g., ServiceAdapter implementing services.Service)
- `interfaces.go`: Abstract interfaces used within a package (e.g., FileSystem, CmdRunner)
- `doc.go`: Package documentation
- `*_test.go`: Unit tests
- `testutil_test.go`: Shared test utilities and fixtures

**Directories:**
- `internal/services/{service-name}/`: One subdirectory per service, lowercase with hyphens
- `internal/{sites,cert,config,system}/`: Domain features, lowercase, plural where appropriate
- `pkg/{package-name}/`: Public packages, lowercase

**Packages:**
- Match directory name: `package nginx` in `internal/services/nginx/`
- Use singular for domain concepts (sites package manages Site type), plural for collections (services.Manager manages multiple services)

**Types:**
- Manager: Primary type in a package (e.g., nginx.Manager, sites.Manager)
- Adapter: Wrapper type implementing interface (ServiceAdapter)
- Service: Interface types (services.Service, services.VersionedService)
- Store: Configuration storage (config.Store)

**Functions/Methods:**
- Constructor: `NewX` for main types (NewApp, NewManager, NewServiceAdapter)
- Getter: No prefix (e.g., Status(), Version(), Logs())
- Setter: Set prefix (e.g., Set() on config.Store)
- Query: Is prefix for booleans (IsInstalled(), IsRunning())
- Lifecycle: Verb names (Start, Stop, Restart, Install, Uninstall)

## Where to Add New Code

**New Service Implementation:**
1. Create `internal/services/{service-name}/` directory
2. Implement `manager.go` with service-specific lifecycle logic
3. Create `service_adapter.go` implementing appropriate interface (Service, VersionedService, or WebAdminService)
4. Create `interfaces.go` defining package-scoped abstractions if needed
5. Register in `app.go` NewApp() function: `mgr.Register("service-name", adapter)`

**New Page in Frontend:**
1. Create component file in `frontend/src/pages/{PageName}.tsx`
2. Import in `frontend/src/App.tsx`
3. Add routing case in `App.tsx` renderPage() switch statement
4. Add navigation item to navItems array if needed

**New Reusable Component:**
1. Create component file in `frontend/src/components/{ComponentName}.tsx`
2. Import and use in page components

**New Feature (e.g., new runtime version):**
1. Implement Manager in `internal/services/{service-name}/`
2. Register with services.Manager in App
3. Add public methods on App for frontend access (e.g., InstallNodeVersion, SwitchNodeVersion)
4. Create corresponding page or dashboard section in frontend

**New Configuration Option:**
1. Add field to config.AppConfig struct in `internal/config/store.go`
2. Update config.json schema/defaults in `internal/config/defaults.go`
3. Access via `a.Config.Get().FieldName` in App
4. Persist via `a.Config.Set(updatedConfig)` after modification

**New Utility Function:**
- Domain-specific: Add to appropriate package in `internal/` (not public)
- Reusable: Add to `pkg/` package (importable by other projects)

## Special Directories

**`frontend/dist/`:**
- Purpose: Compiled frontend build output
- Generated: Yes (built by Vite via `npm run build`)
- Committed: No (git-ignored)
- Embedded: Yes (embedded in binary via `//go:embed all:frontend/dist` in main.go)

**`frontend/wailsjs/`:**
- Purpose: Auto-generated TypeScript bindings for Go methods
- Generated: Yes (generated by Wails during build)
- Committed: No (git-ignored)
- Contents: Type definitions for App methods, models, event types

**`internal/system/embedded/`:**
- Purpose: Embedded binary archives for different architectures
- Generated: Yes (built during release process)
- Committed: Yes (needed for runtime)
- Structure: `embedded/{arch}/` (e.g., darwin-arm64/nginx.tar.gz)

**`build/`:**
- Purpose: Build configuration, code signing, platform-specific assets
- Contains: macOS (darwin/) and Windows (windows/) specific build scripts
- Used: By CI/CD and local build processes

**`~/.lamboserver/` (runtime):**
- Purpose: Application home directory (created at runtime)
- Structure:
  - `bin/`: Managed symlinks and helper binaries
  - `php/`: Installed PHP versions
  - `nodejs/`: Installed Node.js versions
  - `nginx/`: Nginx config and logs
  - `certs/`: CA and per-site certificates
  - `dnsmasq/`: DNSMasq config
  - `logs/`: Application and service logs
  - `services/`: Service-specific data
  - `config.json`: Application configuration (persistent)

