# Phase 8: Production Build - Pattern Map

**Mapped:** 2026-04-17
**Files analyzed:** 2 files (1 new, 1 modified) + 1 optional new file
**Analogs found:** 2 / 2 (with analog) + 1 / 1 (no analog)

---

## Overview

Phase 8 is a verification phase, not a development phase. Phase 7 completed all source-code and config changes. The only file deliverables in Phase 8 are:

1. `build/darwin/entitlements.plist` — NEW, required for Phase 9 signing (content fully specified in RESEARCH.md)
2. `build/darwin/Info.plist` — MODIFY one key only: `LSMinimumSystemVersion` from `10.13.0` to `12.0.0` (Claude's Discretion per CONTEXT.md)
3. `Makefile` — OPTIONAL new file, a single `build` target wrapping `wails build -platform darwin/universal -clean` (Claude's Discretion per CONTEXT.md)

The primary build action (`wails build -platform darwin/universal -clean`) is a CLI command, not a source file.

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `build/darwin/entitlements.plist` | config | file-I/O (consumed by codesign at signing time) | `build/darwin/Info.plist` | role-match (same plist XML format, same directory, different Apple subsystem) |
| `build/darwin/Info.plist` | config | transform (Go template rendered at build time) | `build/darwin/Info.dev.plist` | exact (template twins — identical structure) |
| `Makefile` (optional) | config | batch (invokes wails build) | None | no analog (no Makefile or shell scripts exist in project) |

---

## Pattern Assignments

### `build/darwin/entitlements.plist` (config, file-I/O)

**Analog:** `build/darwin/Info.plist`

This is the closest existing analog: same directory, same XML plist format, same role as a build-time configuration file consumed by Apple toolchain utilities.

**Plist file structure pattern** (from `build/darwin/Info.plist` lines 1-5):
```xml
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        ...
    </dict>
</plist>
```

**Complete file content to create** (from RESEARCH.md Code Examples, sourced from `.planning/research/PITFALLS.md` Pitfall 5):
```xml
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

**Critical constraint:** `com.apple.security.app-sandbox` must NOT appear in this file. Adding it would break every `exec.Command` call to service binaries in `~/.lamboserver/`. This is a hard architectural constraint documented in RESEARCH.md Security Domain.

---

### `build/darwin/Info.plist` (config, transform) — discretionary change

**Analog:** `build/darwin/Info.dev.plist` — exact match (template twins)

**Current state** (`build/darwin/Info.plist` line 21):
```xml
<key>LSMinimumSystemVersion</key>
<string>10.13.0</string>
```

**Recommended change — update line 21 value only:**
```xml
<key>LSMinimumSystemVersion</key>
<string>12.0.0</string>
```

**Rationale (from RESEARCH.md Common Pitfalls, Pitfall 4):** Wails v2 with WKWebView on Apple Silicon practically requires macOS 12.0 (Monterey). `10.13.0` is the Wails default (High Sierra, 2017) and is outdated for this project's target audience. This change does not affect `wails build` success — it is a runtime enforcement value only.

**Context: all other keys must remain unchanged** (from `build/darwin/Info.plist` full file):
- `CFBundleIdentifier`: already `dev.lamboserver.app` (set in Phase 7) — do NOT change
- `CFBundleVersion`, `CFBundleShortVersionString`: `{{.Info.ProductVersion}}` — do NOT change
- `NSHumanReadableCopyright`: `{{.Info.Copyright}}` — do NOT change

**Must also apply same change to `build/darwin/Info.dev.plist` line 21** — the two files are structural twins (confirmed in Phase 7 PATTERNS.md). `Info.dev.plist` has one extra block (lines 62-66 `NSAppTransportSecurity`) that must not be removed.

---

### `Makefile` (config, batch) — optional

**Analog:** None — no Makefile or shell scripts exist anywhere in the project.

**Decision guidance (CONTEXT.md Claude's Discretion):** A Makefile is optional. If created, it should contain a single target. The direct `wails build` command is sufficient per D-04.

**Minimal pattern if created:**
```makefile
.PHONY: build

build:
	wails build -platform darwin/universal -clean
```

**Do not create** if it would add complexity without benefit. The single-command build (D-03/D-04) makes a Makefile low-value for Phase 8. Phase 9 will create a full pipeline script — a Makefile could be more useful there.

---

## Shared Patterns

### Plist XML Format
**Source:** `build/darwin/Info.plist` lines 1-63
**Apply to:** `build/darwin/entitlements.plist` (new file)

Both files live in `build/darwin/`. Both are standard Apple Property List XML files. The DOCTYPE declaration and `<plist version="1.0">` wrapper are required by the Apple toolchain (codesign, Wails build).

The `entitlements.plist` format differs from `Info.plist` in one key way: it uses only `<key>` + `<true/>` boolean pairs (no string values, no template variables). The `Info.plist` uses Go template variables (`{{.Info.ProductVersion}}` etc.) — `entitlements.plist` must NOT contain any template variables, since it is read directly by `codesign`, not by Wails template engine.

### LSMinimumSystemVersion — Applied to Both Plist Files
**Source:** `build/darwin/Info.plist` line 21, `build/darwin/Info.dev.plist` line 21
**Apply to:** Both plist files in the same edit pass

The two plist files are structural twins (Phase 7 PATTERNS.md, Shared Patterns). Any change to `Info.plist` must mirror to `Info.dev.plist`. The `LSMinimumSystemVersion` key appears at line 21 in both files.

### BinaryLocator — No Changes Needed
**Source:** `internal/system/binary.go` lines 14-26
**Apply to:** No files in Phase 8 — documented here as a verified non-change

The `BinaryLocator.Find()` method uses `os.Stat(b.LocalPath)` as its first check, bypassing `exec.LookPath` entirely. This satisfies BUILD-02 without any code changes. No PATH fix is needed. Do not modify `binary.go` in Phase 8.

```go
// internal/system/binary.go lines 14-26 — verified correct, no changes needed
func (b *BinaryLocator) Find() string {
    if _, err := os.Stat(b.LocalPath); err == nil {
        return b.LocalPath  // absolute path, no PATH lookup
    }
    if CommandExists(b.Name) {
        if path, err := RunCommand("which", b.Name); err == nil && path != "" {
            return path
        }
    }
    return ""
}
```

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `build/darwin/entitlements.plist` (format only) | config | file-I/O | No existing entitlements plist in project; content is fully specified in RESEARCH.md Code Examples |
| `Makefile` (optional) | config | batch | No Makefile or shell scripts exist in project; pattern is the standard GNU make `.PHONY` convention |

Note: `entitlements.plist` has a role-match analog in `Info.plist` for XML format conventions — the "no analog" here refers specifically to the entitlements key-value content, which has no existing codebase reference.

---

## Build Command Reference

The primary Phase 8 action is a CLI command, not a file edit:

```bash
wails build -platform darwin/universal -clean
```

**Post-build verification commands** (smoke tests per RESEARCH.md Validation Architecture):

```bash
# Verify universal binary (arm64 + x86_64)
lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer
# Expected: Architectures in the fat file: ... are: x86_64 arm64

# Verify bundle ID
defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier
# Expected: dev.lamboserver.app

# Verify version
defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleShortVersionString
# Expected: 1.0.0

# Verify copyright
defaults read build/bin/LamboServer.app/Contents/Info.plist NSHumanReadableCopyright
# Expected: © 2026 LamboServer
```

**npm PATH note (RESEARCH.md Open Question A3):** npm is at `~/.lamboserver/bin/npm`. Before running `wails build`, verify: `which npm`. If it does not resolve, run: `export PATH="$HOME/.lamboserver/bin:$PATH"` in the same terminal session.

---

## Metadata

**Analog search scope:** `/Users/rizkirmdhn/Documents/Code/lamboserver/build/darwin/` (plist files), `internal/system/binary.go`
**Files scanned:** `build/darwin/Info.plist`, `build/darwin/Info.dev.plist`, `internal/system/binary.go`, `wails.json`, `.planning/phases/07-app-bundle-icon/07-PATTERNS.md`
**Pattern extraction date:** 2026-04-17

**Key observations:**
1. Phase 8 has only one mandatory new file: `build/darwin/entitlements.plist`. Its full content is specified in RESEARCH.md Code Examples — the planner can use it verbatim.
2. `LSMinimumSystemVersion` update is Claude's Discretion — low risk, high correctness value. Recommend including it.
3. The Makefile is low-value for Phase 8 alone; defer to Phase 9 if a pipeline script is needed.
4. Both plist files (`Info.plist` and `Info.dev.plist`) are structural twins — any edit to one requires the same edit to the other.
5. No Go source files change in Phase 8. `BinaryLocator` is verified correct and needs no modification.
