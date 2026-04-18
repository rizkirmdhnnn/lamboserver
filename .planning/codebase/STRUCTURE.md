# Codebase Structure

**Analysis Date:** 2026-04-17

## Directory Layout

```
lamboserver/
├── main.go                  # Wails app entry point, embeds frontend assets
├── app.go                   # Composition root + all frontend-bound methods (~800 lines)
├── doc.go                   # Package-level documentation
├── go.mod                   # Go module definition (Go 1.25, Wails v2.12)
├── go.sum                   # Go dependency checksums
├── wails.json               # Wails project configuration
├── build/                   # Wails build assets (app icons, Info.plist)
├── docs/                    # Project documentation
│   └── adr/                 # Architecture Decision Records
├── scripts/                 # Build and release scripts
├── frontend/                # React + TypeScript frontend (Vite)
│   ├── src/                 # Application source
│   │   ├── main.tsx         # React entry point
│   │   ├── App.tsx          # Root component with sidebar navigation
│   │   ├── style.css        # Global styles
│   │   ├── pages/           # Page-level components (one per nav item)
│   │   ├── components/      # Shared reusable components
│   │   └── assets/          # Static assets (fonts, images)
│   ├── wailsjs/             # Auto-generated Wails TypeScript bindings
│   │   ├── go/main/App.js   # JS functions calling Go App methods
│   │   ├── go/main/App.d.ts # TypeScript declarations for Go methods
│   │   ├── go/models.ts     # TypeScript types generated from Go structs
│   │   └── runtime/         # Wails runtime JS/TS bindings
│   ├── package.json         # Frontend dependencies (React, Vite, lucide-react)
│   ├── tsconfig.json        # TypeScript configuration
│   ├── vite.config.ts       # Vite build configuration
│   └── index.html           # HTML entry point
├── internal/                # Private Go packages (business logic)
│   ├── config/              # Persistent app configuration (JSON-backed store)
│   ├── services/            # Service managers and unified interfaces
│   │   ├── service.go       # Service/VersionedService/WebAdminService interfaces + Manager registry
│   │   ├── nginx/           # Nginx reverse proxy management
│   │   ├── dnsmasq/         # DNS resolver management (.test TLD)
│   │   ├── php/             # PHP version + FPM process management
│   │   ├── nodejs/          # Node.js version management
│   │   ├── mysql/           # MySQL server management
│   │   ├── postgres/        # PostgreSQL server management
│   │   ├── phpmyadmin/      # phpMyAdmin web admin management
│   │   └── pgweb/           # pgweb PostgreSQL web admin
│   ├── sites/               # Virtual host / site management
│   ├── cert/                # Local CA and SSL certificate generation
│   ├── system/              # OS-level abstractions (paths, launchd, shell, fs)
│   │   └── embedded/        # Embedded binaries (nginx, dnsmasq for darwin-arm64)
│   ├── binaries/            # Binary download registry and downloader
│   ├── process/             # Process management abstractions
│   ├── integration/         # Integration tests
│   └── tray/                # System tray support (placeholder)
├── pkg/                     # Public Go packages (shared utilities)
│   ├── logger/              # Debug logging and log file reader
│   └── notify/              # User notification support
└── .github/
    └── workflows/           # GitHub Actions CI/CD
```

## Directory Purposes

**`internal/services/`:**
- Purpose: All managed service logic. Each subdirectory is one service.
- Contains: `manager.go` (business logic), `interfaces.go` (consumer-defined deps), `service_adapter.go` (unified interface adapter), `doc.go`, tests
- Key files:
  - `internal/services/service.go`: Defines `Service`, `VersionedService`, `WebAdminService` interfaces and the `Manager` registry
  - `internal/services/php/manager.go`: PHP version detection, installation, switching
  - `internal/services/php/fpm.go`: PHP-FPM process lifecycle
  - `internal/services/php/service_adapter.go`: Composes Manager+FpmManager into VersionedService
  - `internal/services/nginx/manager.go`: Nginx config generation, launchd daemon lifecycle
  - `internal/services/mysql/manager.go`: MySQL installation, data dir init, database CRUD
  - `internal/services/postgres/manager.go`: PostgreSQL installation, data dir init, database CRUD
  - `internal/services/dnsmasq/manager.go`: DNS resolver for .test TLD
  - `internal/services/nodejs/manager.go`: Node.js version management
  - `internal/services/phpmyadmin/manager.go`: phpMyAdmin web admin setup
  - `internal/services/pgweb/manager.go`: pgweb web admin for PostgreSQL

**`internal/system/`:**
- Purpose: OS-level abstractions and real implementations
- Contains: Path resolution, launchd management, shell integration, filesystem/command adapters
- Key files:
  - `internal/system/paths.go`: Single source of truth for all file locations (~236 lines)
  - `internal/system/launchd.go`: macOS launchd plist generation and service control
  - `internal/system/helper.go`: Privileged helper script (sudoers-based, passwordless after setup)
  - `internal/system/realfs.go`: Real filesystem implementation satisfying all consumer FileSystem interfaces
  - `internal/system/realcmd.go`: Real command runner + admin runner implementations
  - `internal/system/darwin.go`: macOS-specific utilities (osascript admin, RunCommand)
  - `internal/system/shell.go`: Shell PATH integration (~/.zshrc, ~/.bashrc injection)
  - `internal/system/binary.go`: Binary locator (checks local path then system PATH)
  - `internal/system/embed.go`: Go embed directives for bundled binaries
  - `internal/system/embedded/darwin-arm64/`: Pre-compiled nginx and dnsmasq binaries

**`internal/config/`:**
- Purpose: Thread-safe, JSON-backed application configuration
- Key files:
  - `internal/config/store.go`: `Store` struct with `sync.RWMutex`, in-memory cache, atomic file writes
  - `internal/config/defaults.go`: Default configuration values

**`internal/sites/`:**
- Purpose: Nginx virtual host management for local development sites
- Key files:
  - `internal/sites/manager.go`: Site linking/unlinking, Nginx config generation, SSL cert provisioning
  - `internal/sites/interfaces.go`: Consumer-defined FileSystem interface

**`internal/cert/`:**
- Purpose: Local Certificate Authority and per-site SSL certificate generation
- Key files:
  - `internal/cert/ca.go`: CA key/cert generation, system keychain trust, site cert issuance
  - `internal/cert/interfaces.go`: FileSystem and AdminRunner interfaces
  - `internal/cert/errors.go`: Domain-specific error types

**`internal/binaries/`:**
- Purpose: Centralized download registry for service binaries
- Key files:
  - `internal/binaries/registry.go`: Maps service+version to download URL (php, nodejs, mysql, postgresql)
  - `internal/binaries/downloader.go`: HTTP download with optional checksum verification

**`internal/process/`:**
- Purpose: Platform-neutral process management abstraction
- Key files:
  - `internal/process/runner.go`: `ProcessManager` interface and `ProcessConfig` struct
  - `internal/process/launchd.go`: macOS launchd implementation
  - `internal/process/pidfile.go`: PID-file-based process management (for PostgreSQL)

**`pkg/logger/`:**
- Purpose: Application logging and log file reading
- Key files:
  - `pkg/logger/logger.go`: Toggle-able file-based debug logger with timestamps
  - `pkg/logger/reader.go`: Reads log files from app log directories (used by Logs page)
  - `pkg/logger/interfaces.go`: Logger interface definition

**`frontend/src/pages/`:**
- Purpose: One component per navigation page
- Key files:
  - `frontend/src/pages/Dashboard.tsx`: Service status overview, debug toggle
  - `frontend/src/pages/ServicesPage.tsx`: Start/stop/restart services
  - `frontend/src/pages/PhpPage.tsx`: PHP version management
  - `frontend/src/pages/NodePage.tsx`: Node.js version management
  - `frontend/src/pages/SitesPage.tsx`: Virtual host management
  - `frontend/src/pages/DatabasePage.tsx`: MySQL database management
  - `frontend/src/pages/PgDatabasePage.tsx`: PostgreSQL database management
  - `frontend/src/pages/LogsPage.tsx`: Log file viewer

**`frontend/src/components/`:**
- Purpose: Shared UI components
- Key files:
  - `frontend/src/components/VersionList.tsx`: Reusable version list with install/switch/uninstall
  - `frontend/src/components/Toast.tsx`: Toast notification component

## Key File Locations

**Entry Points:**
- `main.go`: Application bootstrap, Wails configuration, frontend asset embedding
- `app.go`: Composition root (`NewApp()`) and all IPC-bound methods
- `frontend/src/main.tsx`: React DOM render
- `frontend/src/App.tsx`: Root React component with navigation routing

**Configuration:**
- `wails.json`: Wails project config (app name, build commands, version)
- `go.mod`: Go module and dependencies
- `frontend/package.json`: Frontend dependencies
- `frontend/tsconfig.json`: TypeScript config
- `frontend/vite.config.ts`: Vite build config

**Core Logic:**
- `internal/services/service.go`: Service interface definitions and registry
- `internal/system/paths.go`: All filesystem path derivation
- `internal/system/launchd.go`: macOS daemon management
- `internal/system/helper.go`: Privileged helper for passwordless root ops
- `internal/config/store.go`: Application config persistence

**Testing:**
- `internal/config/store_test.go`: Config store unit tests
- `internal/services/php/manager_test.go`: PHP manager tests
- `internal/services/php/fpm_test.go`: PHP-FPM tests
- `internal/services/php/detector_test.go`: PHP version detector tests
- `internal/services/nginx/manager_test.go` (if exists): Nginx tests
- `internal/services/nodejs/manager_test.go`: Node.js manager tests
- `internal/integration/`: Integration test suite (startup, site creation, PHP switching, service lifecycle)

## Naming Conventions

**Files:**
- `manager.go`: Primary business logic for a service package
- `interfaces.go`: Consumer-defined dependency interfaces (per D-02)
- `service_adapter.go`: Adapter implementing unified `services.Service` or `services.VersionedService`
- `doc.go`: Package-level documentation
- `*_test.go`: Test files co-located with source
- `testutil_test.go`: Test helpers and mocks (test-only)
- `export_test.go`: Exported test helpers for white-box testing from other packages

**Directories:**
- `internal/services/<service-name>/`: One directory per managed service (lowercase, singular)
- `internal/<domain>/`: Domain packages (config, cert, sites, system, process, binaries)
- `pkg/<name>/`: Public shared utility packages
- `frontend/src/pages/<PageName>.tsx`: PascalCase page components
- `frontend/src/components/<ComponentName>.tsx`: PascalCase shared components

**Go Packages:**
- Package names match directory names (lowercase, no underscores)
- Struct names are PascalCase: `Manager`, `Store`, `ServiceAdapter`, `LaunchdManager`
- Interface methods are PascalCase per Go convention
- Constants use PascalCase: `ServiceLabel`, `StatusRunning`, `AppDirName`

## Where to Add New Code

**New Service (e.g., Redis):**
1. Create `internal/services/redis/`
2. Add `manager.go` with domain logic (install, start, stop, status)
3. Add `interfaces.go` with consumer-defined FileSystem/CommandRunner interfaces
4. Add `service_adapter.go` implementing `services.Service`
5. Add `doc.go` with package documentation
6. Register adapter in `NewApp()` in `app.go`: `mgr.Register("redis", redis.NewServiceAdapter(redisMgr))`
7. Add Manager field to `App` struct in `app.go`
8. Add any service-specific App methods to `app.go`
9. Add paths to `internal/system/paths.go`
10. Add download source to `internal/binaries/registry.go`
11. Add frontend page in `frontend/src/pages/RedisPage.tsx`
12. Add nav item in `frontend/src/App.tsx`

**New Frontend Page:**
1. Create `frontend/src/pages/NewPage.tsx`
2. Add to `navItems` array in `frontend/src/App.tsx`
3. Add case to `renderPage()` switch in `frontend/src/App.tsx`
4. Call Go backend via auto-generated functions from `frontend/wailsjs/go/main/App.js`

**New Shared Component:**
- Place in `frontend/src/components/ComponentName.tsx`

**New System Utility:**
- Place in `internal/system/` if macOS-specific
- Place in `pkg/` if reusable across projects

**New App Method (exposed to frontend):**
- Add public method to `App` struct in `app.go`
- Run `wails generate module` to regenerate TypeScript bindings in `frontend/wailsjs/`

## Special Directories

**`frontend/wailsjs/`:**
- Purpose: Auto-generated TypeScript bindings for Go backend methods
- Generated: Yes (by Wails build toolchain)
- Committed: Yes (checked into git for IDE support)

**`internal/system/embedded/`:**
- Purpose: Pre-compiled binaries bundled into the Go binary via `//go:embed`
- Generated: No (manually placed)
- Committed: Yes (darwin-arm64 nginx and dnsmasq binaries)

**`build/`:**
- Purpose: Wails build configuration (app icons, macOS Info.plist template)
- Generated: Partially (Wails scaffolding)
- Committed: Yes

**`frontend/dist/`:**
- Purpose: Vite build output embedded into Go binary
- Generated: Yes (during `wails build`)
- Committed: No (in .gitignore)

---

*Structure analysis: 2026-04-17*
