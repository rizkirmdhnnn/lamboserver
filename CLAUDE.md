<!-- GSD:project-start source:PROJECT.md -->
## Project

**LamboServer — System Tray Integration**

LamboServer is a macOS local development environment manager that controls Nginx, MySQL, PHP, PostgreSQL, Node.js, and related tools through a native desktop GUI. This milestone adds system tray integration so the app runs in the background with a tray icon, giving users quick access to service controls, status, and navigation without needing the full window open.

**Core Value:** Users can control their local dev services instantly from the system tray without opening the full application window.

### Constraints

- **Platform**: macOS only — system tray implementation is Darwin-specific
- **Framework**: Wails v2 — tray must integrate with Wails lifecycle without conflicts
- **Signing**: Ad-hoc only — no notarization, no entitlements beyond current set
- **Architecture**: Must maintain existing service adapter pattern and App struct as central coordinator
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.25.0 - Backend logic, service management, system integration (`go.mod`)
- TypeScript ^4.6.4 - Frontend UI (`frontend/package.json`)
- CSS - Styling (`frontend/src/style.css`)
- Bash - Build and release scripts (`scripts/build-dmg.sh`)
## Runtime
- macOS only (darwin) - Application uses macOS-specific APIs (launchd, Keychain, /etc/resolver)
- Wails v2 runtime - Go + embedded WebView (WKWebView on macOS)
- Architecture support: darwin/universal (arm64 + amd64) via `wails build -platform darwin/universal`
- Go modules (`go.mod`, `go.sum` present)
- npm for frontend (`frontend/package.json`)
- Lockfile: `go.sum` present; `frontend/package-lock.json` should exist after install
## Frameworks
- Wails v2 (v2.12.0) - Desktop app framework bridging Go backend to WebView frontend (`go.mod`)
- React 18 (^18.2.0) - Frontend UI framework (`frontend/package.json`)
- testify v1.11.1 - Go test assertions and mocking (`go.mod`)
- No frontend testing framework detected
- Vite ^3.0.7 - Frontend bundler and dev server (`frontend/package.json`)
- @vitejs/plugin-react ^2.0.1 - React HMR support (`frontend/vite.config.ts`)
- Wails CLI - Build tool (`wails build`, `wails dev`)
## Key Dependencies
- `github.com/wailsapp/wails/v2` v2.12.0 - Core desktop framework, IPC between Go and React
- `github.com/go-sql-driver/mysql` v1.9.1 - MySQL database driver for database CRUD operations (`internal/services/mysql/manager.go`)
- `github.com/jackc/pgx/v5` v5.9.1 - PostgreSQL database driver via `pgx/stdlib` (`internal/services/postgres/manager.go`)
- `github.com/pkg/browser` v0.0.0-20240102092130 - Opens URLs in default browser for web admin tools (`app.go`)
- `react-router-dom` ^7.14.1 - Client-side routing (declared but App.tsx uses manual state-based navigation)
- `lucide-react` ^1.8.0 - Icon library for sidebar navigation (`frontend/src/App.tsx`)
- `github.com/labstack/echo/v4` v4.13.3 - HTTP server used internally by Wails
- `github.com/gorilla/websocket` v1.5.3 - WebSocket for Wails dev mode communication
- `github.com/google/uuid` v1.6.0 - UUID generation
- `git.sr.ht/~jackmordaunt/go-toast/v2` v2.0.3 - macOS toast notifications
- `golang.org/x/crypto` v0.33.0 - Cryptographic primitives (SSL cert generation)
## Configuration
- `~/.lamboserver/config.json` - JSON-based config persisted by `internal/config/store.go`
- Thread-safe in-memory cache with `sync.RWMutex`, atomic file writes
- Config schema defined in `internal/config/store.go` as `AppConfig` struct
- `wails.json` - Wails project configuration (app name, build commands, author info)
- `frontend/tsconfig.json` - Strict mode enabled, ESNext target, react-jsx transform
- `frontend/vite.config.ts` - Minimal Vite config with React plugin
- `build/darwin/entitlements.plist` - macOS entitlements (JIT, unsigned memory, network, Apple Events)
- `build/darwin/Info.plist` / `Info.dev.plist` - macOS app bundle metadata
## Build Pipeline
- GitHub Actions workflow: `.github/workflows/release.yml`
- Trigger: Push tag matching `v*`
- Runner: `macos-latest`
- Steps: Setup Go 1.25 + Node 20 -> Install Wails -> Build universal .app -> Ad-hoc codesign -> Create DMG -> GitHub Release via `softprops/action-gh-release@v2`
- No Apple Developer signing (ad-hoc only, per project constraint)
## Embedded Assets
- `frontend/dist/` embedded via Go `//go:embed all:frontend/dist` directive in `main.go`
- Served by Wails asset server at runtime
- `internal/system/embedded/darwin-arm64/nginx` - Pre-compiled Nginx binary
- `internal/system/embedded/darwin-arm64/dnsmasq` - Pre-compiled dnsmasq binary
## Platform Requirements
- macOS (required for Wails WebView, launchd, and system integration)
- Go 1.25+
- Node.js 20+ (for frontend build)
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- macOS only (uses launchd, Keychain, /etc/resolver, WKWebView)
- No Apple Developer Account required (ad-hoc signing)
- Distributed as `.dmg` file
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- `manager.go` — primary type and methods for each service package
- `interfaces.go` — consumer-site interfaces (FileSystem, CommandRunner, LaunchdService, AdminRunner)
- `errors.go` — package-level sentinel errors
- `doc.go` — package-level documentation comment
- `service_adapter.go` — adapter implementing `services.Service` for the unified service registry
- `export_test.go` — exports internal constructors for external test packages
- `testutil_test.go` — mock definitions and test helper factories
- `*_test.go` — test files use `_test` suffix (standard Go convention)
- Pages: `PascalCase.tsx` in `frontend/src/pages/` (e.g., `Dashboard.tsx`, `SitesPage.tsx`, `PhpPage.tsx`)
- Components: `PascalCase.tsx` in `frontend/src/components/` (e.g., `Toast.tsx`, `VersionList.tsx`)
- Entry: `main.tsx`, `App.tsx`
- Exported: `PascalCase` — `NewManager`, `IsInstalled`, `EnsureConfig`, `SetupCA`
- Unexported helpers: `camelCase` — `ensureLogFile`, `readLastNLines`, `ensureServicesRunning`
- Constructors: `New<Type>(deps...)` pattern — `NewManager(paths, store, fs, cmd)`
- Test constructors: `newTest<Type>(t *testing.T)` — `newTestManager(t)`, `newTestPaths(t)`
- Exported constants: `PascalCase` — `ServiceLabel`, `AppName`, `TestTLD`
- Unexported fields: `camelCase` — `paths`, `launchd`, `fs`, `admin`, `binary`
- Public structs: `PascalCase` — `Manager`, `ServiceStatus`, `PhpVersion`, `AppConfig`
- Interfaces: `PascalCase` verb-noun — `FileSystem`, `CommandRunner`, `LaunchdService`, `AdminRunner`, `HelperRunner`
- Mock types (in tests): `Mock<InterfaceName>` — `MockFileSystem`, `MockLaunchdService`, `MockAdminRunner`
- Interfaces: `PascalCase` — `DashboardData`
- Functions: `camelCase` — `loadStatus`, `renderPage`
- Components: `PascalCase` function components — `function Dashboard()`, `function App()`
## Code Style
- `gofmt` standard formatting (no custom config)
- Tab indentation (Go default)
- No `.golangci.yml` or linter config detected — relies on `gofmt` only
- No ESLint, Prettier, or Biome config present
- TypeScript strict mode enabled in `frontend/tsconfig.json` (`"strict": true`)
- Vite build with `tsc` type checking: `"build": "tsc && vite build"`
## Import Organization
- Use aliases only to resolve conflicts: `wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"` in `app.go`
- No path aliases configured — uses relative paths (`../../wailsjs/go/main/App`)
## Error Handling
- `try/catch` with `console.error` for Wails IPC calls:
## Logging
- `Info(format, args...)` — informational messages
- `Error(format, args...)` — error messages
- `Action(name, err)` — structured action result logging (logs OK or FAILED)
- Log all service lifecycle events: start, stop, restart
- Log startup sequence steps: helper install, CA setup, shell integration
- Use `Debug.Action(name, err)` for every user-initiated action in `app.go`
- Use `Debug.Info/Error` for internal operations and diagnostics
## Comments
- Every package has a `doc.go` file with a package-level doc comment explaining the package's purpose and scope
- Pattern: `// Package <name> <verb phrase describing purpose>.`
- Example from `internal/services/dnsmasq/doc.go`:
- All exported types and methods have doc comments
- Method comments start with the method name: `// Start installs nginx as a root LaunchDaemon...`
- Struct comments describe purpose and usage: `// Manager handles dnsmasq binary discovery...`
- Unicode box-drawing characters for visual section separation:
- Reference tracking codes for cross-cutting decisions: `// D-08: pgweb must stop when PostgreSQL stops`
- Explain "why" not "what": `// Ensure log files exist with user ownership`
- Every service adapter includes a compile-time interface assertion:
- Real implementations also include compile-time checks:
## Function Design
- All managers use constructor injection with explicit dependencies:
- Dependencies are interfaces (except `*system.Paths` and `*config.Store` which are concrete)
- Service manager methods are typically 10-30 lines
- `app.go` methods are thin delegation: validate input, call manager, log result, return
- Single error return for mutation operations: `Start() error`, `Stop() error`
- Value + error for queries: `ListInstalled() ([]PhpVersion, error)`
- Struct return for status: `Status() ServiceStatus`
## Module Design
- One manager type per package: `dnsmasq.Manager`, `nginx.Manager`, `php.Manager`
- Interfaces defined at consumer site (not in a shared contracts package) per ADR-001
- `internal/` for application-specific code, `pkg/` for reusable utilities
- Each service has a `ServiceAdapter` struct wrapping `*Manager` to implement `services.Service`
- Adapters live in the same package as the manager: `dnsmasq.ServiceAdapter`, `nginx.ServiceAdapter`
- Created via `NewServiceAdapter(m *Manager)` factory function
- No barrel files — each package is imported directly
- `export_test.go` pattern exposes internal constructors for external test packages:
## Architecture Decision Records
- ADRs live in `docs/adr/` directory
- Follow format: `ADR-NNN-<slug>.md` (e.g., `ADR-001-interface-adoption.md`)
- Reference codes (D-01, D-02, D-08) used in source comments to trace back to decisions
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Composition root pattern: all dependency wiring happens in `NewApp()` in `app.go`
- Go backend exposes public methods to React frontend via Wails IPC bindings (no REST API)
- Interface-based dependency injection with consumer-defined interfaces (Go convention D-02)
- Service adapter pattern: each service has a domain-specific Manager + a ServiceAdapter that implements unified `services.Service` or `services.VersionedService`
- macOS launchd for daemon lifecycle (nginx, dnsmasq, mysql run as system LaunchDaemons via a privileged helper)
- All application state stored under `~/.lamboserver/` (config, binaries, certs, logs)
## Layers
- Purpose: React SPA providing the desktop GUI
- Location: `frontend/src/`
- Contains: Pages, components, CSS, Wails-generated TypeScript bindings
- Depends on: Wails IPC bindings (`frontend/wailsjs/go/main/App.js`)
- Used by: End user via native macOS window
- Purpose: Thin glue between frontend and backend; validates input, delegates to managers, returns results
- Location: `app.go`
- Contains: All public methods on the `App` struct (auto-exposed to frontend by Wails)
- Depends on: All internal service managers
- Used by: Frontend via Wails-generated TypeScript functions
- Purpose: Domain-specific business logic for each managed service
- Location: `internal/services/*/manager.go`
- Contains: Installation, configuration generation, start/stop/restart, version management
- Depends on: `internal/system` (paths, launchd, fs, cmd), `internal/config`, `internal/cert`
- Used by: `App` struct methods in `app.go`
- Purpose: Generic service dispatch by name (e.g., `StartService("nginx")`)
- Location: `internal/services/service.go`
- Contains: `Service`, `VersionedService`, `WebAdminService` interfaces; `Manager` registry
- Depends on: Service adapters
- Used by: `App.StartService()`, `App.StopService()`, etc.
- Purpose: OS-specific operations (launchd, shell commands, privilege elevation, filesystem)
- Location: `internal/system/`
- Contains: `Paths`, `LaunchdManager`, `Helper`, `RealFS`, `RealCmdRunner`, `RealAdminRunner`, `Integration`
- Depends on: macOS system APIs (launchctl, osascript, pgrep)
- Used by: All service managers
- Purpose: Persistent, thread-safe application config
- Location: `internal/config/store.go`
- Contains: `AppConfig` struct, `Store` with `sync.RWMutex`
- Depends on: Filesystem (JSON file at `~/.lamboserver/config.json`)
- Used by: All service managers, `App` struct
- Purpose: Cross-cutting utilities
- Location: `pkg/logger/`, `pkg/notify/`, `internal/binaries/`, `internal/cert/`, `internal/process/`
- Contains: Debug logging, notifications, binary download registry, SSL cert generation, process abstractions
## Data Flow
- Backend state: `config.Store` (in-memory cache + JSON file, protected by `sync.RWMutex`)
- Frontend state: React `useState` hooks per page (no global state manager like Redux)
- Service state: Derived at query time from launchd/process inspection (not cached)
## Key Abstractions
- Purpose: Unified API for heterogeneous services
- Examples: `internal/services/service.go`
- Pattern: Three-tier interface hierarchy:
- Purpose: Bridges domain-specific Managers to unified Service interfaces
- Examples: `internal/services/php/service_adapter.go`, `internal/services/nodejs/service_adapter.go`, `internal/services/phpmyadmin/service_adapter.go`
- Pattern: Adapter composes one or more Managers and implements the Service/VersionedService interface. Compile-time check via `var _ services.VersionedService = (*ServiceAdapter)(nil)`
- Purpose: Each package defines only the interface methods it needs, not a god interface
- Examples: `internal/sites/interfaces.go` (FileSystem with 3 methods), `internal/cert/interfaces.go` (FileSystem with 5 methods), `internal/services/nginx/` (separate LaunchdService, FileSystem, HelperRunner)
- Pattern: Real implementations (`RealFS`, `RealCmdRunner`, `RealAdminRunner`) in `internal/system/` satisfy all consumer interfaces via structural typing
- Purpose: Centralizes all file/directory path resolution
- Examples: `internal/system/paths.go`
- Pattern: All paths derive from `~/.lamboserver/`. Any path change only requires editing `paths.go`.
- Purpose: Avoid repeated password prompts for root operations
- Examples: `internal/system/helper.go`
- Pattern: One-time sudoers setup, then `sudo -n lambo-helper <command>` for daemon management
## Entry Points
- Location: `main.go`
- Triggers: macOS app launch
- Responsibilities: Embeds frontend assets, creates App via `NewApp()`, configures Wails window options, calls `wails.Run()`
- Location: `app.go` lines 112-183
- `startup()`: Ensures directories, installs helper, sets up CA, restores symlinks, starts services
- `beforeClose()`: Shows confirmation dialog
- `shutdown()`: Stops all services in dependency order, kills stale processes
- Location: `frontend/src/main.tsx`
- Triggers: Wails WebView load
- Responsibilities: Renders React app root
## Error Handling
- Go errors returned directly from all public `App` methods; Wails converts to Promise rejections
- Startup errors logged via `a.Debug.Error()` but do not prevent app launch (graceful degradation)
- `a.Debug.Action(name, err)` pattern used consistently to log operation outcomes
- Frontend uses try/catch on async calls with `console.error` logging
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
