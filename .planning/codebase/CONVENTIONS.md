# Coding Conventions

**Analysis Date:** 2026-04-16

## Naming Patterns

**Go Files:**
- Lowercase with underscores for multi-word files: `store_test.go`, `shell_unix.go`, `helper.go`
- Package-level doc files: `doc.go` - contains package documentation
- Service-specific files grouped by manager role: `manager.go`, `manager_test.go`, `interfaces.go`, `service_adapter.go`

**Go Functions:**
- PascalCase for exported functions: `NewManager()`, `SetActivePhpVersion()`, `GetDashboardStatus()`
- camelCase for unexported functions: `saveLocked()`, `newTestStore()`
- Descriptive action-oriented names: `Install()`, `Start()`, `Stop()`, `Restart()`
- Test functions use `TestComponentName_Scenario` pattern: `TestStore_NewStore_Defaults()`, `TestStore_SetActivePhpVersion()`

**Go Variables & Constants:**
- PascalCase for exported: `StatusRunning`, `AppConfig`, `ServiceStatus`
- camelCase for unexported: `config`, `filePath`, `mu` (mutex)
- Constants uppercase: `AppName`, `AppDirName`, `TestTLD`
- Inline template constants: `const siteConfTemplate = ...`, `const plistTemplate = ...`

**Go Types & Structs:**
- PascalCase always: `AppConfig`, `SiteConfig`, `ServiceStatus`, `Manager`
- Struct fields exported (PascalCase) with JSON tags for serialization: `Domain`, `ActivePhpVersion`
- Interface names ending with verb-noun pattern or role: `FileSystem`, `CommandRunner`, `AdminRunner`, `VersionedService`

**TypeScript/React:**
- Component files: PascalCase (`Dashboard.tsx`, `PhpPage.tsx`, `SitesPage.tsx`, `VersionList.tsx`)
- Page files in `frontend/src/pages/`: PascalCase ending with `Page` (`Dashboard.tsx`, `DatabasePage.tsx`)
- Props interfaces: `<ComponentName>Props` pattern (e.g., `interface DashboardData {}` for data structures)
- Hook-like state variables: camelCase (`status`, `versions`, `installing`, `showAvailable`)
- Event handlers: `handle<Action>` pattern (`handleInstall`, `handleSetActive`, `loadVersions`)

## Code Style

**Go Formatting:**
- Standard `gofmt` style (enforced by Go tooling)
- Indentation: 1 tab per level
- Line length: No hard limit enforced, but follows Go conventions (typically ~120 chars)
- No special linter config detected; uses Go defaults

**TypeScript/React Formatting:**
- No Prettier config file present (`tsconfig.json` only, no `.prettierrc`)
- Vite-based build (frontend uses `@vitejs/plugin-react`)
- Strict TypeScript mode enabled in `tsconfig.json`: `"strict": true`
- JSX syntax: `jsx: "react-jsx"` (React 17+ automatic runtime)

**Comments:**
- Package-level documentation in `doc.go` files following Go conventions
- Single-line comments for explaining interface contracts: `// FileSystem is the filesystem subset...`
- Test comments describing scenarios: `// --- AddSite ---` as section markers
- Inline comments for non-obvious logic (e.g., `// Modify the returned slice — should not affect the internal store`)
- No JSDoc/TSDoc consistently applied in frontend

## Import Organization

**Go Imports:**
1. Standard library imports (`fmt`, `os`, `sync`, `encoding/json`)
2. Third-party imports (`github.com/stretchr/testify`, `github.com/wailsapp/wails/v2`)
3. Internal imports (`github.com/rizkirmdhnnn/lamboserver/internal/*`, `github.com/rizkirmdhnnn/lamboserver/pkg/*`)

Blank line separates each group. Example from `app.go`:
```go
import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/browser"
	"github.com/rizkirmdhnnn/lamboserver/internal/..."
	"github.com/rizkirmdhnnn/lamboserver/pkg/logger"
	"github.com/wailsapp/wails/v2/..."
)
```

**TypeScript Imports:**
- React hooks from `"react"`: `import { useEffect, useState } from "react"`
- Lucide icons from `"lucide-react"`: `import { LayoutDashboard, Globe, ... } from "lucide-react"`
- Wails-generated bindings from `"../../wailsjs/go/main/App"`
- Relative imports for local pages and components: `import Dashboard from "./pages/Dashboard"`

## Error Handling

**Go Error Pattern:**
- Standard Go idiom: All mutating operations return `error` as last return value
- Functions like `Load()`, `Save()`, `SetActivePhpVersion()`, `AddSite()` return error
- Read-only methods return values without error (e.g., `Get()` returns `AppConfig` directly)
- Errors propagated up to caller: `if err != nil { return err }`
- No panic in application code; errors returned to Wails frontend binding layer

**Frontend Error Pattern:**
- Try-catch blocks in async handlers with `console.error()` fallback
- Example pattern from `PhpPage.tsx`:
```typescript
const loadVersions = async () => {
  try {
    const data = await GetPhpVersions();
    setVersions(data || []);
  } catch (e) {
    console.error(e);
  }
};
```
- Errors silently logged; no error state UI currently implemented

## Logging

**Go Framework:** No third-party logging library; uses `pkg/logger` custom package

**Frontend Logging:** Console-only via `console.error()` and `console.log()`
- Used for debugging API calls and state changes
- No error handling UI; errors disappear silently

## Comments & Documentation

**Go Package Documentation:**
- Every package has `doc.go` file with package-level comments
- Format: `// Package <name> <description>` followed by detailed explanation
- Example from `internal/config/doc.go`:
```go
// Package config provides thread-safe, persistent application configuration
// storage for LamboServer.
//
// The Store type manages AppConfig (active PHP/Node versions, site mappings,
// Nginx ports, and debug mode) using sync.RWMutex for concurrent access.
```

**Exported Type Comments:**
- Every exported interface and type has a comment explaining its purpose
- Comments placed immediately before the type definition
- Example from `internal/services/service.go`:
```go
// Service is the base interface for all daemon-backed services.
// Services that run as background processes (Nginx, DNSMasq, MySQL,
// PostgreSQL, CloudBeaver) implement this interface.
type Service interface { ... }
```

**Test Comments:**
- Test categories marked with `// --- Category ---` comments
- Scenario descriptions above each test function
- Example from `internal/config/store_test.go`:
```go
// --- NewStore / Defaults ---

func TestStore_NewStore_Defaults(t *testing.T) { ... }
```

**Frontend Comments:**
- Minimal comments in `.tsx` files
- Navigation logic comments explaining intent: `// Database dropdown — parent toggles only, does not navigate (per D-02)`

## Function Design

**Go Function Size:**
- Small, focused functions with single responsibility
- Managers typically 50-150 lines per public method
- Helper functions extracted for complex operations (e.g., `saveLocked()` helper)

**Go Function Parameters:**
- Receiver pointer pattern for methods: `func (m *Manager) Method() ...`
- Dependency injection pattern: manager receives interfaces (FileSystem, CommandRunner) in constructor
- Test helpers use `t *testing.T` as first parameter with `t.Helper()` call

**Go Return Values:**
- Single return value for read operations: `Get() AppConfig`
- Error as last return for mutating operations: `SetActivePhpVersion(version string) error`
- Multiple returns for queries: `(value, ok bool)` or `([]T, error)`

**TypeScript/React Function Design:**
- React functional components using hooks
- State management via `useState()`
- Effect-based data loading via `useEffect()`
- Event handlers use arrow functions: `const handleInstall = async (version: string) => { ... }`

## Module Design

**Go Package Structure:**
- Each package defined by directory: `internal/config/`, `internal/services/`, etc.
- Exported types and functions grouped with interfaces at package level
- Service managers follow adapter pattern: `ServiceAdapter`, `Manager`, and interfaces in separate files
- `doc.go` serves as package entrypoint documentation

**Exported Items per Package:**
- `Manager` structs (service orchestrators)
- Service `interface` types (contracts for dependency injection)
- Domain `struct` types (e.g., `Site`, `AppConfig`, `SiteConfig`)
- Factory functions: `NewManager()`, `NewStore()`

**Unexported Helpers:**
- Unexported helper types (e.g., `siteTemplateData`)
- Unexported constructor helpers (e.g., `saveLocked()`)
- Unexported implementation details (e.g., internal state fields with lowercase names)

**Barrel Files (Re-exports):**
- Not used in this codebase
- Each package maintains its own export list

**Frontend Module Structure:**
- Pages in `frontend/src/pages/`: Stateful page components
- Components in `frontend/src/components/`: Reusable UI components
- Wails bindings auto-generated in `frontend/wailsjs/` (not edited manually)
- Main app router in `frontend/src/App.tsx`

## Interface Design (Go)

**Defined at Consumer (Per D-02 Design Decision):**
- Each package defines the interfaces it needs, not what it provides
- Example: `internal/sites/interfaces.go` defines `FileSystem` interface for sites, not at filesystem layer
- This enables loose coupling and testability via mocking

**Small, Focused Interfaces:**
- Interfaces typically 3-6 methods
- Example `sites.FileSystem`:
```go
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	Create(name string) (*os.File, error)
	Remove(name string) error
}
```

**Service Registration Pattern:**
- Central `Manager` in `internal/services/manager.go` holds service registry
- Services implement `Service`, `VersionedService`, or `WebAdminService` interfaces
- Registration via `Register()` and `RegisterWebAdmin()` methods
- Lookup by name: `Get(name)`, `GetVersioned(name)`, `GetWebAdmin(name)`

---

*Convention analysis: 2026-04-16*
