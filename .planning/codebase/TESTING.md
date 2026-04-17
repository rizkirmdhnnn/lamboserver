# Testing Patterns

**Analysis Date:** 2026-04-17

## Test Framework

**Runner:**
- Go standard `testing` package
- No additional test runner (no `gotestsum`, no `goconvey`)

**Assertion Library:**
- `github.com/stretchr/testify` v1.11.1
  - `assert` — soft assertions (test continues on failure)
  - `require` — hard assertions (test stops on failure)
  - `mock` — mock generation with call tracking and expectations

**Run Commands:**
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -count=1 ./...     # Skip cache
go test -cover ./...       # With coverage
go test -race ./...        # Race condition detection
go test ./internal/cert    # Single package
```

## Test File Organization

**Location:**
- Co-located with source files in the same package directory
- Integration tests in dedicated package: `internal/integration/`

**Naming:**
- `*_test.go` for test files (Go convention)
- `testutil_test.go` for mock definitions and test helper factories
- `export_test.go` for exposing internal constructors to external test packages

**Structure per package:**
```
internal/services/dnsmasq/
├── manager.go              # Production code
├── interfaces.go           # Consumer-site interfaces
├── errors.go               # Sentinel errors
├── service_adapter.go      # Adapter for services.Service
├── doc.go                  # Package documentation
├── export_test.go          # Exports internals for _test package
├── testutil_test.go        # Mocks + test helpers
└── manager_test.go         # Unit tests
```

**Test Package Convention:**
- Unit tests use external test package: `package dnsmasq_test` (not `package dnsmasq`)
- This enforces testing through the public API only
- `export_test.go` bridges internal access when needed (uses `package dnsmasq` to export unexported constructors)

## Test Structure

**Suite Organization:**
```go
// Individual test functions — no suite/table grouping except where natural
func TestManager_IsDaemonInstalled_True(t *testing.T) { ... }
func TestManager_IsDaemonInstalled_False(t *testing.T) { ... }
func TestManager_EnsureConfig(t *testing.T) { ... }

// Table-driven tests for parameter variations
func TestStore_SetActivePhpVersion(t *testing.T) {
    tests := []struct {
        name    string
        version string
    }{
        {name: "sets version 8.3", version: "8.3"},
        {name: "overwrites with 8.4", version: "8.4"},
        {name: "sets empty string", version: ""},
    }
    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            store := newTestStore(t)
            require.NoError(t, store.SetActivePhpVersion(tc.version))
            assert.Equal(t, tc.version, store.Get().ActivePhpVersion)
        })
    }
}
```

**Naming Convention for Tests:**
- `Test<Type>_<Method>_<Scenario>` — e.g., `TestManager_EnsureResolver_AlreadyCorrect`
- Test function doc comments describe what is verified, referencing ticket codes:
  ```go
  // TestStartupSequence_DirectoryCreation verifies that EnsureDirectories creates all
  // required subdirectories under the temp home with correct permissions (INTG-01).
  ```

**Patterns:**
- **Setup:** Use `t.Helper()` in all factory functions; use `t.TempDir()` for filesystem isolation
- **Assertions:** Use `require` for preconditions (stops test), `assert` for actual checks (continues)
- **Teardown:** Relies on `t.TempDir()` auto-cleanup; no explicit teardown needed

## Mocking

**Two Mocking Styles Used:**

### Style 1: Hand-rolled function-based mocks (dnsmasq, logger)
Used in: `internal/services/dnsmasq/testutil_test.go`

```go
type MockFileSystem struct {
    StatFunc      func(name string) (fs.FileInfo, error)
    WriteFileFunc func(name string, data []byte, perm fs.FileMode) error
    ReadFileFunc  func(name string) ([]byte, error)
    MkdirAllFunc  func(name string, perm fs.FileMode) error

    // Call tracking
    StatCalls      []string
    WriteFileCalls []WriteFileCall
    ReadFileCalls  []string
    MkdirAllCalls  []string
}

func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
    m.StatCalls = append(m.StatCalls, name)
    if m.StatFunc != nil {
        return m.StatFunc(name)
    }
    return &mockFileInfo{name: filepath.Base(name)}, nil
}
```

**Characteristics:** Explicit call tracking slices, optional function overrides, sensible defaults.

### Style 2: testify/mock-based mocks (php, nginx, cert, integration)
Used in: `internal/services/php/testutil_test.go`, `internal/services/nginx/testutil_test.go`, `internal/integration/helpers_test.go`

```go
type MockFileSystem struct {
    mock.Mock
}

func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
    args := m.Called(name)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(fs.FileInfo), args.Error(1)
}
```

**Characteristics:** Uses `mock.On(...)` for expectations, `mock.MatchedBy(...)` for flexible matching, `mock.AssertExpectations(t)` for verification.

**What to Mock:**
- Filesystem operations (`os.Stat`, `os.ReadFile`, `os.WriteFile`)
- Command execution (`exec.Command`, `exec.LookPath`)
- Launchd service lifecycle (`Install`, `Uninstall`, `IsRunning`)
- Admin privilege elevation (`RunWithPrivileges`)

**What NOT to Mock:**
- `system.Paths` — always use real struct with `Home: t.TempDir()`
- `config.Store` — use real implementation backed by temp file
- Cryptographic operations in cert tests — use real crypto, mock only filesystem

## Fixtures and Factories

**Test Data Factories (in testutil_test.go files):**

```go
// Creates Paths rooted in temp directory
func newTestPaths(t *testing.T) *system.Paths {
    t.Helper()
    return &system.Paths{Home: t.TempDir()}
}

// Creates a config.Store backed by temp file
func newTestStore(t *testing.T) *config.Store {
    t.Helper()
    dir := t.TempDir()
    return config.NewStore(filepath.Join(dir, "config.json"))
}

// Creates Manager with all mocked dependencies
func newTestManager(t *testing.T) (*dnsmasq.Manager, *MockFileSystem, *MockLaunchdService, *MockAdminRunner, *system.Paths) {
    t.Helper()
    paths := newTestPaths(t)
    mockFS := &MockFileSystem{}
    mockLaunchd := &MockLaunchdService{}
    mockAdmin := &MockAdminRunner{}
    mgr := dnsmasq.NewManagerWithDeps(paths, mockLaunchd, mockFS, mockAdmin)
    return mgr, mockFS, mockLaunchd, mockAdmin, paths
}
```

**Integration Test Filesystem (RealTestFS):**
```go
// RealTestFS implements both cert.FileSystem and php.FileSystem using real os.* calls.
// All operations go against the real filesystem (t.TempDir() as root).
type RealTestFS struct {
    root string
}
```
Located in: `internal/integration/helpers_test.go`

**Location:**
- Each package has its own `testutil_test.go` for mocks/factories
- Integration test helpers in `internal/integration/helpers_test.go`
- No shared test fixtures directory

## Coverage

**Current Coverage (2026-04-17):**

| Package | Coverage |
|---------|----------|
| `internal/cert` | 83.6% |
| `internal/config` | 84.8% |
| `internal/sites` | 85.4% |
| `pkg/logger` | 65.3% |
| `internal/services/pgweb` | 85.0% |
| `internal/services/phpmyadmin` | 64.1% |
| `internal/services/nginx` | 55.4% |
| `internal/services/postgres` | 50.6% |
| `internal/services/mysql` | 47.0% |
| `internal/services/php` | 45.9% |
| `internal/services/nodejs` | 38.6% |
| `internal/services/dnsmasq` | 37.5% |
| `internal/system` | 36.6% |
| `internal/process` | 6.8% |
| `main` (app.go) | 0.0% |
| `internal/binaries` | 0.0% |
| `internal/services` (manager.go) | 0.0% |
| `internal/tray` | 0.0% (no tests) |
| `pkg/notify` | 0.0% (no tests) |

**Requirements:** No enforced coverage thresholds. No CI coverage gates.

**View Coverage:**
```bash
go test -cover ./...                          # Summary
go test -coverprofile=coverage.out ./...      # Generate profile
go tool cover -html=coverage.out              # HTML report
```

## Test Types

**Unit Tests:**
- Present in most `internal/` packages
- Test individual manager methods through public API
- Use mocked dependencies (filesystem, command runner, launchd)
- Files: `internal/services/*/manager_test.go`, `internal/config/store_test.go`, `pkg/logger/logger_test.go`

**Integration Tests:**
- Dedicated package: `internal/integration/`
- Exercise multi-component workflows with real filesystem in `t.TempDir()`
- Mock only privileged operations (launchd, admin runner)
- Files:
  - `internal/integration/startup_test.go` — directory creation, CA setup, symlink restore, shell integration
  - `internal/integration/php_switch_test.go` — PHP version switching workflow
  - `internal/integration/site_creation_test.go` — site creation with SSL certs
  - `internal/integration/service_lifecycle_test.go` — service start/stop/restart

**E2E Tests:**
- Not present. No browser/UI testing for the React frontend.
- No Wails IPC integration tests.

**Frontend Tests:**
- Not present. No test framework configured in `frontend/package.json`.
- No Jest, Vitest, or React Testing Library.

**Concurrency Tests:**
- `TestLogger_ConcurrentInfoCalls` in `pkg/logger/logger_test.go` — verifies mutex safety with 20 goroutines
- Run with `-race` flag for detection

## Common Patterns

**Async Testing:**
```go
// Concurrency safety test pattern
func TestLogger_ConcurrentInfoCalls(t *testing.T) {
    dir := t.TempDir()
    l := logger.NewLogger(dir)
    require.NoError(t, l.Enable())

    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            l.Info("msg %d", n)
        }(i)
    }
    wg.Wait()
    l.Disable()
}
```

**Error Testing:**
```go
func TestManager_EnsureResolver_AdminFails(t *testing.T) {
    adminErr := errors.New("admin authentication failed")
    mockAdmin := &MockAdminRunner{
        RunWithPrivilegesFunc: func(command string) error {
            return adminErr
        },
    }
    mgr := dnsmasq.NewManagerWithDeps(paths, &MockLaunchdService{}, mockFS, mockAdmin)

    err := mgr.EnsureResolver()
    require.Error(t, err)
    assert.ErrorIs(t, err, adminErr)
}
```

**Idempotency Testing:**
```go
func TestManager_SetupCA_Idempotent(t *testing.T) {
    mgr, paths, _ := newTestCertManager(t)
    require.NoError(t, mgr.SetupCA())
    certData1, _ := os.ReadFile(paths.CACert())
    
    require.NoError(t, mgr.SetupCA()) // second call
    certData2, _ := os.ReadFile(paths.CACert())
    
    assert.Equal(t, certData1, certData2, "should be unchanged on second call")
}
```

**Filesystem Isolation:**
```go
// Every test that touches the filesystem uses t.TempDir()
func newTestPaths(t *testing.T) *system.Paths {
    t.Helper()
    return &system.Paths{Home: t.TempDir()}
}
```

**Creating Fake Binaries for Testing:**
```go
// Create fake PHP binary so BinaryLocator.IsInstalled() returns true
phpBin := filepath.Join(phpVersionDir, "bin", "php")
require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0755))
require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh\necho PHP/8.3.0"), 0755))
```

## Test Coverage Gaps

**No tests:**
- `app.go` (797 lines, 0% coverage) — the composition root and all Wails-bound methods
- `internal/tray/tray.go` — system tray integration
- `internal/binaries/` — binary downloader and registry
- `pkg/notify/` — notification system
- `internal/services/manager.go` — unified service registry

**Low coverage (<40%):**
- `internal/process/` (6.8%) — process runner, pidfile, port, plist, launchd
- `internal/system/` (36.6%) — shell integration, launchd manager, helper runner, darwin-specific code
- `internal/services/dnsmasq/` (37.5%) — Start path not fully tested (binary install)
- `internal/services/nodejs/` (38.6%)

**Frontend:**
- Zero test infrastructure — no test runner, no test files, no assertions
- All 8 page components and 2 shared components are untested

---

*Testing analysis: 2026-04-17*
