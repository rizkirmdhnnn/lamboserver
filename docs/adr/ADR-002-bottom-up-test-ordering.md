# ADR-002: Bottom-Up Test Ordering

**Status:** Accepted
**Date:** 2026-04-15
**Deciders:** Project maintainer

## Context

LamboServer had zero test files when the refactor started. Adding tests to a working production
application carries risk: tests might catch regressions introduced during refactoring, but the
act of refactoring (to make code testable) might introduce regressions itself.

The challenge: where to start? Options considered:

1. **Top-down** — Test app-layer methods first (the Wails-exposed public API). Problem: these
   methods depend on all service managers, which depend on filesystem and platform primitives.
   Mocking everything at once is brittle and produces tests that test mock behavior, not real
   logic.

2. **Service manager first** — Test PHP, Nginx, etc. directly. Problem: these depend on
   `system.Paths` and `config.Store`, which are not yet tested. Bugs in dependencies would
   surface as confusing manager test failures.

3. **Bottom-up** — Test infrastructure first (packages with fewest dependencies), then service
   managers, then integration. Problem: "interesting" service logic is tested last. Benefit:
   each layer's tests build on proven foundations.

4. **Random / opportunistic** — Test whatever seems easiest. Problem: no systematic coverage;
   easy to miss critical paths.

## Decision

Adopt bottom-up test ordering across the entire refactor roadmap.

**Phase ordering:**
1. **Phase 1 (Testable Foundation):** No tests written. Introduce interfaces and DI so tests
   become possible. Establish the seams needed for isolation.
2. **Phase 2 (Infrastructure Tests):** Test the four infrastructure packages first:
   `internal/config/` (config store), `internal/system/paths.go` (paths),
   `pkg/logger/` (logger), `internal/system/shell.go` (shell integration).
   These have zero external dependencies — no filesystem mocking needed at this level.
3. **Phase 3 (Service Manager Tests):** Test each service manager using the infrastructure
   proven in Phase 2. Fakes substitute `RealFS`, `RealCmdRunner`, `RealAdminRunner`.
4. **Phase 4 (Integration Tests):** Test critical end-to-end workflows through the full stack.
5. **Phase 5 (Documentation):** Capture decisions and provide developer onboarding.

**Test structure decisions (per Phase 2 context D-01, D-02):**
- One test file per source file: `store.go` → `store_test.go`
- Table-driven tests with `t.Run` subtests. `[]struct{ name string; ... }` tables are
  Go-idiomatic and easy to extend without duplicating setup.
- Test helpers and mocks in per-package `testutil_test.go` files. Not exported.

**Fake strategy (per D-03):** Use testify/mock (`mock.On().Return()`) for interface fakes.
The project already depended on `github.com/stretchr/testify` — no new dependency needed.

**Concurrency testing (per D-05):** Config store concurrent tests use the `-race` flag with
10-20 goroutines doing simultaneous reads and writes. This validates the `sync.RWMutex` design
under real concurrent pressure, not just sequential test correctness.

**Coverage approach (per D-06):** Full path coverage — every exported function and every error
branch. No numeric threshold. `go test -cover` used to verify all code paths are exercised.

## Consequences

### Positive

- Each phase leaves the application in a working state. The app runs and behaves correctly
  after every phase, even before testing is complete. No "big bang" risky refactor.
- Infrastructure interfaces are validated by tests before service managers depend on them.
  A bug in `config.Store` surfaces in Phase 2 infrastructure tests, not as a confusing Phase 3
  manager test failure.
- Developers joining mid-project can understand what has been verified: the test files
  document exactly what each package does and what edge cases matter.
- Infrastructure first means: if `Paths.PhpVersionDir()` is wrong, the `php.Manager` tests
  that depend on it will have correct expectations because `paths_test.go` already verified it.

### Negative

- "Interesting" service logic (PHP detection, Nginx config generation, site creation) is tested
  later in the roadmap. Phase 2 tests are arguably less exciting than testing the business
  logic. This is intentional — correctness before features.
- Bottom-up ordering means a bug in a service manager that only manifests under concurrent
  load might not be caught until Phase 3 or 4.
- Table-driven tests with subtests are more verbose than simple test functions. The structure
  pays off when test cases expand, but has upfront ceremony.

### Neutral

- The roadmap is explicitly sequenced. Each phase document captures context and decisions so
  future phases understand the foundation they build on.
- Phase 1 produced no test files — its value is enabling Phase 2+ by introducing the
  interfaces and DI patterns that make isolated tests possible.
