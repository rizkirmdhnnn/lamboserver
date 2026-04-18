# ADR-003: App Orchestration Layer Split

**Status:** Accepted
**Date:** 2026-04-15
**Deciders:** Project maintainer

## Context

The `App` struct in `app.go` was simultaneously the composition root (wiring all managers),
the lifecycle container (Wails startup/shutdown hooks), and the owner of all domain methods
exposed to the frontend. This concentration made `app.go` a large, hard-to-navigate file.

Finding a PHP method required scanning past Nginx, DNS, Node, and debug methods. Adding a new
domain method meant editing the same file as unrelated domains, creating merge conflicts in
team settings. Reading the file to understand one domain required context-switching through
all others.

Partial domain splits (`app_php.go`, `app_node.go`, `app_services.go`, `app_dashboard.go`,
`app_debug.go`) already existed — but they were incomplete. Cert methods remained in
`app_services.go` alongside Nginx and DNS. The split was informal, with no clear rule for
which file a new method should go in.

The `NewApp()` function also lacked all real implementations — it had not yet been wired to
pass `RealFS`, `RealCmdRunner`, and `RealAdminRunner` to the managers that needed them.

## Decision

Complete the app-layer split and harden `NewApp()` as the sole composition root.

**Domain split principle:** One `app_<domain>.go` file per service manager domain:
- `app_php.go` — PHP version methods (`ListPhpVersions`, `InstallPhp`, `UninstallPhp`,
  `SetActivePhp`, `GetPhpFpmStatus`, `StartPhpFpm`, `StopPhpFpm`, `RestartPhpFpm`)
- `app_node.go` — Node.js version methods
- `app_services.go` — Nginx and DNS service lifecycle methods
- `app_dashboard.go` — Dashboard status aggregation
- `app_debug.go` — Debug logging toggle methods
- `app_cert.go` — Certificate and CA methods (`IsCAInstalled`, `SetupCA`)

**`app.go` responsibilities are narrowed** to:
- `App` struct definition (fields for all managers)
- `NewApp()` composition root
- Wails lifecycle hooks (`startup`, `beforeClose`, `shutdown`)
- Private helper methods used by multiple domains (`restoreSymlinks`, `cleanupStaleAgents`,
  `ensureServicesRunning`)

**`NewApp()` as single composition root** (per D-07 from Phase 1 context): All real
implementations are created here and nowhere else. The wiring is explicit:

```go
fs    := system.RealFS{}
cmd   := system.RealCmdRunner{}
admin := system.RealAdminRunner{}
certMgr := cert.NewManager(paths, fs, admin)

return &App{
    Php:    php.NewManager(paths, store, fs, cmd),
    PhpFpm: php.NewFpmManager(paths, store, launchd, fs),
    Nginx:  nginx.NewManager(paths, launchd, fs, launchd.Helper()),
    Dns:    dns.NewManager(paths, launchd, fs, admin),
    Certs:  certMgr,
    Sites:  site.NewManager(paths, store, certMgr, fs),
    Node:   node.NewManager(paths, store, fs, cmd),
    // ...
}
```

**Thin delegation** (per D-08): App methods are thin wrappers. They log entry, delegate to the
manager, log completion, and return:

```go
func (a *App) StartNginx() error {
    a.Debug.Info("StartNginx called")
    err := a.Nginx.Start()
    a.Debug.Action("StartNginx", err)
    return err
}
```

No business logic in App methods. Business logic lives in managers.

**Public method signatures unchanged:** Wails auto-generates TypeScript bindings from public
methods on `App`. Changing signatures breaks the frontend. The split is purely a file
organization change — no signature changes, no behavior changes.

**`app_cert.go` extraction:** `IsCAInstalled()` and `SetupCA()` moved from `app_services.go`
to `app_cert.go`. The startup-internal CA calls in `startup()` remain in `app.go` because
they are part of the lifecycle hook, not user-facing Wails methods.

## Consequences

### Positive

- Each `app_<domain>.go` file is small and focused. Navigating to a PHP method means opening
  `app_php.go` — no scrolling past unrelated domains.
- `NewApp()` is the single place that reveals the application's wiring. A developer reading
  only this function understands every dependency relationship.
- Adding a new service manager has a clear template: create `internal/myservice/`, wire in
  `NewApp()`, add `app_myservice.go` with thin delegation methods.
- Merge conflicts are localized. Changes to PHP methods only touch `app_php.go`; changes to
  cert methods only touch `app_cert.go`.

### Negative

- More files in the root package (`main`). The package is structurally flat — all `app_*.go`
  files are in the `main` package. Go's package model means there's no sub-package for each
  domain at this layer; the split is file-level only.
- The `App` struct definition in `app.go` must list all manager fields. A developer jumping
  directly to `app_cert.go` still needs to look at `app.go` to understand what fields exist.

### Neutral

- Go allows a package to span multiple files freely. The compiler sees all `app_*.go` files
  as one package. This is standard Go practice for large packages.
- The split does not affect Wails binding generation — Wails scans the bound struct's methods
  regardless of which file they appear in.
- A future refactor could move domains into sub-packages if the app layer grows further.
  The current file-split approach is low-risk and sufficient for the current scale.
