# Codebase Concerns

**Analysis Date:** 2026-04-16

## Tech Debt

**Monolithic app.go Composition Root:**
- Issue: `app.go` (773 lines) contains all dependency injection and wiring in a single `NewApp()` function. No factory patterns or composition helpers separate concerns. Adding new services requires modifying this central file.
- Files: `app.go` lines 61-114
- Impact: Difficulty scaling to new services; high merge conflict risk in team environments; tight coupling between service initialization and App struct fields
- Fix approach: Extract service factories into separate `pkg/services/` packages (e.g., `php_factory.go`, `nginx_factory.go`) that return configured managers. Use a builder pattern or service container (Wire, fx) to reduce manual wiring.

**Shell Integration Path Mutation Without Validation:**
- Issue: `internal/system/shell.go` modifies `~/.zshrc` and `~/.bashrc` with `appendToFile()` but doesn't validate the shell marker before writing. Could result in duplicate PATH entries if called multiple times or if previous writes were corrupted.
- Files: `internal/system/shell.go` lines 46-82, 39-44
- Impact: Shell RC files accumulate duplicate PATH entries, bloating shell startup. Users may have to manually clean files.
- Fix approach: Before appending, always check for marker existence with `fileContains(rc, shellMarker)`. Implement idempotent write that replaces existing block if found.

**Unvalidated Shell Command Injection Risk in Downloads:**
- Issue: `internal/system/download.go` and `internal/binaries/downloader.go` use `curl` and `tar` via shell command execution: `fmt.Sprintf("curl -sfL '%s' | tar xz %s -C '%s'", url, strip, destDir)` (line 51 in download.go). While URL is from hardcoded constants in most cases, user-provided paths in `destDir` could escape quotes.
- Files: `internal/system/download.go` line 51, `internal/binaries/downloader.go`, `internal/services/phpmyadmin/manager.go` line 90
- Impact: Potential command injection if `destDir` contains shell metacharacters; risk grows if path generation becomes dynamic.
- Fix approach: Replace shell invocation with Go's `archive/tar` and `io.Copy()` libraries. Avoid subprocess for unpacking entirely.

**No Timeout on HTTP Downloads:**
- Issue: `internal/system/download.go` line 17 and `internal/services/php/manager.go` line 135 use bare `http.Get()` without timeout. Large file downloads (PHP/Node.js tarballs, CloudBeaver) can hang indefinitely on network stalls.
- Files: `internal/system/download.go`, `internal/services/nodejs/manager.go` line 106, `internal/services/php/manager.go` line 135
- Impact: Application appears frozen during slow downloads; no user feedback; critical for first-run experience where users expect responsiveness.
- Fix approach: Create a `DownloadWithTimeout(url, dest, timeout time.Duration)` helper wrapping `http.Get()` with context cancellation. Default 5-minute timeout. Expose to frontend for progress UI.

**Database Connection Not Closed After Use:**
- Issue: `internal/services/postgres/manager.go` and `internal/services/mysql/manager.go` open database connections via `m.dbOpen.Open()` for operations like `ListDatabases()` and `CreateDatabase()` but do not explicitly `defer db.Close()` in all code paths. Visible in `postgres/manager.go` after line 260+ in ListDatabases/DropDatabase methods (not shown in excerpt but pattern is consistent).
- Files: `internal/services/postgres/manager.go`, `internal/services/mysql/manager.go`
- Impact: Database connection leaks accumulate over session. After many list/create operations, connection pool exhaustion can cause "too many connections" errors, blocking subsequent operations until app restart.
- Fix approach: Wrap all database operations in a helper `func (m *Manager) withDB(ctx context.Context, fn func(*sql.DB) error) error` that ensures `defer db.Close()`. Enforce this pattern via code review.

**Hardcoded Service Versions Without Update Mechanism:**
- Issue: Service versions are pinned as constants: MySQL 8.4.3 (`internal/services/mysql/manager.go` line 19), CloudBeaver 24.1.0 (`internal/services/cloudbeaver/installer.go` line 12), phpMyAdmin 5.2.2 (`internal/services/phpmyadmin/manager.go` line 15). No in-app mechanism to check for or prompt updates.
- Files: `internal/services/mysql/manager.go` line 19, `internal/services/cloudbeaver/installer.go` line 12, `internal/services/phpmyadmin/manager.go` line 15
- Impact: Security patches and bug fixes require code commits and app rebuilds. Users running older versions indefinitely; no security update visibility.
- Fix approach: Move versions to `config.json` with a separate `AvailableVersions` field. Add a background check endpoint to fetch latest available versions from GitHub/CDN. Show update prompts in UI when newer versions exist.

## Known Bugs

**pgAdmin 4 Pre-Configuration Not Used:**
- Symptoms: pgAdmin 4 opens but users must manually import the pre-configured `servers.json` via Tools > Import/Export. The file is generated but pgAdmin desktop mode does not auto-read it on startup.
- Files: `internal/services/pgadmin/manager.go` lines 58-73
- Trigger: User clicks "Open pgAdmin" button in UI
- Workaround: Manual import via pgAdmin UI menu. File is written to `~/Library/Application Support/pgAdmin/servers.json` for user reference.

**Port Availability Check Uses 127.0.0.1 Only:**
- Symptoms: `IsPortAvailable()` in `internal/process/port.go` only checks `127.0.0.1`, not `0.0.0.0`. A process bound to `0.0.0.0:8978` (CloudBeaver port) will not be detected as in-use, allowing Start to attempt launching CloudBeaver and fail with "port in use" error from the process itself.
- Files: `internal/process/port.go` line 10
- Trigger: User tries to start CloudBeaver when port 8978 is already bound to all interfaces
- Workaround: Kill the offending process manually and retry
- Fix approach: Check both `127.0.0.1` and `0.0.0.0` loopback interfaces

**Shutdown Doesn't Wait for Process Termination:**
- Symptoms: `app.go` shutdown (lines 177-188) calls `Stop()` on services but immediately issues `pkill` commands without waiting for graceful shutdown. LaunchDaemon/Agent plist uninstalls may race with process termination, leaving stale plist files.
- Files: `app.go` lines 177-188
- Trigger: User closes LamboServer window during active service operation
- Workaround: None; stale plists require manual `launchctl bootout` cleanup
- Fix approach: Add `time.Sleep(500*time.Millisecond)` after calling Stop() and before pkill. Better: implement graceful shutdown handshake with 5-second timeout before forceful pkill.

**CloudBeaver Install Assumes Linux Tarball Format:**
- Symptoms: `downloadURL()` in `internal/services/cloudbeaver/installer.go` constructs URLs with `cloudbeaver-ce-{version}-{arch}-linux.tar.gz` (line 54), but runtime.GOARCH returns "amd64"/"arm64" on macOS, not the "x86_64"/"aarch64" labels CloudBeaver uses. URL construction works but assumes Linux tarball structure matches macOS.
- Files: `internal/services/cloudbeaver/installer.go` lines 47-57
- Trigger: User installs CloudBeaver; the downloaded tarball is labeled "linux" but is actually macOS-compatible
- Workaround: Download works because CloudBeaver releases use same tarball for both platforms
- Fix approach: Document that CloudBeaver CE releases are platform-agnostic; clarify URL comment. Test extraction on both arm64 and amd64 macs in CI.

## Security Considerations

**Shell Marker Injection in RC Files:**
- Risk: The shell marker `# Added by LamboServer` (line 12 in shell.go) is used to detect previous installs. User could manually add this exact string to `~/.zshrc`, causing Install() to skip real installation. Subsequent PATH setup would be incomplete.
- Files: `internal/system/shell.go` line 12
- Current mitigation: Marker is unique and unlikely to appear in user files
- Recommendations: Use a UUID-based marker (e.g., `# Added by LamboServer-{uuid}`) to reduce collision risk. Add validation that the following block contains expected PATH export, not just the marker.

**Privilege Escalation Via Hardcoded Helper Script Path:**
- Risk: `internal/system/helper.go` installs a helper script to `/Library/Application Support/lamboserver/` and registers sudoers entry granting passwordless execution. The sudoers line grants access to the script by exact path. If the path is guessable or modifiable by user, privilege escalation possible.
- Files: `internal/system/helper.go` (not fully shown but visible in system architecture)
- Current mitigation: Path is under `/Library/Application Support/` (root-writable only); sudoers entry is read-only after install
- Recommendations: Verify script is owned by root:wheel and mode 0755 on install. In uninstall, remove sudoers entry before deleting script to prevent privilege window.

**No HTTPS Verification for Binary Downloads:**
- Risk: `http.Get()` in download.go uses default TLS verification but no certificate pinning or checksum validation. A MITM attacker on the network (or DNS compromise) can serve malicious binaries for PHP/Node/MySQL/etc.
- Files: `internal/system/download.go` line 17, `internal/services/php/manager.go` line 135
- Current mitigation: Downloads use HTTPS (https:// URLs); TLS handshake validates certificate chain
- Recommendations: Add SHA-256 checksum validation for downloaded tarballs. Store checksums in `go.sum`-like manifest or fetch from GitHub releases metadata. Validate before extraction.

**Database Credentials Hardcoded for Local Operations:**
- Risk: MySQL and PostgreSQL use hardcoded default credentials (root/empty for MySQL, postgres/empty for PostgreSQL) suitable only for local socket connections. If user exposes services to network (custom config), default credentials become a vector.
- Files: `internal/services/mysql/manager.go`, `internal/services/postgres/manager.go`
- Current mitigation: Services bind to Unix sockets by default (not TCP), limiting network exposure. App is single-user desktop tool
- Recommendations: Document that services should never be exposed publicly. Add validation to prevent binding to 0.0.0.0. Consider auto-generating random passwords on install and storing in config.json (encrypted).

## Performance Bottlenecks

**Synchronous Service Status Polling:**
- Problem: `app.go` GetDashboardStatus() (line 622) calls Status() on all services sequentially without parallelization. Each launchd check is a subprocess call. With 8+ services, dashboard load is serial and slow.
- Files: `app.go` lines 622-641
- Cause: Simple sequential implementation; no goroutine concurrency
- Improvement path: Launch `Status()` calls in parallel goroutines with a timeout; use `sync.WaitGroup` to wait for all. Cap response time at 1 second (return partial results if some checks timeout).

**Blocking Process Listing on Dashboard Refresh:**
- Problem: `PhpFpm.Status()` and other service status checks likely spawn subprocess calls (`ps`, `launchctl`). Frontend may poll GetDashboardStatus() frequently (e.g., every 2 seconds), causing CPU spikes.
- Files: `app.go` line 622, services using subprocess-based status checks
- Cause: No caching of status results; each check is a fresh subprocess call
- Improvement path: Cache status results with a 500ms TTL. Invalidate on explicit service control actions (Start/Stop/Restart). Return cached result if check is too frequent.

**ListInstalled PHP Versions Scans Filesystem Sequentially:**
- Problem: `internal/services/php/manager.go` ListInstalled() detects PHP from system PATH and manual installations. Each detection spawns a subprocess to query `php --version`. No parallelization across directories.
- Files: `internal/services/php/manager.go` line 59 and detectAll() implementation (not shown)
- Cause: Sequential file system walks and subprocess calls
- Improvement path: Parallelize detection across known PHP directories using goroutines. Cache results for 30 seconds. Invalidate on install/uninstall actions.

**Nginx Config Regeneration on Every Site Change:**
- Problem: CreateSite() and DeleteSite() call `Nginx.EnsureConfig()` which regenerates the entire master nginx.conf (not just the per-site file). With many sites, regeneration involves file I/O and regex parsing on every change.
- Files: `app.go` lines 652-681, `internal/services/nginx/manager.go` line 116
- Cause: Master config includes all site configs via `include /path/sites/*.conf;`. Any change triggers full regen
- Improvement path: Decouple master config generation from per-site vhost generation. Master config is generated once at startup. Per-site changes only touch site-specific files. Master config reload is already hot-reloadable without process restart.

## Fragile Areas

**PHP Version Switching Assumes Shell Symlinks Persist:**
- Files: `internal/services/php/service_adapter.go` lines 66-72, `internal/system/shell.go` LinkPhpVersion()
- Why fragile: SwitchVersion() updates config and restarts FPM but relies on shell integration (symlinks in ~/.lamboserver/bin/) being present. If user uninstalls shell integration or manually deletes symlinks, shell command `php` will fail even though FPM is running.
- Safe modification: In SwitchVersion(), after config update, verify symlinks exist or recreate them. Add integration test that exercises uninstall-reinstall-switch workflow.
- Test coverage: No explicit test for symlink persistence across version switches. `internal/integration/` tests exist but may not cover this edge case.

**Stale PID Files Cause Zombie Process Detection:**
- Files: `internal/process/pidfile.go` lines 94-114, CloudBeaver's IsRunning()
- Why fragile: IsRunning() checks a PID file but doesn't validate the process is actually CloudBeaver (could be any process with that PID). If PID is reused by another process (common after long uptime), Stop() will terminate the wrong process.
- Safe modification: Extend PID check to verify process name matches expected binary. Read `/proc/{pid}/comm` (Linux) or `ps -o comm=` (macOS) to confirm.
- Test coverage: No test for PID reuse scenario. Integration tests assume PID files are managed correctly without adversarial scenarios.

**Database Initialization Doesn't Check for Existing Data:**
- Files: `internal/services/mysql/manager.go` InitDataDir(), `internal/services/postgres/manager.go` InitDataDir() (not shown in excerpt)
- Why fragile: InitDataDir() creates data directories and runs `mysqld --initialize-insecure` / `initdb`. If called twice on an existing data directory, behavior is undefined (may skip, may fail, may corrupt).
- Safe modification: Check if `mysql_data/mysql` directory exists (MySQL marker database) or `postgresql_data/PG_VERSION` file exists (Postgres marker). Return error if already initialized. Expose separate Reinitialize() method if user wants to reset.
- Test coverage: Tests likely mock the filesystem and command runner, not testing the actual `--initialize-insecure` logic against a real MySQL binary.

**No Validation of Site Domain Before Nginx Vhost Generation:**
- Files: `app.go` CreateSite() line 652, `internal/sites/manager.go` Link()
- Why fragile: Domain validation for Nginx vhost syntax is not enforced. User can create domains like `my site.test` (space) or `my@domain.test` (invalid chars), leading to invalid Nginx config that won't reload.
- Safe modification: Add domain validation regex: `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*\.test$`. Reject invalid domains in CreateSite() with user-facing error before vhost generation.
- Test coverage: No test for invalid domain names. `site_creation_test.go` tests valid domains but not edge cases.

## Test Coverage Gaps

**Web Admin Service Installation Not Tested:**
- What's not tested: pgAdmin 4 detection, URL construction, and launch behavior in `internal/services/pgadmin/`. CloudBeaver installation and lifecycle in `internal/services/cloudbeaver/` are also minimally tested.
- Files: `internal/services/pgadmin/manager.go`, `internal/services/cloudbeaver/cloudbeaver.go`
- Risk: Bugs in pgAdmin/CloudBeaver launch paths go undetected. User opens app, server fails silently, user frustrated.
- Priority: Medium (pgAdmin/CloudBeaver are optional features, not core)

**Privilege Helper Installation and Sudoers Entry Not Validated:**
- What's not tested: Helper script extraction, executable permission setting, sudoers entry syntax, and validation that helper can actually be called without password.
- Files: `internal/system/helper.go`, `internal/system/launchd.go` (privilege parts)
- Risk: Helper installation could fail silently, leaving app unable to perform privileged operations (nginx reload, database init as root). First-run experience broken without error message.
- Priority: High (critical for app startup)

**Configuration File Corruption Recovery:**
- What's not tested: What happens if config.json is truncated, malformed JSON, or has missing required fields. Store.Load() may panic or leave app in inconsistent state.
- Files: `internal/config/store.go` line 61 json.Unmarshal()
- Risk: A single corrupted config file can render the app non-functional. No recovery mechanism.
- Priority: High (single point of failure)

**Shell RC File Editing Edge Cases:**
- What's not tested: File permission errors, disk full during write, concurrent edits to ~/.zshrc by user while installation runs. appendToFile() may fail partially.
- Files: `internal/system/shell.go` appendToFile()
- Risk: Partial shell integration (PATH not exported) causes confusing "command not found" errors.
- Priority: Medium (edge case, but affects user experience on first run)

**Concurrent Service Lifecycle Operations:**
- What's not tested: What happens if user clicks Stop Nginx while Stop is already in progress, or clicks Install MySQL while a prior install is running. No mutual exclusion.
- Files: `app.go` StartService(), StopService(), InstallVersion() (generic dispatcher)
- Risk: Concurrent operations may corrupt state, leave processes orphaned, or cause undefined behavior.
- Priority: Medium (requires specific user timing to trigger, but reproducible in tests)

---

*Concerns audit: 2026-04-16*
