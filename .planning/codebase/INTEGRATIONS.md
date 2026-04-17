# External Integrations

**Analysis Date:** 2026-04-16

## APIs & External Services

**Version Distribution:**
- Node.js - Fetches available versions from https://nodejs.org/dist/index.json
  - SDK/Client: net/http (Go standard library)
  - Implementation: `internal/services/nodejs/manager.go` - ListAvailable() method

**Browser Integration:**
- Opens URLs in default browser for web admin panels
  - SDK/Client: github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
  - Used in: `app.go` for opening phpMyAdmin and pgAdmin UIs

**phpMyAdmin Downloads:**
- Downloads phpMyAdmin releases from https://files.phpmyadmin.net/phpMyAdmin/{version}/phpMyAdmin-{version}-all-languages.zip
  - Implementation: `internal/services/phpmyadmin/manager.go` - downloadURL constant
  - Installed to: `~/.lamboserver/services/phpmyadmin/`

**Node.js Distribution Downloads:**
- Downloads Node.js binaries from https://nodejs.org/dist/{version}/node-{version}-darwin-{arch}.tar.gz
  - Implementation: `internal/services/nodejs/manager.go` - Install() method
  - Installed to: `~/.lamboserver/nodejs/{version}/`

## Data Storage

**Databases:**
- MySQL 8.4.3 LTS
  - Connection: Unix socket at `~/.lamboserver/services/mysql/mysql.sock`
  - Client: github.com/go-sql-driver/mysql v1.9.1
  - Manager: `internal/services/mysql/manager.go` (realDBOpener uses database/sql)
  - Lifecycle: Managed via LaunchDaemon at `~/Library/LaunchDaemons/com.lamboserver.mysql.plist`

- PostgreSQL (latest stable)
  - Connection: Unix socket at `~/.lamboserver/services/postgres/postgres.sock`
  - Client: github.com/jackc/pgx/v5 v5.9.1 (pgx/v5/stdlib driver)
  - Manager: `internal/services/postgres/manager.go` (realDBOpener uses database/sql)
  - Lifecycle: Managed via pg_ctl and LaunchDaemon

**File Storage:**
- Local filesystem only
- All application data stored in `~/.lamboserver/` directory structure:
  - `bin/` - Managed symlinks and helper binaries
  - `php/` - PHP version installations
  - `nodejs/` - Node.js version installations
  - `nginx/` - Nginx configuration and binaries
  - `certs/` - CA and site SSL certificates
  - `logs/` - Application and service logs
  - `services/` - Service-specific data (MySQL, PostgreSQL, etc.)
  - `dnsmasq/` - dnsmasq configuration
  - `config.json` - Application configuration

**Caching:**
- None detected - application uses in-memory configuration cache with mutex-protected read/write

## Authentication & Identity

**Auth Provider:**
- Custom - No external authentication service
- Implementation: `internal/system/` - Privilege escalation via sudoers for admin operations
- Approach: One-time admin password prompt for privileged commands (port binding, service lifecycle)

## Monitoring & Observability

**Error Tracking:**
- None detected - No external error tracking service (Sentry, etc.)

**Logs:**
- Approach: File-based logging to `~/.lamboserver/logs/` directory
- Reader: `pkg/logger/reader.go` - Reads log files for UI display
- Services logged:
  - Nginx: `~/.lamboserver/nginx/logs/access.log` and `error.log`
  - PHP-FPM: Logged to service-specific files
  - dnsmasq: Logged to service-specific files
- Debug mode: Controlled by config.json `DebugMode` flag in `internal/config/store.go`

## CI/CD & Deployment

**Hosting:**
- macOS desktop application only
- No cloud deployment or hosted infrastructure
- Self-contained with no network requirements (except for downloading versions)

**CI Pipeline:**
- None detected - No automated CI/CD configuration found

## Environment Configuration

**Required env vars:**
- Not used - Application relies on filesystem paths and local configuration only
- Configuration completely file-based in `~/.lamboserver/config.json`
- No environment variables for database credentials or service configuration

**Secrets location:**
- MySQL: Socket authentication (no password file)
- PostgreSQL: Socket authentication (no password file)
- SSL certificates: `~/.lamboserver/certs/` (self-signed, trusted in macOS Keychain)
- Application secrets: None - No external API keys or credentials required
- Privilege escalation: macOS sudoers (prompted at runtime)

## Webhooks & Callbacks

**Incoming:**
- None detected - Application does not expose webhook endpoints

**Outgoing:**
- None detected - Application does not call external webhooks

## Web Admin Panels

**phpMyAdmin:**
- Purpose: MySQL database administration UI
- URL: `http://localhost:{nginx_port}/phpmyadmin` or https with SSL
- Installation: Downloaded from https://files.phpmyadmin.net
- Location: `~/.lamboserver/services/phpmyadmin/`
- Server: Served by Nginx via `internal/services/phpmyadmin/manager.go`
- Config: Generated with MySQL socket connection configuration
- Manager: `internal/services/phpmyadmin/manager.go` (phpmyadmin.NewServiceAdapter)

**pgAdmin:**
- Purpose: PostgreSQL database administration UI
- URL: `http://localhost:{nginx_port}/pgadmin` or https with SSL
- Location: `~/.lamboserver/services/pgadmin/`
- Manager: `internal/services/pgadmin/manager.go` (pgadmin.NewServiceAdapter)
- Implementation: Placeholder - configuration structure in place

**CloudBeaver:**
- Purpose: Universal database connection client (supports MySQL, PostgreSQL, etc.)
- URL: `http://localhost:{nginx_port}/cloudbeaver` or https with SSL
- Location: `~/.lamboserver/services/cloudbeaver/`
- Manager: `internal/services/cloudbeaver/manager.go`
- Config: Auto-configures connection presets pointing to local PostgreSQL instance
- Configuration: `internal/services/cloudbeaver/config.go` with connection presets

## System Integration

**Notifications:**
- Implementation: `pkg/notify/notify.go`
- macOS: Uses AppleScript via osascript (`display notification` command)
- Linux/Windows: Placeholder for notify-send and toast support
- Used for: Service state changes, installation completion, error alerts

**Browser Integration:**
- Opens phpMyAdmin, pgAdmin, CloudBeaver web UIs in default browser
- Uses: `github.com/pkg/browser` package in `app.go`

**System Tray:**
- Placeholder implementation in `internal/tray/tray.go`
- Future: macOS menu bar icon with quick access to start/stop services
- Awaiting suitable Go tray library selection (systray, fyne, etc.)

---

*Integration audit: 2026-04-16*
