# ADR-001: Interface Adoption for Testability

**Status:** Accepted
**Date:** 2026-04-15
**Deciders:** Project maintainer

## Context

LamboServer service managers (PHP, Nginx, DNS, Node, Sites, Certs) made direct system calls —
filesystem operations (`os.ReadDir`, `os.WriteFile`), process execution (`exec.Command`), and
launchd daemon management — without any abstraction layer. This made unit testing impossible:
running tests required a real macOS system with launchd, Homebrew, and actual PHP/Nginx
installations present.

The root problem was a lack of seams for test isolation. Without interfaces, any test of, say,
`php.Manager.ListInstalled()` would actually invoke `os.ReadDir` on filesystem paths that
don't exist in CI environments. There was no way to substitute fake implementations.

## Decision

Introduce fine-grained interfaces defined at the consumer site (per D-02 from Phase 1 context).
Each service manager defines exactly the interface subset it needs from filesystem, exec, and
launchd. No central contracts package.

**Interface placement:** Each package owns its interfaces in `interfaces.go`:
- `php.FileSystem` — Stat, ReadDir, MkdirAll, WriteFile, Remove, RemoveAll, Chmod
- `php.CommandRunner` — Run, LookPath
- `php.LaunchdService` — Install, Uninstall, IsRunning
- Similar fine-grained interfaces in `nginx/`, `dns/`, `site/`, `cert/`, `node/`

**Real implementations** live in `internal/system/`:
- `system.RealFS{}` — delegates to `os` package
- `system.RealCmdRunner{}` — delegates to `exec.Command`
- `system.RealAdminRunner{}` — delegates to osascript/admin-elevated exec

**Constructor injection** (per D-05): All managers accept their interface dependencies as
constructor arguments. Example:

```go
php.NewManager(paths *system.Paths, store *config.Store, fs FileSystem, cmd CommandRunner) *Manager
```

No global state. No `init()` side effects.

**Test fakes** are per-package in `testutil_test.go` files (per D-04). Fakes are test-only,
not exported from packages. This keeps production code free of test infrastructure.

**Build tags** retained for platform-specific code (per D-03). `_darwin.go` files continue to
exist for launchd, osascript, and keychain operations. Interface implementations are
platform-specific at compile time; tests use fake implementations without build constraints.

**Wiring** happens in `NewApp()` — the single composition root in `app.go`. All real
implementations are instantiated once and passed down:

```go
fs   := system.RealFS{}
cmd  := system.RealCmdRunner{}
admin := system.RealAdminRunner{}
// ...
php.NewManager(paths, store, fs, cmd)
```

## Consequences

### Positive

- Every service manager is unit-testable without a real macOS system. Tests pass fake
  implementations that record calls or return controlled errors.
- Interfaces document exactly what each manager needs from its environment — the interface
  declaration is an implicit contract.
- New implementations (e.g., a mock for CI, an alternative backend) are trivially substituted
  by satisfying the interface structurally (Go duck typing).
- Composition root (`NewApp()`) is the single place that knows about real implementations.
  Everything else is isolated from platform specifics.

### Negative

- Each package defines its own interface subset. Some method signatures appear in multiple
  `interfaces.go` files (e.g., `Stat`, `WriteFile`). This is mild duplication — acceptable
  because the interfaces are per-consumer and serve different purposes.
- Constructor signatures grew longer (4-6 arguments vs 2 previously). More verbose but
  explicit and unambiguous.
- Fakes must be maintained alongside source changes. When `RealFS` gains a new method,
  packages that need it must update their interface and fakes.

### Neutral

- Go's structural typing means interfaces are satisfied implicitly. No `implements` keyword
  required. Adding a method to an interface breaks all existing fakes at compile time —
  this is a feature (compile-time safety) not a drawback.
- The `internal/system/` package now contains both interface implementations (RealFS, etc.)
  and platform helpers. This is consistent with the existing layout.
