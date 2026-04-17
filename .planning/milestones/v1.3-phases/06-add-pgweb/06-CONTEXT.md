# Phase 6: Add pgweb - Context

**Gathered:** 2026-04-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Add pgweb as a standalone PostgreSQL web admin tool. Users can install, start, stop, and open pgweb from the PostgreSQL database page. pgweb is a Go binary with its own HTTP server — it runs as a daemon process, unlike phpMyAdmin which is served via Nginx+PHP-FPM.

</domain>

<decisions>
## Implementation Decisions

### UI Placement & Controls
- **D-01:** pgweb gets a dedicated "Web Admin (pgweb)" card on the PostgreSQL page, separate from the PostgreSQL service card
- **D-02:** Card position: between the PostgreSQL service card and the Create Database card (Service → Web Admin → Create DB → DB list)
- **D-03:** The Web Admin card only appears when PostgreSQL is running. When PostgreSQL is stopped or not installed, no Web Admin card is shown.
- **D-04:** Not-installed state: compact card with "Not installed" text and [Install] button — no icon or lengthy description
- **D-05:** Stopped state: status dot (stopped) + [Start] button
- **D-06:** Running state: status dot (running) + port display + [Stop] and [Open] buttons

### Lifecycle Coupling
- **D-07:** pgweb does NOT auto-start when PostgreSQL starts. User manually starts pgweb when they need the web UI.
- **D-08:** pgweb IS auto-stopped when PostgreSQL stops. pgweb can't function without PostgreSQL — clean shutdown prevents orphan processes.
- **D-09:** No state persistence across app restarts. pgweb always starts in stopped state. No surprise processes at launch.

### Installation Experience
- **D-10:** pgweb has its own [Install] button in the Web Admin card. It is NOT bundled with PostgreSQL installation.
- **D-11:** Install progress uses step labels matching PostgreSQL's pattern: "Downloading..." → "✓ Installed" → card updates to stopped state.
- **D-12:** pgweb binary is downloaded from GitHub releases to ~/.lamboserver/pgweb/

### Browser Open Behavior
- **D-13:** pgweb launches pre-configured to auto-connect to the local PostgreSQL via socket (trust auth). User sees databases immediately — no connection form.
- **D-14:** pgweb binds to localhost only (127.0.0.1). Not accessible from other machines on the network.
- **D-15:** Default port: 8081 (pgweb default). Configurable in the future if needed.

### Claude's Discretion
- Service architecture: how pgweb fits into the existing Service/WebAdminService interface hierarchy (pgweb is a daemon unlike phpMyAdmin — may need hybrid approach or new interface)
- pgweb process management: how to track the running pgweb process (PID file, os/exec process handle, etc.)
- Error handling: what to show if pgweb download fails, if port 8081 is in use, etc.
- pgweb version selection: which version to pin, how to detect platform (darwin-amd64 vs darwin-arm64)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Service Architecture
- `internal/services/service.go` — Defines Service, VersionedService, WebAdminService interfaces. pgweb needs Start/Stop (like Service) AND URL/IsInstalled (like WebAdminService)
- `internal/services/phpmyadmin/manager.go` — Reference implementation for WebAdminService (Install, URL, IsInstalled, Uninstall)
- `internal/services/phpmyadmin/service_adapter.go` — Adapter pattern for WebAdminService interface
- `internal/services/manager.go` — Service registry: Register(), RegisterWebAdmin(), Get(), GetWebAdmin()

### App Integration
- `app.go` — Composition root: service creation, registration, and public IPC methods (GetWebAdminURL, OpenWebAdmin, InstallWebAdmin)
- `internal/system/paths.go` — Single source of truth for all ~/.lamboserver/ paths. Must add PgwebDir() here.
- `internal/config/defaults.go` — Default config values. May need pgweb port constant.

### Frontend
- `frontend/src/pages/PgDatabasePage.tsx` — PostgreSQL page where the Web Admin card will be added
- `frontend/src/pages/DatabasePage.tsx` — MySQL database page (reference for how phpMyAdmin panel might be structured)
- `frontend/src/components/Toast.tsx` — Toast error component used for install error display

### Existing Patterns
- `internal/services/postgres/manager.go` — PostgreSQL manager (socket path, trust auth config)
- `internal/services/postgres/service_adapter.go` — PostgreSQL service adapter

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `system.Paths` struct — add PgwebDir() method for ~/.lamboserver/pgweb/ path
- `system.RunCommand()` — for executing pgweb binary
- `pkg/browser` — already imported in app.go for opening URLs in default browser
- Toast component — for displaying install errors with summary/detail
- Install step pattern from PgDatabasePage.tsx — reusable for pgweb install progress

### Established Patterns
- ServiceAdapter pattern: Manager implements business logic, ServiceAdapter wraps it to implement unified interface
- WebAdminService interface: Install(version), URL(), IsInstalled(), Version() — but lacks Start/Stop/Status
- Service registration in app.go NewApp(): create manager → create adapter → mgr.Register/RegisterWebAdmin
- Frontend service control: GetAllStatuses() → status map → conditional UI rendering

### Integration Points
- `app.go` NewApp() — create pgweb manager and register it
- `app.go` OpenWebAdmin() — already has generic web admin open logic via browser.OpenURL
- `app.go` shutdown() — must stop pgweb if running (auto-stop on app close)
- `app.go` StopService("postgresql") — must also stop pgweb (D-08: auto-stop coupling)
- PgDatabasePage.tsx — add Web Admin card section between service controls and database CRUD

</code_context>

<specifics>
## Specific Ideas

- pgweb connection flags: `--host /tmp --user postgres --db postgres` (matches existing PostgreSQL trust auth via Unix socket)
- pgweb bind flags: `--bind 127.0.0.1 --listen 8081`
- Card visual consistency: Web Admin card should use the same status-dot + service-row pattern used by the PostgreSQL service card

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 06-add-pgweb*
*Context gathered: 2026-04-16*
