<!-- GSD:project-start source:PROJECT.md -->
## Project

**LamboServer Refactor**

A refactoring initiative for LamboServer — a macOS Wails-based desktop app that manages local development infrastructure (PHP, Node.js, Nginx, DNS, SSL). The goal is to restructure the codebase for better maintainability, reliability, and testability without changing user-facing behavior.

**Core Value:** The codebase should be easy to navigate, modify, and extend — any developer can find, understand, and change any feature without fear of breaking something else.

### Constraints

- **Behavior preservation**: All existing user-facing behavior must remain identical
- **Wails binding**: Public methods on `App` struct are auto-exposed to frontend — method signatures cannot change
- **macOS only**: Platform-specific code (launchd, keychain, osascript) stays Darwin-only
- **No frontend changes**: Refactor is backend-only; frontend code untouched
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.23.0 - Backend application, service managers, system integration
- TypeScript 4.6.4 - Frontend type checking
- JavaScript/TSX - React UI layer
- Shell - Platform-specific shell integration (`.bashrc`, `.zshrc`)
## Runtime
- Go 1.23.0 (server-side)
- Node.js (managed versions via internal installer, not required as host runtime)
- Wails v2 runtime for desktop UI
- Go modules - Lockfile: `go.mod`, `go.sum` present
- npm - Lockfile: `package-lock.json` present
## Frameworks
- Wails v2.12.0 - Desktop application framework (binds Go backend to React frontend)
- React 18.2.0 - UI framework
- React Router DOM 7.14.1 - Frontend routing
- Vite 3.0.7 - Frontend bundler and dev server
- TypeScript 4.6.4 - Type checking
- @vitejs/plugin-react 2.0.1 - React plugin for Vite
- Stretchr Testify 1.10.0 - Assertion library for Go tests
## Key Dependencies
- github.com/wailsapp/wails/v2 v2.12.0 - Desktop application framework, enables Go-React binding and native OS integration
- github.com/labstack/echo/v4 v4.13.3 - HTTP framework (included by Wails)
- github.com/gorilla/websocket v1.5.3 - WebSocket support for frontend-backend communication
- github.com/wailsapp/go-webview2 v1.0.22 - WebView2 integration for Windows
- github.com/go-ole/go-ole v1.3.0 - Windows OLE support
- github.com/godbus/dbus/v5 v5.1.0 - Linux D-Bus support
- git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 - Desktop notifications
- github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c - Browser opening utility
- github.com/google/uuid v1.6.0 - UUID generation
- github.com/samber/lo v1.49.1 - Functional utilities
- golang.org/x/crypto v0.33.0 - Cryptographic functions for CA generation
- golang.org/x/net v0.35.0 - Network utilities
- golang.org/x/sys v0.30.0 - System calls
- lucide-react 1.8.0 - Icon library
## Configuration
- Configuration stored in `~/.lamboserver/config.json` (JSON-based AppConfig)
- No environment variables required for base operation
- Platform-specific paths managed via `internal/platform/paths.go`
- Wails config: `wails.json` (app name, output filename, frontend build commands)
- Frontend TypeScript config: `frontend/tsconfig.json` (ESNext target, React JSX)
- Frontend build config: `frontend/vite.config.ts` (React plugin)
## Platform Requirements
- Go 1.23.0+
- Node.js (for frontend development only, not required at runtime)
- npm/yarn
- macOS (primary platform, code references Darwin-specific features)
- Xcode command line tools (for compilation)
- macOS (application targets macOS via `mac.Options` in Wails config)
- Supports x64 and arm64 architectures
- System services: launchd, dnsmasq, nginx (bundled or from Homebrew)
- PHP versions (downloaded from herdphp.com, not bundled)
- Node.js versions (downloaded from nodejs.org, not bundled)
## Build & Runtime
- Wails generates native macOS application
- Frontend assets embedded via Go's `//go:embed` directive in `main.go`
- Final binary: `LamboServer` (configurable output filename)
- Wails embeds WebView runtime (WebKit on macOS)
- Frontend served from embedded assets, no external CDN
- Bound methods: Go App struct methods exposed to JavaScript via Wails binding
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Go files use `snake_case` with `_test.go` suffix for test files (currently no test files present)
- Package names match directory names in lowercase: `package nginx`, `package php`, `package config`
- No test files found in codebase - testing gap identified
- Receiver functions (methods) use PascalCase: `NewManager()`, `ListInstalled()`, `SetActive()`
- Exported functions (public) start with capital letter: `NewApp()`, `Enable()`, `Status()`
- Unexported functions (private) start with lowercase: `addVersion()`, `detectHerd()`, `findExistingBin()`
- Action-oriented verbs: `Get`, `Set`, `List`, `Start`, `Stop`, `Link`, `Unlink`, `Install`, `Uninstall`
- Local variables use camelCase: `err`, `version`, `activeVersion`, `seen`, `filePath`
- Constants use UPPER_CASE: `AppName`, `AppDirName`, `TestTLD`, `ServiceLabel`, `nginxVersion`
- Struct fields use PascalCase: `Paths`, `Config`, `Php`, `Nginx`, `Active`, `Version`, `Binary`
- Private struct fields use camelCase: `mu` (mutex), `enabled`, `file`, `logger`, `logPath`
- Struct names use PascalCase: `Manager`, `AppConfig`, `PhpVersion`, `ServiceStatus`, `Site`
- Interface names follow Go convention: `Reader`, `Logger`
- Struct fields for JSON serialization use PascalCase with struct tags: 
## Code Style
- Standard Go formatting (gofmt)
- 2-space indentation in plist templates
- Line length: Generally follows Go conventions (no strict limit enforced)
- No `.eslintrc` or `.prettierrc` configuration files detected
- No explicit linting configuration - implies reliance on `go fmt` and `go vet`
- None detected - all imports use full module paths
- Module root: `github.com/rizkirmdhnnn/lamboserver`
- Internal packages accessed as: `github.com/rizkirmdhnnn/lamboserver/internal/...`
## Error Handling
- Errors are returned explicitly from functions, never panicked
- Error wrapping uses `fmt.Errorf(..., %w)` for context preservation:
- Early returns on error are standard:
- Errors are checked individually at each step, no error grouping
- Some operations ignore errors silently (e.g., `os.UserHomeDir()` in anonymous functions)
- File operations use specific error checking: `if os.IsNotExist(err)`
## Logging
- Standard library `log` package with custom wrapper
- Custom `Logger` struct in `internal/debug/logger.go` wraps Go's `log.Logger`
- Debug logger is optional and toggled via `EnableDebug()` / `DisableDebug()`
- All public API methods log info messages: `a.Debug.Info("GetPhpVersions called")`
- Action results logged with `Action()` method: `a.Debug.Action("SetActivePhp(...)", err)`
- Format strings used for info logging: `a.Debug.Info("Found %d PHP versions", len(versions))`
- Error logging includes full error context: `a.Error("ACTION %s -> FAILED: %v", action, err)`
- Logger uses timestamps: `timestamp := time.Now().Format("2006-01-02 15:04:05.000")`
- Log levels: INFO, ERROR, and special ACTION level for tracking user actions
- Log file location: `~/.lamboserver/logs/debug.log`
## Comments
- Package-level documentation: Each package has a doc comment:
- Function documentation: Not consistently applied to private functions
- Exported types documented: `// App struct holds all service managers`
- Complex logic receives inline comments:
- Not applicable - this is a Go codebase
- No equivalent documentation generation (no godoc comments in most files)
## Function Design
- Functions generally kept compact (10-50 lines)
- Handler methods in `app.go` follow consistent short pattern (5-10 lines):
- Receiver pattern: Methods receive value `a *App`, `m *Manager`, `l *Logger`
- Dependency injection via constructor: `func NewManager(paths *platform.Paths, store *config.Store) *Manager`
- Multiple return values standard: `([]PhpVersion, error)`, `(string, error)`
- Variadic arguments used for logging: `Info(format string, args ...any)`
- Error always last return value: `([]Version, error)` or `(string, error)` or just `error`
- Boolean for status checks: `IsInstalled() bool`, `IsEnabled() bool`
- Void functions don't return error: `Disable()` (uses side effects only)
- Multiple values when needed: `(certPath, keyPath string, error)`
## Module Design
- Capitalized first letter for exported identifiers
- Each package typically exports: `NewManager()` constructor, `Manager` struct, domain types
- Minimal exports - internal implementations stay private
- No barrel files (index.ts equivalent) in Go
- Each package is standalone - `internal/php/manager.go` is self-contained
- Related constants with their types: service labels with type definitions
## Manager Pattern
- All major subsystems implemented as `Manager` structs
- Each `Manager` holds dependencies via struct fields:
- Constructor: `func New<Service>Manager(deps) *Manager`
- Methods implement actions: `Start()`, `Stop()`, `Install()`, `List*()`
- `php.Manager` - `internal/php/manager.go`
- `nginx.Manager` - `internal/nginx/manager.go`
- `dns.Manager` - `internal/dns/manager.go`
- `sites.Manager` - `internal/sites/manager.go`
- `nodejs.Manager` - `internal/nodejs/manager.go`
- `services.LaunchdManager` - `internal/services/launchd.go`
- `certs.Manager` - `internal/certs/ca.go`
- `logs.Reader` - `internal/logs/reader.go` (Reader pattern instead of Manager)
## Concurrency Patterns
- `sync.RWMutex` used for thread-safe config storage: `internal/config/store.go`
- Lock-guard pattern: `Lock()` then `defer Unlock()` in critical sections
- `sync.Mutex` used in logger: `internal/debug/logger.go`
## Dependency Injection
- All services use constructor functions for initialization
- Example: `func NewApp() *App` creates fully initialized App with all managers
- Dependencies passed to constructors: `func NewManager(paths *platform.Paths, store *config.Store) *Manager`
- Central composition in `NewApp()`: 
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Desktop application built with Wails (Go + React)
- Event-driven service management through macOS launchd
- Stateful configuration persistence via JSON config file
- Bidirectional bridge between frontend (React) and backend (Go) via Wails IPC
- Platform-specific code path detection (Homebrew, System, Laravel Herd, Manual PHP)
## Layers
- Purpose: User interface for managing development environment
- Location: `frontend/src/`
- Contains: React components, pages, and styling
- Depends on: Wails JavaScript bindings (auto-generated from Go methods)
- Used by: User interactions
- Key files:
- Purpose: Coordinate all service managers and expose API to frontend
- Location: `app.go`
- Contains: `App` struct holding all managers, startup logic, and public methods exposed to frontend
- Depends on: All service packages
- Used by: Wails runtime for frontend IPC calls
- Initialization: Creates instances of all managers in `NewApp()`, restores state on startup
- Purpose: Manage individual services and runtime versions
- Location: `internal/*/manager.go` files
- Contains: Package-specific managers (PHP, Node, Nginx, DNS, Certs, Sites, Logs)
- Depends on: Platform paths, config store, shell integration
- Used by: Application orchestration layer
- Purpose: Provide cross-cutting functionality
- Location: `internal/platform/`, `internal/config/`, `internal/debug/`, `internal/shell/`, `internal/logs/`, `internal/services/`
## Data Flow
## Key Abstractions
- Purpose: Encapsulate lifecycle management for each service/component
- Examples: `php.Manager`, `nginx.Manager`, `dns.Manager`, `nodejs.Manager`, `sites.Manager`
- Pattern: Each manager has `Status()`, `Install()`, `Uninstall()`, `Start()`, `Stop()` methods where applicable
- Purpose: Thread-safe, persistent application state
- Examples: `internal/config/store.go`
- Pattern: RwMutex protects concurrent access; all writes trigger JSON persistence
- Purpose: Represent runtime versions and their metadata
- Examples: `php.PhpVersion`, `nodejs.NodeVersion`, `sites.Site`
- Pattern: Structs with JSON tags for serialization; carry path, binary location, active status
- Purpose: Single source of truth for all file/directory locations
- Examples: `platform.Paths` with methods like `PhpVersionDir()`, `NginxBin()`, `ConfigFile()`
- Pattern: Paths struct instantiated once; passed to managers that need it
## Entry Points
- Location: `main.go`
- Triggers: macOS app launch via Wails runtime
- Responsibilities: Creates Wails app with React frontend, embeds dist/ files, configures window properties
- Location: `app.go` - `NewApp()`
- Triggers: Called by Wails during startup
- Responsibilities: Instantiates all managers with dependencies, creates App struct
- Location: `app.go` - `startup(ctx)`
- Triggers: Wails calls on first app window creation
- Responsibilities:
- Location: `frontend/src/main.tsx`
- Triggers: Vite dev server or bundled in app
- Responsibilities: Mounts React app to DOM, initializes React StrictMode
- Location: `frontend/src/App.tsx`
- Triggers: React component render
- Responsibilities: Tab-based navigation; renders page components based on activeState
## Error Handling
- Go methods return `(result, error)` tuples
- Errors logged via Debug logger when debug mode enabled
- Frontend catches promise rejections from Wails calls
- No explicit error UI found; errors logged to console
- File system operations (permissions, missing paths)
- External command execution (Homebrew, launchd, admin elevation)
- Configuration persistence (JSON marshal/unmarshal)
- Version download failures
## Cross-Cutting Concerns
- Approach: Optional file-based debug logging via `internal/debug/logger.go`
- Enabled via `EnableDebug()` method on App
- Stores in `~/.lamboserver/logs/debug.log`
- Called at key decision points and operation completion
- Approach: Inline in manager methods
- Examples: Checking if version already installed before installing, validating domain format for sites
- No centralized validation framework
- Approach: macOS admin privilege elevation via osascript when needed
- Used for: DNS resolver configuration, launchd operations requiring root
- Prompts user with macOS security dialog
- Approach: On app startup, restore symlinks for active versions
- Ensures terminal environment matches app state
- Re-applies shell integration if removed manually
## Architecture Diagrams
```
```
```
```
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
