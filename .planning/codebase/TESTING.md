# Testing Patterns

**Analysis Date:** 2026-04-16

## Test Framework & Setup

**Go Testing:**
- Framework: Go standard `testing` package
- Assertion Library: `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`
- Mocking: `github.com/stretchr/testify/mock` for structured mocking
- Configuration: No separate config file; tests discovered via `*_test.go` naming convention

**Frontend Testing:**
- No test files present in `frontend/src/`
- No test runner configured (Jest, Vitest, etc. not in `package.json`)
- No test assertions library installed
- Frontend testing not implemented

**Run Commands:**
```bash
go test ./...                    # Run all Go tests
go test -v ./...                # Run with verbose output
go test -race ./...             # Run with race detector enabled
go test -cover ./...            # Show coverage
go test -run TestName ./...     # Run specific test
```

## Test File Organization

**Location & Naming:**
- Co-located with implementation: `manager.go` paired with `manager_test.go` in same directory
- Test-specific helpers in `testutil_test.go` files (e.g., `internal/config/testutil_test.go`)
- Integration tests in `internal/integration/` directory: `*_test.go` files

**Directory Structure:**
```
internal/
├── config/
│   ├── store.go
│   ├── store_test.go          # Unit tests for store
│   ├── testutil_test.go       # Test helpers
│   └── defaults.go
├── integration/
│   ├── startup_test.go        # Integration test
│   ├── site_creation_test.go  # Integration test
│   ├── helpers_test.go        # Shared integration test helpers
│   └── ...
├── services/
│   ├── mysql/
│   │   ├── manager.go
│   │   ├── manager_test.go
│   │   └── ...
│   └── ...
```

**Test Package Naming:**
- Unit tests: `package config_test` (underscore-separated package name)
- Integration tests: `package integration_test`
- Allows importing and testing unexported functions/types

## Test Structure & Patterns

**Unit Test Pattern:**
```go
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
		tc := tc  // capture loop variable
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			require.NoError(t, store.SetActivePhpVersion(tc.version))
			assert.Equal(t, tc.version, store.Get().ActivePhpVersion)
		})
	}
}
```

**Key Elements:**
- Table-driven tests using `[]struct` with test cases
- `t.Run()` for subtests (automatic test name hierarchies)
- `require.NoError()` for setup (fails test immediately if error)
- `assert.Equal()` for assertions (records failure but continues)
- Loop variable capture: `tc := tc` to avoid closure issues
- Helper function calls first: `store := newTestStore(t)`

**Subtests Within Test:**
```go
func TestSiteCreation_NginxConfigGenerated(t *testing.T) {
	t.Run("file created at expected path", func(t *testing.T) {
		// test body
	})

	t.Run("contains expected config", func(t *testing.T) {
		// test body
	})
}
```

**Integration Test Pattern:**
```go
// TestStartupSequence_DirectoryCreation verifies that EnsureDirectories creates all
// required subdirectories under the temp home with correct permissions (INTG-01).
func TestStartupSequence_DirectoryCreation(t *testing.T) {
	paths := newTestPaths(t)

	err := paths.EnsureDirectories()
	require.NoError(t, err, "EnsureDirectories should not return an error")

	expectedDirs := []string{
		paths.Home,
		paths.BinDir(),
		// ... more directories
	}

	for _, dir := range expectedDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			info, err := os.Stat(dir)
			require.NoError(t, err, "directory %s should exist", dir)
			assert.True(t, info.IsDir(), "%s should be a directory", dir)
		})
	}
}
```

## Setup & Teardown

**Per-Test Cleanup:**
- `t.TempDir()` automatically cleans up temporary files after test
- Used in nearly all tests: `dir := t.TempDir()`
- No manual cleanup required

**Suite-Level Setup:**
- Not used; each test is independent
- Test helpers (`newTestStore()`, `newTestPaths()`) create fresh instances per test

**Test Helper Functions:**
From `internal/config/testutil_test.go`:
```go
// newTestStore creates a Store backed by a temporary directory.
// The config file is placed at <t.TempDir()>/config.json.
// The temp dir is automatically cleaned up when the test completes.
func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	dir := t.TempDir()
	return config.NewStore(filepath.Join(dir, "config.json"))
}

// newTestStoreWithData creates a Store pre-populated with the given sites.
// Each site is added via AddSite using the site's Domain and Path fields.
func newTestStoreWithData(t *testing.T, sites []config.SiteConfig) *config.Store {
	t.Helper()
	store := newTestStore(t)
	for _, s := range sites {
		require.NoError(t, store.AddSite(s.Domain, s.Path))
	}
	return store
}
```

**Helper Pattern:**
- First line: `t.Helper()` marks function as test helper (improves error reporting)
- Parameters: `t *testing.T` as first parameter
- Returns: Fully constructed object ready for test
- No side effects or assertions in helpers

## Mocking

**Mocking Framework:** `github.com/stretchr/testify/mock`

**Mock Implementation Pattern:**
```go
// MockAdminRunner is a testify mock implementing cert.AdminRunner.
// Default: RunWithPrivileges returns nil (success) for TrustCA calls.
type MockAdminRunner struct {
	mock.Mock
}

func (m *MockAdminRunner) RunWithPrivileges(command string) error {
	args := m.Called(command)
	return args.Error(0)
}

// Compile-time interface check.
var _ cert.AdminRunner = (*MockAdminRunner)(nil)
```

**Mock Usage in Tests:**
```go
func TestSetupCA(t *testing.T) {
	admin := &MockAdminRunner{}
	admin.On("RunWithPrivileges", mock.MatchedBy(func(cmd string) bool {
		return strings.Contains(cmd, "security add-trusted-cert")
	})).Return(nil)

	certMgr := cert.NewManager(paths, rfs, admin)
	err := certMgr.SetupCA()

	require.NoError(t, err)
	admin.AssertCalled(t, "RunWithPrivileges", mock.Anything)
}
```

**Mock Assertions:**
- `mock.On()` - Expect a call with given arguments
- `mock.MatchedBy()` - Match using custom function
- `mock.Anything` - Match any value
- `.Return()` - What to return when called
- `AssertCalled()` - Verify call was made
- `AssertNotCalled()` - Verify call was NOT made

**Mocking Strategy - What to Mock:**

From `internal/integration/helpers_test.go`:
- `AdminRunner` - Avoids privilege elevation (osascript) in tests
- `LaunchdService` - Avoids real launchctl calls
- `CommandRunner` - Avoids real `exec.Command` calls for PHP detection

**Mocking Strategy - What NOT to Mock:**

From integration tests:
- `FileSystem` - Use real filesystem with `t.TempDir()`
- `Paths` - Use real paths with temp directories
- `Config` - Use real JSON file I/O (it's simple)
- `Cert` - Use real certificate generation (crypto operations)

**Real Filesystem Adapter Pattern:**
```go
// RealTestFS implements both cert.FileSystem and php.FileSystem using real os.* calls.
// All operations go against the real filesystem (t.TempDir() as root).
type RealTestFS struct {
	root string
}

func (r *RealTestFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r *RealTestFS) Create(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.Create(name)
}

// Compile-time interface checks.
var _ cert.FileSystem = (*RealTestFS)(nil)
var _ php.FileSystem = (*RealTestFS)(nil)
```

## Fixtures & Test Data

**Test Data Pattern:**
From `internal/config/testutil_test.go`:
```go
// Pre-populate store with data using table of sites
func newTestStoreWithData(t *testing.T, sites []config.SiteConfig) *config.Store {
	t.Helper()
	store := newTestStore(t)
	for _, s := range sites {
		require.NoError(t, store.AddSite(s.Domain, s.Path))
	}
	return store
}
```

**Inline Fixtures in Tests:**
```go
t.Run("adds site", func(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.AddSite("example.test", "/var/www/example"))
	// ...
})
```

**Factory Functions:**
- `newTestStore(t *testing.T) *config.Store` - Creates fresh store
- `newTestPaths(t *testing.T) *system.Paths` - Creates paths in temp dir
- `newTestCertManager(t *testing.T, paths) (*cert.Manager, *MockAdminRunner)` - Wired cert manager
- `newRealTestFS(root string) *RealTestFS` - Real filesystem adapter

## Coverage

**Requirements:** No coverage requirements enforced

**View Coverage:**
```bash
go test -cover ./...                    # Show summary coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # HTML report
```

**Observed Coverage Gaps:**
- Frontend has zero tests (no test files)
- Some manager tests exist but not comprehensive coverage
- Integration tests focus on critical workflows (startup, site creation, service lifecycle)

## Test Types

**Unit Tests:**
- Scope: Single manager or function in isolation
- Setup: Mocks for external dependencies (AdminRunner, CommandRunner)
- Real: Filesystem operations using `t.TempDir()`
- Examples:
  - `internal/config/store_test.go` - Config store operations
  - `internal/services/php/manager_test.go` - PHP version management
  - `internal/services/mysql/manager_test.go` - MySQL operations

**Integration Tests:**
- Location: `internal/integration/`
- Scope: Multiple managers working together (cert, nginx, sites, config)
- Real: Filesystem, cert generation, config persistence
- Mocked: Launchd, admin runner (privilege elevation), commands
- Examples:
  - `startup_test.go` - Directory creation, CA setup
  - `site_creation_test.go` - Nginx config, SSL certs, config persistence
  - `service_lifecycle_test.go` - Service installation and switching
  - `php_switch_test.go` - PHP version switching with FPM lifecycle

**E2E Tests:**
- Not used; no Selenium, Playwright, or Cypress
- Frontend testing not implemented

## Common Test Patterns

**Async Testing (N/A - Go):**
Not applicable; Go concurrency is synchronous in tests.

**Error Testing Pattern:**
```go
func TestStore_Load_InvalidJSON(t *testing.T) {
	// Create invalid JSON file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte("{invalid json"), 0644))

	// Load should return error
	store := config.NewStore(configPath)
	err := store.Load()
	assert.Error(t, err, "Load should error on invalid JSON")
}
```

**Concurrent Access Testing Pattern:**
```go
func TestStore_ConcurrentReadWrite(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.AddSite("a.test", "/a"))

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Get()
		}()
	}

	// Concurrent write
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = store.SetActivePhpVersion("8.3")
	}()

	wg.Wait()
	// Test completes without deadlock = success
}
```

**State Mutation Testing Pattern:**
```go
func TestStore_GetSites_ReturnsCopy(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.AddSite("copy.test", "/var/www/copy"))

	sites := store.GetSites()
	originalLen := len(sites)

	// Modify the returned slice — should not affect the internal store.
	sites = append(sites, config.SiteConfig{Domain: "extra.test"})

	assert.Len(t, store.GetSites(), originalLen, 
		"internal sites slice should be unaffected by modification of returned copy")
}
```

## Assertion Patterns

**testify/assert vs testify/require:**
- `require.NoError(t, err)` - Fail immediately if error (for setup)
- `assert.NoError(t, err)` - Record failure but continue (not used much)
- `require.Equal(t, expected, actual)` - Fail immediately if mismatch
- `assert.Equal(t, expected, actual)` - Record failure, continue
- `assert.True(t, condition)` - Simple boolean assertion
- `assert.Len(t, slice, expectedLen)` - Check slice/map length
- `assert.Contains(t, slice, item)` - Check membership

**Custom Assertion Messages:**
```go
assert.Equal(t, tc.version, store.Get().ActivePhpVersion)
assert.True(t, info.IsDir(), "%s should be a directory", dir)
assert.Len(t, store.GetSites(), 1, "expected exactly one site")
```

---

*Testing analysis: 2026-04-16*
