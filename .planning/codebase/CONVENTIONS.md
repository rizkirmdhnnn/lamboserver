# Coding Conventions

**Analysis Date:** 2026-04-17

## Naming Patterns

**Files (Go):**
- `manager.go` — primary type and methods for each service package
- `interfaces.go` — consumer-site interfaces (FileSystem, CommandRunner, LaunchdService, AdminRunner)
- `errors.go` — package-level sentinel errors
- `doc.go` — package-level documentation comment
- `service_adapter.go` — adapter implementing `services.Service` for the unified service registry
- `export_test.go` — exports internal constructors for external test packages
- `testutil_test.go` — mock definitions and test helper factories
- `*_test.go` — test files use `_test` suffix (standard Go convention)

**Files (Frontend/TypeScript):**
- Pages: `PascalCase.tsx` in `frontend/src/pages/` (e.g., `Dashboard.tsx`, `SitesPage.tsx`, `PhpPage.tsx`)
- Components: `PascalCase.tsx` in `frontend/src/components/` (e.g., `Toast.tsx`, `VersionList.tsx`)
- Entry: `main.tsx`, `App.tsx`

**Go Functions:**
- Exported: `PascalCase` — `NewManager`, `IsInstalled`, `EnsureConfig`, `SetupCA`
- Unexported helpers: `camelCase` — `ensureLogFile`, `readLastNLines`, `ensureServicesRunning`
- Constructors: `New<Type>(deps...)` pattern — `NewManager(paths, store, fs, cmd)`
- Test constructors: `newTest<Type>(t *testing.T)` — `newTestManager(t)`, `newTestPaths(t)`

**Go Variables:**
- Exported constants: `PascalCase` — `ServiceLabel`, `AppName`, `TestTLD`
- Unexported fields: `camelCase` — `paths`, `launchd`, `fs`, `admin`, `binary`

**Go Types:**
- Public structs: `PascalCase` — `Manager`, `ServiceStatus`, `PhpVersion`, `AppConfig`
- Interfaces: `PascalCase` verb-noun — `FileSystem`, `CommandRunner`, `LaunchdService`, `AdminRunner`, `HelperRunner`
- Mock types (in tests): `Mock<InterfaceName>` — `MockFileSystem`, `MockLaunchdService`, `MockAdminRunner`

**TypeScript:**
- Interfaces: `PascalCase` — `DashboardData`
- Functions: `camelCase` — `loadStatus`, `renderPage`
- Components: `PascalCase` function components — `function Dashboard()`, `function App()`

## Code Style

**Formatting (Go):**
- `gofmt` standard formatting (no custom config)
- Tab indentation (Go default)
- No `.golangci.yml` or linter config detected — relies on `gofmt` only

**Formatting (Frontend):**
- No ESLint, Prettier, or Biome config present
- TypeScript strict mode enabled in `frontend/tsconfig.json` (`"strict": true`)
- Vite build with `tsc` type checking: `"build": "tsc && vite build"`

## Import Organization

**Go Import Order (3 groups separated by blank lines):**
1. Standard library (`"context"`, `"fmt"`, `"os"`, `"crypto/rsa"`)
2. Third-party (`"github.com/stretchr/testify/assert"`, `"github.com/pkg/browser"`)
3. Internal project (`"github.com/rizkirmdhnnn/lamboserver/internal/..."`, `"github.com/rizkirmdhnnn/lamboserver/pkg/..."`)

**Import Aliases:**
- Use aliases only to resolve conflicts: `wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"` in `app.go`

**TypeScript Import Order:**
1. React/library imports (`import { useState } from "react"`)
2. Third-party icon libraries (`import { Globe, ... } from "lucide-react"`)
3. Wails bindings (`import { GetDashboardStatus, ... } from "../../wailsjs/go/main/App"`)
4. Local components/pages

**Path Aliases (TypeScript):**
- No path aliases configured — uses relative paths (`../../wailsjs/go/main/App`)

## Error Handling

**Go Error Patterns:**

1. **Wrap with context using `fmt.Errorf` and `%w`:**
   ```go
   return fmt.Errorf("failed to generate dnsmasq config: %w", err)
   return fmt.Errorf("failed to open debug log: %w", err)
   ```

2. **Sentinel errors for domain-specific failures** — defined in `errors.go` per package:
   ```go
   // internal/cert/errors.go
   var ErrCANotFound = errors.New("certificate authority not found")
   var ErrCorruptedCA = errors.New("CA file is corrupted or not a valid PEM file")
   
   // internal/services/dnsmasq/errors.go
   var ErrDnsmasqNotInstalled = errors.New("dnsmasq not installed")
   ```

3. **Log-and-continue in lifecycle methods** — `app.go` startup logs errors but does not abort:
   ```go
   if err := a.Nginx.Start(); err != nil {
       a.Debug.Error("Failed to start nginx: %v", err)
   }
   ```

4. **Return nil for idempotent operations** — `SetupCA()` returns nil if CA already exists; `Enable()` returns nil if already enabled.

5. **Action logging** — `Debug.Action(name, err)` logs both success and failure paths for user-initiated operations in `app.go`.

**TypeScript Error Patterns:**
- `try/catch` with `console.error` for Wails IPC calls:
  ```tsx
  try {
      const data = await GetDashboardStatus();
      setStatus(data);
  } catch (e) {
      console.error("Failed to load dashboard:", e);
  }
  ```

## Logging

**Framework:** Custom `pkg/logger.Logger` — file-based, disabled by default

**Log Levels:**
- `Info(format, args...)` — informational messages
- `Error(format, args...)` — error messages
- `Action(name, err)` — structured action result logging (logs OK or FAILED)

**Log Format:** `[YYYY-MM-DD HH:MM:SS.mmm] [LEVEL] message`

**When to Log (Go backend):**
- Log all service lifecycle events: start, stop, restart
- Log startup sequence steps: helper install, CA setup, shell integration
- Use `Debug.Action(name, err)` for every user-initiated action in `app.go`
- Use `Debug.Info/Error` for internal operations and diagnostics

**Frontend:** Uses `console.error` only for IPC call failures. No structured logging.

## Comments

**Package Documentation:**
- Every package has a `doc.go` file with a package-level doc comment explaining the package's purpose and scope
- Pattern: `// Package <name> <verb phrase describing purpose>.`
- Example from `internal/services/dnsmasq/doc.go`:
  ```go
  // Package dns manages dnsmasq for local DNS resolution in LamboServer.
  //
  // The Manager type handles dnsmasq installation detection, writes resolver
  // configuration files under /etc/resolver/ for the local TLD...
  ```

**Type and Method Comments:**
- All exported types and methods have doc comments
- Method comments start with the method name: `// Start installs nginx as a root LaunchDaemon...`
- Struct comments describe purpose and usage: `// Manager handles dnsmasq binary discovery...`

**Section Dividers in app.go:**
- Unicode box-drawing characters for visual section separation:
  ```go
  // -- Lifecycle ─────────────────────────────────────────────────────────
  // -- Generic Service Methods ───────────────────────────────────────────
  ```

**Inline Comments:**
- Reference tracking codes for cross-cutting decisions: `// D-08: pgweb must stop when PostgreSQL stops`
- Explain "why" not "what": `// Ensure log files exist with user ownership`

**Compile-Time Interface Checks:**
- Every service adapter includes a compile-time interface assertion:
  ```go
  var _ services.Service = (*ServiceAdapter)(nil)
  ```
- Real implementations also include compile-time checks:
  ```go
  var _ FileSystem = system.RealFS{}
  var _ CommandRunner = system.RealCmdRunner{}
  ```

## Function Design

**Constructor Pattern:**
- All managers use constructor injection with explicit dependencies:
  ```go
  func NewManager(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager
  ```
- Dependencies are interfaces (except `*system.Paths` and `*config.Store` which are concrete)

**Method Size:**
- Service manager methods are typically 10-30 lines
- `app.go` methods are thin delegation: validate input, call manager, log result, return

**Return Values:**
- Single error return for mutation operations: `Start() error`, `Stop() error`
- Value + error for queries: `ListInstalled() ([]PhpVersion, error)`
- Struct return for status: `Status() ServiceStatus`

## Module Design

**Package Organization:**
- One manager type per package: `dnsmasq.Manager`, `nginx.Manager`, `php.Manager`
- Interfaces defined at consumer site (not in a shared contracts package) per ADR-001
- `internal/` for application-specific code, `pkg/` for reusable utilities

**Service Adapter Pattern:**
- Each service has a `ServiceAdapter` struct wrapping `*Manager` to implement `services.Service`
- Adapters live in the same package as the manager: `dnsmasq.ServiceAdapter`, `nginx.ServiceAdapter`
- Created via `NewServiceAdapter(m *Manager)` factory function

**Barrel Files / Exports:**
- No barrel files — each package is imported directly
- `export_test.go` pattern exposes internal constructors for external test packages:
  ```go
  // export_test.go
  func NewManagerWithDeps(...) *Manager { return newManagerWithDeps(...) }
  ```

## Architecture Decision Records

- ADRs live in `docs/adr/` directory
- Follow format: `ADR-NNN-<slug>.md` (e.g., `ADR-001-interface-adoption.md`)
- Reference codes (D-01, D-02, D-08) used in source comments to trace back to decisions

---

*Convention analysis: 2026-04-17*
