# Codebase Concerns

**Analysis Date:** 2026-04-17

## Security Concerns

**S-01: ReadLog Path Traversal — Arbitrary File Read**
- Issue: `ReadLog(filePath, lines)` in `app.go:623` is exposed to the frontend via Wails IPC. The `filePath` parameter is passed directly from the frontend with zero validation. The backend `Reader.ReadLastN()` in `pkg/logger/reader.go:57` opens any file the user process can read.
- Files: `app.go:622-624`, `pkg/logger/reader.go:57-74`
- Impact: A crafted IPC call (e.g., from browser devtools or a malicious frontend) can read any file accessible to the user — `~/.ssh/id_rsa`, `~/.aws/credentials`, `/etc/passwd`, etc. This is the highest severity concern.
- Fix approach: Validate `filePath` is under `paths.LogsDir()` or `paths.NginxLogsDir()` before opening. Use `filepath.Rel` or `strings.HasPrefix` after `filepath.Clean` to enforce directory containment.

**S-02: Domain Name Injection in Nginx Config Templates**
- Issue: `CreateSite(domain, path, phpVersion)` at `app.go:676` passes the domain through to `sites/manager.go:88` which injects it into an Nginx `server_name` directive via Go `text/template`. While `text/template` HTML-escapes by default, Nginx config is not HTML — a domain like `_; include /etc/passwd;` could inject arbitrary Nginx directives.
- Files: `internal/sites/manager.go:26-50`, `internal/sites/manager.go:88-150`, `app.go:676-689`
- Impact: Nginx config injection could lead to serving arbitrary files or proxying to attacker-controlled backends.
- Fix approach: Validate `domain` against a strict regex (e.g., `^[a-zA-Z0-9.-]+$`) before template rendering. Reject domains containing semicolons, braces, or whitespace.

**S-03: Project Path Injection in Nginx Config**
- Issue: The `projectPath` parameter in `sites/manager.go:88` is used as the `root` directive in Nginx config. While it checks the path exists via `fs.Stat()`, it does not sanitize the path string for Nginx config injection. A path containing `;` or newlines could break the config.
- Files: `internal/sites/manager.go:88-92`, `internal/sites/manager.go:26-50`
- Impact: Malformed path strings could inject Nginx directives.
- Fix approach: Validate that `projectPath` does not contain Nginx-unsafe characters (`; { } \n`).

**S-04: Helper Script Allows Unrestricted Daemon Installation**
- Issue: The `lambo-helper` bash script at `internal/system/helper.go:107-145` accepts `install-daemon` with an arbitrary plist path. The sudoers entry grants passwordless root execution of this script. Any local process that can invoke `sudo lambo-helper install-daemon /tmp/evil.plist` can install arbitrary root LaunchDaemons.
- Files: `internal/system/helper.go:34-77`, `internal/system/helper.go:106-145`
- Impact: Local privilege escalation. Any process running as the same user can install a root-level daemon without a password prompt.
- Fix approach: Validate the plist source path is under a known temp prefix and validate the plist label starts with `com.lamboserver.`. Or embed allowed labels directly in the helper script.

**S-05: Shell Command Injection via `sh -c` with String Interpolation**
- Issue: Multiple locations construct shell commands via `fmt.Sprintf` with `sh -c` and interpolate URLs or file paths using single-quote wrapping. If any input contains a single quote, the shell command breaks and could be exploited.
- Files:
  - `internal/services/mysql/manager.go:121-123` — URL interpolation in `sh -c`
  - `internal/services/postgres/manager.go:113-115` — URL interpolation in `sh -c`
  - `internal/services/nodejs/manager.go:161-162` — URL in `sh -c` (note: `url` is NOT quoted with single quotes here)
  - `internal/services/phpmyadmin/manager.go:76-93` — URL and path interpolation in `sh -c`
  - `internal/services/pgweb/manager.go:82-88` — URL interpolation in `sh -c`
  - `internal/system/download.go:50-52` — URL interpolation in `sh -c`
  - `internal/binaries/downloader.go:88-89` — URL interpolation in `sh -c`
- Impact: While URLs are currently hardcoded constants, the pattern is fragile. If any dynamic user input ever flows into these paths, shell injection is immediate. The Node.js install at `nodejs/manager.go:162` does NOT single-quote the URL, making it directly exploitable if a version string contains shell metacharacters (mitigated by the regex check at line 143).
- Fix approach: Use `exec.Command` with explicit arguments instead of `sh -c` string concatenation. For tar piping, download to a temp file first, then extract.

**S-06: MySQL and PostgreSQL Run with No Authentication**
- Issue: MySQL uses `--initialize-insecure` (root with no password) at `mysql/manager.go:161` and connects via `root@unix(socket)/` with no password at `mysql/manager.go:268`. PostgreSQL uses `trust` auth for all local connections at `postgres/manager.go:177-182`.
- Files: `internal/services/mysql/manager.go:161,268`, `internal/services/postgres/manager.go:176-182`
- Impact: Any local process can connect to these databases with full privileges. Acceptable for local dev, but users should be warned. If ports are exposed beyond localhost (currently they're bound to 127.0.0.1), this becomes critical.
- Current mitigation: Both services bind to `127.0.0.1` only. This is a known/accepted dev trade-off (comment T-07-06 in mysql/manager.go).

**S-07: PHP-FPM Socket Permissions Set to 0777**
- Issue: The PHP-FPM config sets `listen.mode = 0777` at `internal/services/php/fpm.go:126`.
- Files: `internal/services/php/fpm.go:107-142`
- Impact: Any local user can connect to the PHP-FPM socket and execute PHP code as the current user. On a multi-user system, this is a privilege concern.
- Fix approach: Use `0660` and rely on the `listen.group = staff` for Nginx access.

**S-08: phpMyAdmin Config with Auto-Login as Root**
- Issue: The generated `config.inc.php` uses `auth_type = 'config'` with `user = 'root'` and empty password at `phpmyadmin/manager.go:130-135`.
- Files: `internal/services/phpmyadmin/manager.go:124-148`
- Impact: Anyone who can reach `phpmyadmin.test` (any local user via DNS resolution) gets full MySQL root access. The `.test` domain resolves to 127.0.0.1 so it is somewhat contained, but any browser on the machine can access it.
- Current mitigation: Nginx binds to 127.0.0.1 only. This is a standard local dev configuration.

## Performance Concerns

**P-01: Log Reader Loads Entire File into Memory**
- Issue: `Reader.ReadLastN()` at `pkg/logger/reader.go:57-74` reads the ENTIRE log file line by line into a `[]string` slice, then returns only the last N lines. For large log files (MySQL error logs, Nginx access logs can grow to hundreds of MB), this causes massive memory allocation.
- Files: `pkg/logger/reader.go:57-74`
- Impact: Memory spike proportional to log file size. Could cause OOM for large access logs.
- Fix approach: Read the file from the end using `io.SeekEnd` and scan backwards for newlines, or use `tail -n` via exec.

**P-02: Blocking HTTP Calls on UI Thread**
- Issue: `ListAvailable()` methods for PHP (`php/manager.go:95-121`) and Node.js (`nodejs/manager.go:105-133`) make synchronous HTTP requests to external APIs (herdphp.com, nodejs.org). These are called from Wails IPC which blocks the Go goroutine handling the frontend request.
- Files: `internal/services/php/manager.go:95-121,130-150`, `internal/services/nodejs/manager.go:105-133`
- Impact: If the external API is slow or down, the UI freezes. No timeout is set on the HTTP clients (uses `http.Get` which has no timeout by default).
- Fix approach: Use `http.Client{Timeout: 10 * time.Second}` instead of `http.Get`. Consider caching results.

**P-03: Download Operations Use Default HTTP Client with No Timeout**
- Issue: `system/download.go:17` and `binaries/downloader.go:46` use `http.Get()` which has no timeout. Large binary downloads (MySQL ~500MB) will hang indefinitely if the connection stalls.
- Files: `internal/system/download.go:17`, `internal/binaries/downloader.go:46`
- Impact: Hung downloads with no way to cancel or timeout. UI shows "Installing..." forever.
- Fix approach: Use `http.Client{Timeout: 300 * time.Second}` or implement progress reporting with context cancellation.

**P-04: MySQL Stop Polls for 30 Seconds Synchronously**
- Issue: `mysql/manager.go:244-249` and `postgres/manager.go:244-249` both poll with `time.Sleep(time.Second)` in a loop for up to 30 seconds, blocking the calling goroutine.
- Files: `internal/services/mysql/manager.go:244-249`, `internal/services/postgres/manager.go:244-249`
- Impact: The UI freezes for up to 30 seconds during Stop operations. The Wails IPC call does not return until polling completes.
- Fix approach: Accept a `context.Context` and use `time.NewTicker` with context cancellation, or run the poll in a background goroutine and notify the frontend via events.

## Maintainability Concerns

**M-01: app.go is a 797-line God Object**
- Issue: `app.go` (797 lines) contains all Wails-exposed methods in a single struct with 16 dependency fields. Every new feature adds more methods here. The file mixes service lifecycle, database CRUD, certificate management, shell integration, logging, and dashboard aggregation.
- Files: `app.go:1-798`
- Impact: Hard to navigate, high risk of merge conflicts, difficult to test in isolation.
- Fix approach: Split into multiple files by domain: `app_services.go`, `app_database.go`, `app_sites.go`, `app_config.go`, etc. The `App` struct can remain unified, but methods should be organized by file.

**M-02: Duplicated FileSystem/CommandRunner Interfaces Across Every Package**
- Issue: Every service package defines its own `FileSystem`, `CommandRunner`, `AdminRunner`, and `LaunchdService` interfaces with identical or near-identical signatures. Each also defines its own production implementations (`osFileSystem`, `systemCommandRunner`).
- Files: `internal/services/mysql/interfaces.go`, `internal/services/postgres/interfaces.go`, `internal/services/nginx/interfaces.go`, `internal/services/dnsmasq/interfaces.go`, `internal/services/nodejs/interfaces.go`, `internal/services/phpmyadmin/interfaces.go`, `internal/services/pgweb/interfaces.go`, `internal/services/php/interfaces.go`
- Impact: Code duplication across 8+ packages. Adding a new method to the interface requires updating every package. The production implementations are copy-pasted.
- Fix approach: Define shared interfaces in a `pkg/contracts` or `internal/platform` package. Keep package-specific extensions where needed but share the common base.

**M-03: SaveConfig Method Ignores Its Parameter**
- Issue: `App.SaveConfig(cfg config.AppConfig)` at `app.go:716` accepts an `AppConfig` parameter but calls `a.Config.Save()` which persists the in-memory config, completely ignoring the passed `cfg` value.
- Files: `app.go:716-717`
- Impact: Frontend callers believe they are persisting their config changes, but the passed config is silently discarded. This is a bug.
- Fix approach: Either use the passed `cfg` (set it in the store then save), or remove the parameter and have the frontend call specific setter methods.

**M-04: Embedded Binaries Only for arm64**
- Issue: The `go:embed embedded/darwin-arm64/*` directive at `internal/system/embed.go:11` only embeds arm64 binaries. The `ExtractBinary` function constructs the path using `runtime.GOARCH`, but on amd64 systems, the path `embedded/darwin-x86_64/` does not exist.
- Files: `internal/system/embed.go:11-40`
- Impact: The application cannot extract nginx/dnsmasq on Intel Macs. Build fails or runtime extraction fails silently. The `BinaryLocator.Find()` method falls back to system PATH, but this defeats the purpose of embedded binaries.
- Fix approach: Either embed binaries for both architectures, or remove embedded binaries for the non-supported arch and document the limitation.

**M-05: System Tray Package is a Stub**
- Issue: `internal/tray/tray.go` is an empty package with only a doc comment. The current branch is `feat/systemtray` but no implementation exists.
- Files: `internal/tray/tray.go`
- Impact: No system tray functionality despite being on the feature branch. Dead code.
- Fix approach: Implement using `fyne` or `systray` library, or remove the package until implementation begins.

**M-06: Process Package is Unused**
- Issue: The `internal/process/` package defines `ProcessManager` interface, `ProcessConfig`, launchd implementation, PID file management, and plist helpers. However, the actual codebase uses `internal/system/launchd.go` and `internal/system/helper.go` directly for process management. The `process` package appears to be a planned refactor that was never wired in.
- Files: `internal/process/runner.go`, `internal/process/launchd.go`, `internal/process/pidfile.go`, `internal/process/plist.go`
- Impact: Dead code that adds confusion. Two competing process management approaches exist.
- Fix approach: Either migrate all services to use the `process` package, or remove it.

## Platform-Specific Risks

**PL-01: macOS-Only, No Cross-Platform Path**
- Issue: The entire codebase is hardcoded for macOS: launchd, osascript admin prompts, /Library/LaunchDaemons, /etc/resolver, `security` CLI for keychain, Mach-O embedded binaries.
- Files: `internal/system/darwin.go`, `internal/system/launchd.go`, `internal/system/helper.go`, `internal/system/paths.go`
- Impact: Zero portability to Linux or Windows. This is by design but limits the user base.
- Current mitigation: Accepted trade-off for a macOS-focused tool.

**PL-02: No Code Signing or Notarization**
- Issue: Per project memory, there is no Apple Developer Account. The app uses ad-hoc signing only.
- Impact: macOS Gatekeeper will block the app on first launch. Users must right-click > Open or disable Gatekeeper. Automatic updates cannot use Apple's notarization service.
- Current mitigation: DMG distribution via GitHub releases with manual Gatekeeper bypass instructions.

**PL-03: Hardcoded Port Numbers**
- Issue: Port 80 (Nginx), 443 (Nginx SSL), 53 (dnsmasq), 3306 (MySQL), 5432 (PostgreSQL), 8081 (pgweb) are all hardcoded throughout the codebase. The config has `NginxPort` and `NginxSSLPort` fields but they are never used by the actual Nginx config generation.
- Files: `internal/config/store.go:28-29` (unused fields), `internal/services/nginx/manager.go:129` (hardcoded 80), `internal/services/mysql/manager.go:182` (hardcoded 3306), `internal/services/postgres/manager.go:161` (hardcoded 5432), `internal/services/dnsmasq/manager.go:99` (hardcoded 53)
- Impact: Port conflicts with existing services have no resolution path. Users cannot customize ports. The config fields exist but are dead code.
- Fix approach: Wire `config.NginxPort` and `config.NginxSSLPort` into the Nginx config generation. Add port fields for MySQL and PostgreSQL.

## Missing Error Handling

**E-01: pem.Encode Errors Silently Ignored**
- Issue: Four calls to `pem.Encode` in `cert/ca.go` ignore the return error with comments like "error intentionally ignored: pem.Encode to file rarely fails". While rare, disk full or permission errors would silently produce truncated certificates.
- Files: `internal/cert/ca.go:83,91,180,188`
- Impact: Corrupted or incomplete certificate files with no error reported.
- Fix approach: Check and return errors. "Rarely fails" is not "never fails."

**E-02: UserHomeDir Errors Silently Ignored**
- Issue: `os.UserHomeDir()` errors are silently discarded via `home, _ := os.UserHomeDir()` in several locations. If `$HOME` is unset, the app would use empty string paths.
- Files: `internal/system/paths.go:28`, `internal/system/paths.go:204`, `internal/system/shell.go:74-76`
- Impact: Paths like `/.lamboserver/` would be created at filesystem root, requiring root permissions and potentially corrupting the root filesystem.
- Fix approach: Return errors from `NewPaths()` or panic on startup if home dir cannot be determined.

**E-03: Frontend Errors Swallowed with console.error Only**
- Issue: All frontend pages catch errors with `console.error(e)` and do not display them to the user. The exception is `DatabasePage.tsx` which shows errors.
- Files: `frontend/src/pages/SitesPage.tsx:29,44,54`, `frontend/src/pages/LogsPage.tsx:26,35`, `frontend/src/pages/PhpPage.tsx`, `frontend/src/pages/NodePage.tsx`
- Impact: Silent failures. Site creation, log loading, and version management can fail with no user feedback.
- Fix approach: Add error state and display error messages in all pages, following the pattern in `DatabasePage.tsx`.

**E-04: Shutdown Uses pkill as Fallback Without Error Handling**
- Issue: `app.go:179-181` runs `pkill -f` commands during shutdown to forcibly kill any orphaned processes. These commands have no error handling and use broad pattern matching that could kill unrelated processes.
- Files: `app.go:179-181`
- Impact: `pkill -f "nginx.*lamboserver"` could match and kill processes from other applications that happen to have "nginx" and "lamboserver" in their command line or environment.
- Fix approach: Use PID-based termination instead of pattern-based `pkill`. Store PIDs during service start and use them for targeted cleanup.

## Dependencies at Risk

**D-01: Go 1.25.0 Specified in go.mod**
- Issue: `go.mod` specifies `go 1.25.0` which is an unreleased future Go version. As of the analysis date, Go 1.22.x is the latest stable.
- Files: `go.mod:3`
- Impact: Build will fail on any system that does not have Go 1.25+ installed. This may be intentional for bleeding-edge features, but limits contributor onboarding.
- Fix approach: Verify this is intentional. If not, downgrade to a stable Go version.

**D-02: Wails v2 — v3 is the Active Development Branch**
- Issue: The project uses `github.com/wailsapp/wails/v2 v2.12.0`. Wails v3 is in active development and v2 receives only maintenance updates.
- Files: `go.mod:10`
- Impact: Wails v2 may stop receiving security patches. Migration to v3 will require significant refactoring of bindings and build system.
- Fix approach: Plan for Wails v3 migration. Monitor the Wails v2 deprecation timeline.

**D-03: Reliance on Third-Party Download URLs Without Integrity Verification**
- Issue: PHP binaries are downloaded from `herdphp.com` with no checksum verification. MySQL and PostgreSQL downloads also skip checksum verification in the actual manager code (the `binaries/downloader.go` has checksum support but the managers use `sh -c curl | tar` directly instead of the downloader).
- Files: `internal/services/php/manager.go:182-201`, `internal/services/mysql/manager.go:108-128`, `internal/services/postgres/manager.go:93-126`, `internal/services/nodejs/manager.go:153-163`
- Impact: MITM attacks could inject malicious binaries. The `herdphp.com` domain is third-party and could be compromised.
- Fix approach: Use the `binaries/downloader.go` with SHA256 checksums for all downloads. Pin checksums for known versions.

## Test Coverage Gaps

**T-01: No Tests for app.go (Wails Binding Layer)**
- Issue: The 797-line `app.go` has no test file. All Wails-exposed methods are untested at the unit level.
- Files: `app.go` (no corresponding `app_test.go`)
- Impact: The entire API surface exposed to the frontend is untested. Input validation, error handling, and service coordination logic are at risk.
- Priority: High — this is the glue layer where bugs manifest as user-facing issues.

**T-02: No Tests for Security-Critical Helper Script Generation**
- Issue: `internal/system/helper.go` generates a bash script that runs as root via sudoers. There are no tests for the script generation or the sudoers content generation.
- Files: `internal/system/helper.go` (no `helper_test.go`)
- Impact: A bug in helper script generation could create a root-level security vulnerability.
- Priority: High

**T-03: Integration Tests Exist but Require Real System Access**
- Issue: Integration tests in `internal/integration/` test real service lifecycle scenarios but require actual macOS launchd, file system access, and network. They cannot run in CI without a macOS runner.
- Files: `internal/integration/service_lifecycle_test.go`, `internal/integration/site_creation_test.go`, `internal/integration/php_switch_test.go`, `internal/integration/startup_test.go`
- Impact: Integration tests may not be run regularly, reducing their value.
- Priority: Medium

**T-04: No Frontend Tests**
- Issue: The React frontend has no test files. There are 1818 lines of TSX with no component tests, no E2E tests, and no snapshot tests.
- Files: `frontend/src/pages/*.tsx`, `frontend/src/components/*.tsx`
- Impact: Frontend regressions are only caught manually.
- Priority: Medium

---

*Concerns audit: 2026-04-17*
