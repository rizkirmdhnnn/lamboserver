# External Integrations

**Analysis Date:** 2026-04-17

## APIs & External Services

**PHP Binary Downloads (herdphp.com):**
- Purpose: Download pre-compiled PHP binaries for macOS
- URL pattern: `https://download.herdphp.com/{version}/php{version}-{arch}`
- Also fetches version metadata: `https://herdphp.com/api/php-update?php={series}&arch={arch}`
- SDK/Client: Direct `net/http` calls in `internal/services/php/manager.go`
- Auth: None (public API)
- Supported series: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4, 8.5

**Node.js Downloads (nodejs.org):**
- Purpose: Download Node.js LTS binaries
- URL pattern: `https://nodejs.org/dist/{version}/node-{version}-darwin-{arch}.tar.gz`
- Version listing: `https://nodejs.org/dist/index.json`
- SDK/Client: Direct `net/http` calls in `internal/services/nodejs/manager.go`
- Auth: None (public API)

**MySQL Downloads (cdn.mysql.com):**
- Purpose: Download MySQL 8.4 LTS binary distribution
- URL pattern: `https://cdn.mysql.com/archives/mysql-8.4/mysql-{version}-macos14-{arch}.tar.gz`
- Pinned version: `8.4.3` (constant in `internal/services/mysql/manager.go`)
- SDK/Client: Direct HTTP download via `internal/binaries/downloader.go`
- Auth: None (public CDN)

**PostgreSQL Downloads (enterprisedb.com):**
- Purpose: Download PostgreSQL binary distribution
- URL pattern: `https://get.enterprisedb.com/postgresql/postgresql-{version}-1-osx-binaries.zip`
- SDK/Client: Direct HTTP download via `internal/binaries/downloader.go`
- Auth: None (public CDN)

**phpMyAdmin Downloads (phpmyadmin.net):**
- Purpose: Download phpMyAdmin web interface
- URL pattern: `https://files.phpmyadmin.net/phpMyAdmin/{version}/phpMyAdmin-{version}-all-languages.zip`
- Pinned version: `5.2.2` (constant in `internal/services/phpmyadmin/manager.go`)
- SDK/Client: Direct HTTP download
- Auth: None (public CDN)

**pgweb Downloads (GitHub Releases):**
- Purpose: Download pgweb PostgreSQL web admin binary
- URL pattern: `https://github.com/sosedoff/pgweb/releases/download/v{version}/pgweb_darwin_{arch}.zip`
- Pinned version: `0.17.0` (constant in `internal/services/pgweb/interfaces.go`)
- SDK/Client: Direct HTTP download
- Auth: None (public GitHub release)

## Data Storage

**Databases (managed by the app, not used by the app itself):**
- MySQL 8.4 LTS - Managed local development database
  - Connection: Unix socket at `~/.lamboserver/mysql/mysql.sock`
  - Client: `database/sql` + `github.com/go-sql-driver/mysql` driver (`internal/services/mysql/manager.go`)
  - DSN format: `root@unix(~/.lamboserver/mysql/mysql.sock)/`
  - Used for: Database CRUD operations (list, create, drop databases)
- PostgreSQL - Managed local development database
  - Connection: Unix socket at `~/.lamboserver/postgresql/run/`
  - Client: `database/sql` + `github.com/jackc/pgx/v5/stdlib` driver (`internal/services/postgres/manager.go`)
  - Used for: Database CRUD operations (list, create, drop databases)

**Application State:**
- `~/.lamboserver/config.json` - JSON file managed by `internal/config/store.go`
  - Contains: active PHP/Node versions, site configs, service toggles, debug mode
  - No database; purely file-based persistence

**File Storage:**
- All data under `~/.lamboserver/` (defined in `internal/system/paths.go`)
  - `bin/` - Symlinks and helper binaries
  - `php/` - Installed PHP versions
  - `nodejs/` - Installed Node.js versions
  - `nginx/` - Nginx config, sites, logs
  - `certs/` - CA and site SSL certificates
  - `logs/` - Application log files
  - `mysql/` - MySQL data dir, config, binaries
  - `postgresql/` - PostgreSQL data dir, binaries
  - `dnsmasq/` - dnsmasq configuration
  - `phpmyadmin/` - phpMyAdmin web files
  - `pgweb/` - pgweb binary

**Caching:**
- None (no cache layer)

## Authentication & Identity

**Auth Provider:**
- None - This is a local desktop application with no user accounts
- macOS admin privileges obtained via `osascript` (AppleScript) for privileged operations
  - Implementation: `internal/system/admin.go` (RunWithAdminPrivileges)
  - Privileged helper script installed at first run via `internal/system/helper.go`

## SSL/TLS Certificate Management

**Local CA:**
- Self-signed CA generated and trusted in macOS Keychain (`internal/cert/manager.go`)
- CA files: `~/.lamboserver/certs/ca.pem`, `~/.lamboserver/certs/ca-key.pem`
- Per-site SSL certs: `~/.lamboserver/certs/{domain}.pem`, `~/.lamboserver/certs/{domain}-key.pem`
- Trust: Added to macOS system keychain via `security add-trusted-cert`

## System Integrations (macOS-specific)

**launchd (Service Management):**
- System daemons in `/Library/LaunchDaemons/` (run as root): Nginx, dnsmasq, MySQL
- Implementation: `internal/system/launchd.go`
- Service labels: `com.lamboserver.nginx`, `com.lamboserver.dnsmasq`, `com.lamboserver.mysql`
- Uses `launchctl bootstrap` / `launchctl bootout` for lifecycle

**DNS Resolution (/etc/resolver):**
- Writes `/etc/resolver/test` to delegate `.test` TLD to local dnsmasq on 127.0.0.1
- Implementation: `internal/services/dnsmasq/manager.go`
- Requires admin privileges to write to `/etc/resolver/`

**Shell Integration:**
- Injects PATH export into shell rc files (~/.zshrc, ~/.bashrc)
- Adds `~/.lamboserver/bin` to PATH for CLI access to managed PHP/Node binaries
- Implementation: `internal/system/integration.go`

**macOS Keychain:**
- Used to trust the local CA certificate
- Accessed via `security` CLI tool

## Monitoring & Observability

**Error Tracking:**
- None (no external error tracking service)

**Logs:**
- Custom file-based logger: `pkg/logger/logger.go`
- Log reader for UI display: `pkg/logger/reader.go`
- Log directory: `~/.lamboserver/logs/`
- Per-service logs (Nginx access/error logs, MySQL logs, PostgreSQL logs)
- Debug mode toggleable at runtime via config

## CI/CD & Deployment

**Hosting:**
- GitHub Releases (DMG file distribution)

**CI Pipeline:**
- GitHub Actions (`.github/workflows/release.yml`)
- Triggered on `v*` tags
- Builds on `macos-latest` runner
- Outputs: `LamboServer-{version}.dmg`
- Release created via `softprops/action-gh-release@v2`

## Environment Configuration

**Required env vars:**
- None - Application has no environment variable dependencies

**Secrets location:**
- No secrets required - All downloads use public unauthenticated CDNs
- No `.env` files in the project

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

## Network Ports (Local)

**Services bind to these local ports:**
- Nginx: ports 80 (HTTP) and 443 (HTTPS) - configurable via `config.json`
- dnsmasq: port 53 (DNS) on 127.0.0.1
- MySQL: Unix socket only (`~/.lamboserver/mysql/mysql.sock`)
- PostgreSQL: Unix socket only (`~/.lamboserver/postgresql/run/`)
- pgweb: port 8081 (HTTP web UI)
- phpMyAdmin: served via Nginx on `phpmyadmin.test` domain

## Binary Download Registry

All external binary downloads are centralized in `internal/binaries/registry.go`:

| Service    | Source                          | Version Strategy     |
|------------|---------------------------------|---------------------|
| PHP        | download.herdphp.com            | Dynamic (API query) |
| Node.js    | nodejs.org                      | Dynamic (API query) |
| MySQL      | cdn.mysql.com                   | Pinned: 8.4.3       |
| PostgreSQL | get.enterprisedb.com            | User-specified       |
| phpMyAdmin | files.phpmyadmin.net            | Pinned: 5.2.2       |
| pgweb      | github.com/sosedoff/pgweb       | Pinned: 0.17.0      |
| Nginx      | Embedded in binary              | Build-time           |
| dnsmasq    | Embedded in binary              | Build-time           |

Download + extraction logic in `internal/binaries/downloader.go` with optional SHA256 checksum verification.

---

*Integration audit: 2026-04-17*
