# Phase 5: Remove CloudBeaver - Context

**Gathered:** 2026-04-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete removal of CloudBeaver from the codebase — Go backend, frontend UI, config constants, and service registration. Also clean up lingering pgAdmin dead code. After this phase, a codebase-wide search for "cloudbeaver" (case-insensitive) returns zero results in source files.

</domain>

<decisions>
## Implementation Decisions

### PostgreSQL Page After Removal
- **D-01:** Remove the CloudBeaver panel entirely from PgDatabasePage.tsx — no placeholder, no "coming soon" hint. The page should only show PostgreSQL service controls (start/stop/status). Phase 6 will add the pgweb panel fresh.

### Config & Port Cleanup
- **D-02:** Remove `DefaultCloudBeaverPort` (8978) and all CloudBeaver-related config constants entirely. Clean slate — Phase 6 defines pgweb's port from scratch.

### pgAdmin Dead Code
- **D-03:** Remove the `internal/services/pgadmin/` directory as part of this phase. pgAdmin was replaced in v1.0 but the directory still lingers. Clean sweep alongside CloudBeaver removal.

### Claude's Discretion
- Ordering of removal steps (backend-first vs frontend-first)
- Whether to consolidate into one plan or split into backend/frontend plans
- Handling of any test files that reference CloudBeaver or pgAdmin

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

No external specs — requirements fully captured in decisions above.

### Codebase Files to Remove/Modify
- `internal/services/cloudbeaver/` — entire package (cloudbeaver.go, service_adapter.go, config.go, installer.go)
- `internal/services/pgadmin/` — entire package (dead code from v1.0)
- `app.go` — CloudBeaver import, field, registration, URL/OpenWebAdmin special-casing
- `internal/services/service.go` — CloudBeaver comments
- `internal/config/defaults.go` — `DefaultCloudBeaverPort` constant
- `frontend/src/pages/PgDatabasePage.tsx` — CloudBeaver panel, status checks, install/start/stop/open calls
- `docs/DEVELOPMENT.md` — CloudBeaver references

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None needed — this is a pure removal phase

### Established Patterns
- ServiceAdapter pattern: CloudBeaver follows the same adapter pattern as other services. Removal means deleting the package and unregistering from service manager.
- `app.go` composition root: CloudBeaver is registered via `mgr.Register("cloudbeaver", ...)` and has special-case URL/OpenWebAdmin handling that must be removed.
- PgDatabasePage.tsx: CloudBeaver panel is a distinct section in the PostgreSQL page. Removing it leaves the core PostgreSQL service controls intact.

### Integration Points
- `app.go` NewApp() — CloudBeaver manager creation and service registration
- `app.go` GetServiceURL() — special case for "cloudbeaver" name
- `app.go` OpenWebAdmin() — special case for "cloudbeaver" browser open
- `internal/services/service.go` — comment references only
- `internal/config/defaults.go` — port constant

</code_context>

<specifics>
## Specific Ideas

No specific requirements — straightforward removal following the same pattern used to remove pgAdmin in Phase 1.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 05-remove-cloudbeaver*
*Context gathered: 2026-04-16*
