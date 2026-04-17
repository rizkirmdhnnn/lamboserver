# Phase 1: Remove pgAdmin - Context

**Gathered:** 2026-04-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Delete all pgAdmin code from the codebase so the Go backend compiles cleanly without it. This is a pure removal phase — no new features, no replacement wiring. CloudBeaver already exists as the replacement (Phase 2 verifies it).

</domain>

<decisions>
## Implementation Decisions

### Wails Binding Cleanup
- **D-01:** Regenerate Wails bindings using `wails generate module` after removing Go methods. Do NOT manually edit App.d.ts or App.js — let Wails rebuild them automatically to guarantee no stale pgAdmin references remain.

### Post-Deletion Verification
- **D-02:** Run both `go build ./...` AND `go test ./...` after all deletions. Full test suite required to catch transitive breakage, not just compilation.

### Claude's Discretion
- Deletion order (which files to remove first) is up to Claude — optimize for clean intermediate states where the build passes at each step if possible.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — REM-01 through REM-06 define every deletion target for this phase

### Codebase Maps
- `.planning/codebase/STRUCTURE.md` — Directory layout showing pgAdmin package location and service registration pattern
- `.planning/codebase/CONVENTIONS.md` — Service adapter pattern, naming conventions, and how services register in app.go

</canonical_refs>

<code_context>
## Existing Code Insights

### Files to Delete
- `internal/services/pgadmin/` — Entire directory (manager.go, service_adapter.go, interfaces.go, manager_test.go)

### Files to Modify
- `app.go` — Remove pgAdmin import, field, initialization, `OpenPgAdmin()` method, and WebAdmin registration
- `internal/system/paths.go` — Remove pgAdmin path helpers
- `internal/system/paths_test.go` — Remove pgAdmin path tests

### Auto-Generated Files to Regenerate
- `frontend/wailsjs/go/main/App.d.ts` — Contains pgAdmin binding types (will be regenerated)
- `frontend/wailsjs/go/main/App.js` — Contains pgAdmin binding functions (will be regenerated)

### Established Patterns
- Services register via `mgr.RegisterWebAdmin("pgadmin", adapter)` in app.go — removal means deleting this registration call
- Path helpers follow `PgAdminXxx()` naming in paths.go — grep and remove all matches
- WebAdminService interface pattern: pgadmin implements the same interface as phpmyadmin and cloudbeaver

### Integration Points
- Service registry in `internal/services/manager.go` — no changes needed here, just stop registering pgAdmin
- Frontend pages may reference pgAdmin bindings — Wails regeneration handles this

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches. The phase is mechanical deletion guided by the exhaustive requirements list.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-remove-pgadmin*
*Context gathered: 2026-04-16*
