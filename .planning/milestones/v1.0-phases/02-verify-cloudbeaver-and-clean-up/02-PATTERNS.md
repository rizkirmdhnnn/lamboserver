# Phase 2: Verify CloudBeaver and Clean Up - Pattern Map

**Mapped:** 2026-04-16
**Files analyzed:** 2 (1 modified, 1 verified-only)
**Analogs found:** 2 / 2

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `docs/DEVELOPMENT.md` | documentation | n/a | `docs/DEVELOPMENT.md` itself (line edit) | exact |
| `internal/services/cloudbeaver/config.go` | config | file-I/O | `internal/services/postgres/manager.go` (writeConf/writeHba) | role-match |

**Verification targets (read-only, no edits expected):**

| File | Role | Verification Goal |
|---|---|---|
| `internal/services/cloudbeaver/cloudbeaver.go` | service | Confirm Manager is correctly wired, no pgAdmin remnants |
| `internal/services/cloudbeaver/service_adapter.go` | service adapter | Confirm implements `services.Service`, not `WebAdminService` |
| `frontend/src/pages/PgDatabasePage.tsx` | component | Confirm only CloudBeaver panel, no pgAdmin UI |
| `app.go` | composition root | Confirm CloudBeaver registered as regular service, `OpenWebAdmin` handles it |

---

## Pattern Assignments

### `docs/DEVELOPMENT.md` — Delete pgAdmin block (lines 127–130)

**Analog:** The file itself — the pgAdmin entry mirrors the phpmyadmin entry immediately above it.

**Current state** (lines 122–130):
```
│   │   ├── phpmyadmin/  # phpMyAdmin web admin
│   │   │   ├── manager.go           # Install, Uninstall, Status, URL
│   │   │   ├── service_adapter.go   # Adapts Manager to services.WebAdminService
│   │   │   └── interfaces.go        # FileSystem
│   │   │
│   │   └── pgadmin/     # pgAdmin web admin
│   │       ├── manager.go           # Status, Open
│   │       ├── service_adapter.go   # Adapts Manager to services.WebAdminService
│   │       └── interfaces.go        # Interfaces
```

**Required edit:** Replace the `└── pgadmin/` block (lines 127–130) with a CloudBeaver entry. The phpmyadmin block (lines 122–125) becomes the new closing branch (`└──`), and a new cloudbeaver entry is appended below it, OR the pgadmin block is simply replaced with cloudbeaver — matching the directory structure that already exists on disk.

**Target state** — replace lines 127–130 with:
```
│   │   └── cloudbeaver/ # CloudBeaver web database manager
│   │       ├── cloudbeaver.go       # Manager: install, start, stop, URL
│   │       ├── config.go            # Generates cloudbeaver.conf with PostgreSQL preset
│   │       └── service_adapter.go   # Adapts Manager to services.Service
```

**Note:** The `phpmyadmin` entry connector must change from `├──` to `├──` (unchanged) since `cloudbeaver` becomes the new last entry under `services/` — check the tree characters carefully when editing. The `└──` on phpmyadmin's line becomes `├──` only if cloudbeaver is appended after it; if pgadmin is directly replaced by cloudbeaver (same position), phpmyadmin stays `├──` as-is.

---

### `internal/services/cloudbeaver/config.go` — Connection Preset Verification

**This file is NOT modified.** The task is to cross-check the preset values against the PostgreSQL service configuration.

**Preset in `config.go`** (lines 17–51):
```go
"datasources": {
    "postgresql-local": {
        "provider": "postgresql",
        "driver": "postgres",
        "name": "LamboServer PostgreSQL",
        "host": "localhost",
        "port": "5432",
        "database": "postgres",
        "user": "postgres",
        "saveCredentials": true
    }
}
```

**Cross-check target — PostgreSQL `writeConf()` in `internal/services/postgres/manager.go`** (lines 156–169):
```go
conf := fmt.Sprintf(`# LamboServer PostgreSQL configuration
listen_addresses = 'localhost'
port = 5432
unix_socket_directories = '%s'
...
`, m.paths.PostgreSQLSocketDir())
```

**Cross-check target — PostgreSQL `writeHba()` in `internal/services/postgres/manager.go`** (lines 173–183):
```go
hba := `# LamboServer pg_hba.conf — trust all local connections for dev
# TYPE  DATABASE  USER  ADDRESS       METHOD
local   all       all                 trust
host    all       all   127.0.0.1/32  trust
host    all       all   ::1/128       trust
`
```

**Cross-check target — `internal/config/defaults.go`** (lines 22–24):
```go
// DefaultPostgresPort is the default port for PostgreSQL.
DefaultPostgresPort = 5432

// DefaultCloudBeaverPort is the port for the standalone CloudBeaver server.
DefaultCloudBeaverPort = 8978
```

**Verification result (pre-confirmed from reading):**
- CloudBeaver preset: `host=localhost`, `port=5432`, `user=postgres` — MATCHES
- PostgreSQL listens on `localhost:5432`, trust auth for `127.0.0.1/32` — MATCHES
- CB-02 requirement is satisfied as-is; no code change needed

---

## Shared Patterns

### pgAdmin Search Pattern (DOC-02 compliance check)

**Apply to:** Search scope — source code and `docs/` only, excluding `.planning/` and `.claude/`

The only pgAdmin reference found in scope is `docs/DEVELOPMENT.md`. Confirmed via:
```bash
grep -r "pgadmin" . --include="*.go" --include="*.ts" --include="*.tsx" --include="*.md" \
  --exclude-dir=".planning" --exclude-dir=".claude" -l
# Returns: docs/DEVELOPMENT.md only
```

### Build Verification Pattern (D-03)

**Source:** Project's `wails.json` and standard Go toolchain

After the doc edit, the planner should include these verification steps:
```bash
go build ./...
# Confirms no pgAdmin Go imports remain in compiled code

cd frontend && npm run build
# Confirms no pgAdmin TS/TSX imports remain in frontend bundle
```

### Service Registration Pattern (app.go lines 80–87)

**Source:** `/Users/rizkirmdhn/Documents/Code/lamboserver/app.go`

CloudBeaver is registered as a **regular Service** (not WebAdminService), unlike phpMyAdmin:
```go
mgr.Register("cloudbeaver", cloudbeaver.NewServiceAdapter(cloudbeaverMgr))
mgr.RegisterWebAdmin("phpmyadmin", phpmyadmin.NewServiceAdapter(phpMyAdminMgr))
```

`OpenWebAdmin` in `app.go` handles CloudBeaver via a special-case branch (lines 455–459):
```go
// CloudBeaver is a regular Service but has a browser-accessible URL.
if name == "cloudbeaver" {
    return browser.OpenURL(a.CloudBeaver.URL())
}
```

This is the correct wiring for CB-01. No change required.

### ServiceAdapter Interface Pattern

**Source:** `/Users/rizkirmdhn/Documents/Code/lamboserver/internal/services/cloudbeaver/service_adapter.go` (lines 12–23)

CloudBeaver's adapter implements `services.Service` (not `WebAdminService`). Compile-time check confirms this:
```go
var _ services.Service = (*ServiceAdapter)(nil)
```

This is intentional — CloudBeaver's URL is surfaced directly from `cloudbeaver.Manager.URL()` rather than through the WebAdmin registry.

---

## No Analog Found

None — all files in scope have close analogs or are self-referential edits.

---

## Verification Summary (Read-Only Findings)

| Check | Status | Evidence |
|---|---|---|
| CB-01: CloudBeaver sole PostgreSQL admin UI | PASS | `PgDatabasePage.tsx` has only CloudBeaver panel; no pgAdmin component |
| CB-02: Connection preset matches PostgreSQL config | PASS | Both use `localhost:5432`, user `postgres`, trust auth |
| DOC-01: `docs/DEVELOPMENT.md` pgAdmin reference | NEEDS EDIT | Line 127–130: pgAdmin block must be replaced with cloudbeaver |
| DOC-02: No other pgAdmin refs in source/docs scope | PASS | grep returns only `docs/DEVELOPMENT.md` |
| Build validity | PENDING | Planner should include `go build ./...` + `npm run build` in plan |

---

## Metadata

**Analog search scope:** `/Users/rizkirmdhn/Documents/Code/lamboserver` (excluding `.planning/`, `.claude/`)
**Files read:** 9
**Pattern extraction date:** 2026-04-16
