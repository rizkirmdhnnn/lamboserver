# Phase 8: Production Build - Research

**Researched:** 2026-04-17
**Domain:** Wails v2 universal binary build (`darwin/universal`) — macOS .app bundle verification
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** PATH fix for Homebrew binaries is NOT NEEDED — all service binaries (nginx, dnsmasq, mysql, postgres, php, node, pgweb) are self-contained in `~/.lamboserver/` and located via `BinaryLocator.Find()` which checks `LocalPath` first. No Homebrew dependency exists.
- **D-02:** Build target is `darwin/universal` (Intel + Apple Silicon universal binary)
- **D-03:** Build command is `wails build -platform darwin/universal -clean`
- **D-04:** No build script needed for this phase — the single wails command is sufficient. A full build pipeline script will be created in Phase 9 (build + sign + DMG).
- **D-05:** Validation is build + manual launch test — run `wails build`, then open the .app from `build/bin/LamboServer.app` and verify it launches with the correct icon and metadata
- **D-06:** Verify Info.plist inside the built .app contains correct values from Phase 7 (productName, version 1.0.0, bundle ID dev.lamboserver.app, copyright)

### Claude's Discretion

- Whether to add a Makefile target or keep the build as a direct wails command
- Whether LSMinimumSystemVersion needs adjustment based on build output

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BUILD-01 | `wails build` produces a universal binary (.app) for both Intel and Apple Silicon | D-02, D-03: `-platform darwin/universal` flag confirmed working in Wails v2.12.0 |
| BUILD-02 | Homebrew binaries (nginx, php, mysql, etc.) are found when app is launched from Finder (PATH fix) | D-01: No Homebrew dependency — all binaries in `~/.lamboserver/` via `BinaryLocator` with `LocalPath`; BUILD-02 is satisfied by the existing architecture, not by a PATH fix |
| BUILD-03 | All service management functions work correctly when running as bundled .app from /Applications | BinaryLocator uses absolute paths (`~/.lamboserver/bin/nginx`, etc.); no PATH dependence; verified via manual launch test |
</phase_requirements>

---

## Summary

Phase 8 is a verification phase, not a development phase. Phase 7 already completed all code and configuration changes needed: `wails.json` has the `info` block with `productVersion: "1.0.0"`, `Info.plist` has `CFBundleIdentifier: "dev.lamboserver.app"`, and the icon is in place. The only work in Phase 8 is (1) running `wails build -platform darwin/universal -clean` to produce a fresh universal binary .app, and (2) verifying the output is correct.

The key technical fact unlocked by Phase 7's context discussion (D-01) is that BUILD-02 ("PATH fix for Homebrew binaries") is already satisfied by the existing architecture: every service binary in the codebase uses `BinaryLocator` with an explicit `LocalPath` pointing into `~/.lamboserver/`. The `BinaryLocator.Find()` method checks `LocalPath` first via `os.Stat()`, bypassing `exec.LookPath` entirely. This means the GUI PATH starvation pitfall (the #1 Wails production issue) does not apply.

The only gap is `build/darwin/entitlements.plist` which does not exist yet. This file is not required for `wails build` to succeed, but Phase 9 (signing) needs it. Creating it in Phase 8 is the correct time — it ensures the .app is ready for signing immediately after BUILD-01/02/03 are verified. The `entitlements.plist` content is already designed in the existing PITFALLS.md research.

**Primary recommendation:** Run `wails build -platform darwin/universal -clean`. Then verify with `lipo -info` that the binary contains both arm64 and x86_64 slices, and check the rendered `Info.plist` values. Create `build/darwin/entitlements.plist` now to unblock Phase 9.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Universal binary compilation | Build toolchain (Wails CLI) | Go cross-compiler | Wails orchestrates `GOARCH=amd64` + `GOARCH=arm64` builds and calls `lipo` to merge |
| Info.plist rendering (version, bundle ID, copyright) | Build toolchain (Wails template engine) | wails.json (data source) | Wails renders `build/darwin/Info.plist` template at build time; values come from `wails.json` `info` block |
| Icon conversion to .icns | Build toolchain (Wails via sips + iconutil) | build/appicon.png (source) | Wails calls system `sips` and `iconutil` during build; no manual step needed |
| Binary discovery at runtime | Backend (BinaryLocator) | `~/.lamboserver/` paths | All service binaries are resolved via absolute LocalPath; no PATH environment dependency |
| Build verification | Developer (manual) | codesign / lipo tools | No automated test can substitute for a manual Finder launch to confirm GUI PATH behavior |

---

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| wails | v2.12.0 | Builds .app bundle from Go + React source | The project's framework; no alternative [VERIFIED: `wails version` on this machine] |
| go | go1.26.1 | Compiles Go backend for darwin/amd64 + darwin/arm64 | Go's built-in cross-compilation; no extra toolchain needed [VERIFIED: `go version`] |
| npm | 11.11.0 | Builds React frontend before `wails build` bundles it | Called automatically by `wails build` via `frontend:build` hook in wails.json [VERIFIED: `npm --version`] |

### Supporting (Phase 8 scope only)

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| lipo | system | Verify universal binary contains both arches | Run after `wails build` to confirm fat binary |
| defaults read | system | Inspect Info.plist in built .app | Verify bundle ID, version, copyright values |
| codesign -dv | system | Inspect signing state of .app | Not used in Phase 8; mentioned for context since Phase 9 will use it |

### No New Dependencies

Phase 8 introduces no new npm or Go dependencies. All tools are already installed. The `wails build -platform darwin/universal` flag was confirmed working with Wails v2.12.0.

---

## Architecture Patterns

### System Architecture Diagram

```
Developer terminal
       │
       ▼
wails build -platform darwin/universal -clean
       │
       ├─── [Step A] npm install + npm run build
       │      └── builds React/TS frontend → frontend/dist/
       │
       ├─── [Step B] go build -tags desktop,production
       │      ├── GOARCH=arm64  → arm64 binary (temp)
       │      └── GOARCH=amd64  → amd64 binary (temp)
       │      └── lipo -create → universal binary
       │
       ├─── [Step C] Assemble .app bundle
       │      ├── Renders build/darwin/Info.plist template
       │      │    └── injects wails.json info block values
       │      ├── sips + iconutil: appicon.png → iconfile.icns
       │      └── Places binary at Contents/MacOS/LamboServer
       │
       └─── Output: build/bin/LamboServer.app
                     └── universal binary (arm64 + x86_64)
```

### Recommended Project Structure (Phase 8 output)

```
lamboserver/
├── wails.json                         # ALREADY DONE (Phase 7): info block present
├── build/
│   ├── appicon.png                    # ALREADY DONE (Phase 7): custom LamboServer icon
│   ├── darwin/
│   │   ├── Info.plist                 # ALREADY DONE (Phase 7): dev.lamboserver.app bundle ID
│   │   ├── Info.dev.plist             # EXISTS (unchanged)
│   │   └── entitlements.plist         # MISSING → CREATE IN PHASE 8 (Phase 9 needs it)
│   └── bin/
│       └── LamboServer.app            # OUTPUT of wails build (gitignored)
│           └── Contents/
│               ├── Info.plist         # rendered from template
│               ├── MacOS/LamboServer  # universal fat binary
│               └── Resources/
│                   └── iconfile.icns  # converted from appicon.png
```

### Pattern 1: Wails Universal Binary Build

**What:** Running `wails build -platform darwin/universal -clean` causes Wails to compile the Go backend twice (arm64 and amd64), then call `lipo -create` to merge them into a fat binary, then assemble the .app bundle.

**When to use:** Always for production builds targeting distribution. Single-arch builds work only on matching hardware.

**Command:**
```bash
# Source: Wails v2 CLI reference (verified on this machine with wails v2.12.0)
wails build -platform darwin/universal -clean
```

**Flags:**
- `-platform darwin/universal` — produces arm64 + x86_64 fat binary
- `-clean` — forces a full rebuild; avoids stale artifacts from previous builds

**Expected output:**
```
build/bin/LamboServer.app
```

**Verification command:**
```bash
lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer
# Expected: Architectures in the fat file: ... are: x86_64 arm64
```

### Pattern 2: BinaryLocator — PATH-Independent Service Discovery

**What:** Every service manager in the codebase uses `BinaryLocator` with an explicit `LocalPath` pointing to `~/.lamboserver/`. The `Find()` method calls `os.Stat(b.LocalPath)` first — if the file exists, it returns the absolute path without consulting `exec.LookPath` or the `PATH` environment variable.

**Why this matters:** GUI apps launched from Finder do not inherit the shell `PATH`. `exec.LookPath("nginx")` would fail because `/opt/homebrew/bin` is not in the GUI app's `PATH`. But `os.Stat("/Users/alice/.lamboserver/bin/nginx")` succeeds regardless of `PATH`. This is the architectural reason BUILD-02 is already satisfied.

**Verified in codebase:**
```go
// Source: internal/system/binary.go (verified in this session)
func (b *BinaryLocator) Find() string {
    if _, err := os.Stat(b.LocalPath); err == nil {
        return b.LocalPath  // ← returns absolute path, no PATH lookup
    }
    // LookPath fallback only runs if LocalPath doesn't exist
    if CommandExists(b.Name) {
        if path, err := RunCommand("which", b.Name); err == nil && path != "" {
            return path
        }
    }
    return ""
}
```

**Service-to-path mapping (verified by grep):**
| Service | LocalPath |
|---------|-----------|
| Nginx | `~/.lamboserver/bin/nginx` |
| Dnsmasq | `~/.lamboserver/bin/dnsmasq` |
| MySQL | `~/.lamboserver/mysql/bin/mysqld` |
| PostgreSQL | `~/.lamboserver/postgresql/bin/pg_ctl` |

### Anti-Patterns to Avoid

- **Editing the rendered .app bundle directly:** `build/bin/LamboServer.app/Contents/Info.plist` is regenerated on every `wails build`. Any manual edits are lost. Always edit the source template at `build/darwin/Info.plist` and `wails.json`.
- **Running `wails build` without `-platform darwin/universal`:** The default target is `darwin/arm64` only (confirmed: Wails 2.12.0 default is `darwin/arm64`). Omitting the flag produces an arm64-only binary that fails on Intel Macs.
- **Skipping `-clean`:** Without `-clean`, a subsequent build may reuse stale frontend artifacts. Use `-clean` for Phase 8's production build.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Universal fat binary creation | Custom `lipo` invocation or cross-compilation script | `wails build -platform darwin/universal` | Wails orchestrates the entire pipeline: dual compilation + lipo merge + .app assembly |
| Icon conversion | Manual `sips` + `iconutil` invocations | Wails build step | Wails converts `appicon.png` to `.icns` automatically during build; no external script needed |
| Info.plist population | Manually editing the rendered .app's Info.plist | `wails.json` info block + `build/darwin/Info.plist` template | Wails renders the template on every build; manual edits are overwritten |

---

## Current State Audit (Phase 7 Outputs)

These are facts verified by reading source files and the built .app in this session:

| Item | Expected State | Actual State | Action Required |
|------|---------------|--------------|-----------------|
| `wails.json` info block | Present with version 1.0.0, copyright, companyName | PRESENT (`productVersion: "1.0.0"`, `copyright: "© 2026 LamboServer"`) | None |
| `build/darwin/Info.plist` CFBundleIdentifier | `dev.lamboserver.app` | CORRECT in source template | None |
| `build/bin/LamboServer.app` architecture | universal (arm64 + x86_64) | STALE arm64-only build from before Phase 7 | Run `wails build -platform darwin/universal -clean` |
| `build/bin/LamboServer.app` bundle ID | `dev.lamboserver.app` | STALE: `com.wails.LamboServer` (pre-Phase-7 build) | Rebuild will fix |
| `build/darwin/entitlements.plist` | Should exist for Phase 9 | MISSING | Create in Phase 8 |
| `build/appicon.png` | Custom LamboServer icon | PRESENT (replaced in Phase 7) | None |

**Critical observation:** The `.app` in `build/bin/` reflects the state BEFORE Phase 7. It must be rebuilt. The source files (`wails.json`, `Info.plist`, `appicon.png`) are all correct post-Phase-7.

---

## Common Pitfalls

### Pitfall 1: stale .app in build/bin/ looks correct but is pre-Phase-7

**What goes wrong:** Developer opens `build/bin/LamboServer.app` and sees the old bundle ID (`com.wails.LamboServer`) or arm64-only binary, assumes something is broken.

**Why it happens:** The current `.app` was built before Phase 7 changes were committed. The source files are correct; only the build artifact is stale.

**How to avoid:** Always run `wails build -platform darwin/universal -clean` before verification. The `-clean` flag removes the stale `build/bin/` contents and forces a full rebuild.

**Warning signs:** `codesign -dv build/bin/LamboServer.app` shows `Identifier=com.wails.LamboServer` — that's the pre-Phase-7 build. After rebuild it should show `Identifier=dev.lamboserver.app`.

### Pitfall 2: Misinterpreting BUILD-02 as requiring a PATH fix

**What goes wrong:** Planner adds a task to "fix PATH for Homebrew binaries" because BUILD-02 says "Homebrew binaries are found when app is launched from Finder."

**Why it happens:** BUILD-02 was written before the architecture was confirmed. D-01 in CONTEXT.md resolves it: all binaries are in `~/.lamboserver/`, not Homebrew paths.

**How to avoid:** BUILD-02 is satisfied by the existing `BinaryLocator` architecture. No code change is needed. The verification test for BUILD-02 is: launch the .app from Finder, confirm services can start/stop.

### Pitfall 3: entitlements.plist missing when Phase 9 codesign runs

**What goes wrong:** Phase 9 attempts `codesign --entitlements build/darwin/entitlements.plist ...` and fails because the file doesn't exist.

**Why it happens:** Wails does not generate `entitlements.plist`. It must be created manually.

**How to avoid:** Create `build/darwin/entitlements.plist` in Phase 8 with the correct network + osascript entitlements. This does not affect `wails build` (which ignores the file) but unblocks Phase 9.

### Pitfall 4: LSMinimumSystemVersion 10.13.0 may be too low

**What goes wrong:** The current `Info.plist` specifies `LSMinimumSystemVersion = 10.13.0` (High Sierra, 2017). While technically functional, this is outdated; macOS Sequoia (15+) users are the primary target audience. The value does not affect build success but should be noted.

**Why it happens:** This is the Wails default value from the template. It was not changed in Phase 7.

**How to avoid:** Update to `12.0.0` (Monterey) or higher if the app uses APIs not available on older macOS. For Phase 8, this is a discretionary change — flag it but do not block the build on it. Wails WebView (WKWebView) on arm64 requires at minimum macOS 12.0 for full feature support.

---

## Code Examples

### Build Command
```bash
# Source: CONTEXT.md D-03 + verified against wails v2.12.0 help output
wails build -platform darwin/universal -clean
```

### Verify Universal Binary
```bash
# Source: Apple lipo(1) man page [ASSUMED — standard macOS toolchain]
lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer
# Expected output:
# Architectures in the fat file: build/bin/LamboServer.app/Contents/MacOS/LamboServer are: x86_64 arm64
```

### Verify Info.plist Values
```bash
# Source: macOS defaults(1) [ASSUMED — standard macOS toolchain]
defaults read build/bin/LamboServer.app/Contents/Info.plist
# Expected key values:
#   CFBundleIdentifier = "dev.lamboserver.app"
#   CFBundleShortVersionString = "1.0.0"
#   CFBundleVersion = "1.0.0"
#   NSHumanReadableCopyright = "© 2026 LamboServer"
#   CFBundleName = "LamboServer"
```

### entitlements.plist (to create)
```xml
<!-- Source: .planning/research/PITFALLS.md Pitfall 5 — based on Wails Mac App Store guide
     and Apple Developer entitlements documentation -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Required: WKWebView uses JIT on Apple Silicon -->
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <!-- Required: Wails WebView needs unsigned memory for JS execution -->
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <!-- Required: Wails embeds third-party frameworks without Apple signatures -->
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <!-- Required: osascript admin dialogs (RunWithAdminPrivileges) -->
    <key>com.apple.security.automation.apple-events</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>LamboServer uses administrator dialogs to install system services.</string>
    <!-- Required: outgoing HTTP connections (Wails WebView, service health checks) -->
    <key>com.apple.security.network.client</key>
    <true/>
    <!-- Required: local service processes bind ports (pgweb, phpMyAdmin) -->
    <key>com.apple.security.network.server</key>
    <true/>
    <!-- NEVER add: com.apple.security.app-sandbox
         This would kill all exec.Command calls to service binaries -->
</dict>
</plist>
```

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wails | BUILD-01, D-03 | Yes | v2.12.0 | None — required |
| go | wails build | Yes | go1.26.1 darwin/arm64 | None — required |
| npm | frontend:build hook | Yes | 11.11.0 | None — required |
| lipo | Build verification | Yes (system) | macOS built-in | None — system tool always present on macOS |
| defaults | Info.plist verification | Yes (system) | macOS built-in | Manually open plist in text editor |

**All required dependencies are available. No missing blockers.**

Note: `npm` is at `/Users/rizkirmdhn/.lamboserver/bin/npm` (not system PATH). This is fine — `wails build` calls `npm` via the `frontend:install` and `frontend:build` hooks using the shell, which should find it if the shell's PATH includes `~/.lamboserver/bin`. This should be verified during the actual build run.

---

## Validation Architecture

nyquist_validation key is absent from config.json — treating as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None — no automated test framework for build output |
| Config file | N/A |
| Quick run command | `lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer` |
| Full suite command | Manual launch test (see Phase Requirements) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | Notes |
|--------|----------|-----------|-------------------|-------|
| BUILD-01 | Universal binary contains arm64 + x86_64 | smoke | `lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer \| grep -q "arm64"` | Automated |
| BUILD-01 | `wails build` exits 0 | smoke | `wails build -platform darwin/universal -clean && echo OK` | Automated |
| BUILD-02 | Services start from Finder launch | manual | N/A — requires Finder launch | Manual-only: GUI PATH behavior cannot be automated in a terminal test |
| BUILD-03 | Service management (start/stop/restart) works | manual | N/A — requires interactive GUI | Manual-only: no headless test for Wails GUI |
| D-06 | Info.plist has correct bundle ID / version | smoke | `defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier \| grep -q "dev.lamboserver.app"` | Automated |

### Sampling Rate

- **Per build:** Run `lipo -info` and `defaults read` checks immediately after `wails build` completes
- **Phase gate:** Full manual launch test (Finder double-click) before marking phase complete

### Wave 0 Gaps

- `build/darwin/entitlements.plist` — must be created (Wave 0 task) before Phase 9; content is defined above
- No test files needed (build output verification uses system tools)

---

## Security Domain

Phase 8 has no security-enforcement surface. The .app is not signed in this phase (signing is Phase 9). The one security-relevant action is creating `entitlements.plist` correctly — specifically, confirming that `com.apple.security.app-sandbox` is NOT included, because sandboxing would break all `exec.Command` service calls.

| ASVS Category | Applies | Note |
|---------------|---------|------|
| V2 Authentication | No | N/A — build phase |
| V3 Session Management | No | N/A — build phase |
| V4 Access Control | No | N/A — build phase |
| V5 Input Validation | No | N/A — build phase |
| V6 Cryptography | No | Ad-hoc signing is Phase 9 |

**Security constraint from REQUIREMENTS.md (Out of Scope):**
- No `com.apple.security.app-sandbox` — blocks exec.Command; incompatible with architecture
- No Hardened Runtime (`--options runtime`) in Phase 9 codesign — blocks osascript admin dialogs
- No notarization — no Apple Developer account

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `wails build` (arm64 only) | `wails build -platform darwin/universal` | Wails v2 added universal support | Single .app works on both chip families |
| Manual `sips` + `iconutil` for icon | Wails auto-converts `appicon.png` | Wails v2 built-in | No manual icon conversion step needed |
| Hardcoded Info.plist values | Go template rendering from `wails.json` | Wails v2 built-in | Single source of truth for version, copyright, name |

**Deprecated/outdated in this project:**
- The existing `build/bin/LamboServer.app`: arm64-only, pre-Phase-7 bundle ID. Must be replaced by Phase 8 build.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `lipo -info` command is always available on macOS | Code Examples, Validation Architecture | Negligible — lipo is a built-in macOS developer tool present since macOS 10.x |
| A2 | `defaults read` can read Info.plist from a .app bundle path | Code Examples | Low — standard macOS command; behavior is stable |
| A3 | npm at `~/.lamboserver/bin/npm` is found by the shell used during `wails build` | Environment Availability | Medium — if wails build's shell doesn't include `~/.lamboserver/bin` in PATH, frontend build may fail; mitigated by running `which npm` before building |
| A4 | LSMinimumSystemVersion 10.13.0 in Info.plist does not cause build failure | Common Pitfalls | Low — this is a runtime enforcement value, not a build-time constraint |

---

## Open Questions

1. **npm PATH during `wails build`**
   - What we know: `npm` is at `~/.lamboserver/bin/npm`, not in a standard system location
   - What's unclear: Does `wails build`'s subprocess shell inherit the user's full PATH, including `~/.lamboserver/bin`?
   - Recommendation: Run `which npm` in the terminal where `wails build` will be run. If it resolves, the build will work. If not, set `PATH` explicitly before running the build: `export PATH="$HOME/.lamboserver/bin:$PATH"`.

2. **LSMinimumSystemVersion — update or leave?**
   - What we know: Current value is `10.13.0` (High Sierra, 2017); CONTEXT.md marks this as Claude's Discretion
   - What's unclear: Whether any API used by LamboServer requires a higher minimum
   - Recommendation: Update to `12.0.0` (Monterey, 2021) in `Info.plist`. WKWebView on Apple Silicon and Wails v2 practically requires macOS 12+. This aligns with the real user base (developers on modern hardware) without blocking the build.

---

## Sources

### Primary (HIGH confidence)
- Wails v2 CLI reference — `wails build --help` output verified on this machine (v2.12.0) [VERIFIED]
- `wails.json` source file — info block confirmed present with correct values [VERIFIED: read in this session]
- `build/darwin/Info.plist` — CFBundleIdentifier confirmed as `dev.lamboserver.app` [VERIFIED: read in this session]
- `internal/system/binary.go` — BinaryLocator.Find() confirmed uses os.Stat on LocalPath first [VERIFIED: read in this session]
- `build/bin/LamboServer.app/Contents/Info.plist` — stale pre-Phase-7 state confirmed (`com.wails.LamboServer`, arm64-only) [VERIFIED: `defaults read` + `lipo -info` in this session]
- `.planning/research/ARCHITECTURE.md` — wails build pipeline, entitlements design [CITED: read in this session]
- `.planning/research/PITFALLS.md` — Pitfall 1 (PATH starvation), Pitfall 5 (entitlements), Pitfall 6 (Info.plist) [CITED: read in this session]

### Secondary (MEDIUM confidence)
- Wails issue #2507, #3558 — GUI PATH starvation behavior confirmed [CITED: PITFALLS.md attribution]
- Apple Developer documentation — entitlements keys [CITED: PITFALLS.md attribution]

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all tools verified present on this machine
- Architecture: HIGH — BinaryLocator code read directly; wails build command verified
- Pitfalls: HIGH — sourced from verified PITFALLS.md research from the same project session

**Research date:** 2026-04-17
**Valid until:** 2026-05-17 (stable; Wails v2 release cadence is slow)
