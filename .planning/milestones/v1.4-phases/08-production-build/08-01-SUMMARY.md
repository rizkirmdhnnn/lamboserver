---
phase: 08-production-build
plan: "01"
subsystem: build-config
tags: [codesigning, plist, entitlements, macos, ad-hoc]
dependency_graph:
  requires: []
  provides:
    - build/darwin/entitlements.plist (Phase 9 codesign --entitlements input)
    - build/darwin/Info.plist (updated minimum OS)
    - build/darwin/Info.dev.plist (updated minimum OS)
  affects:
    - Phase 9 codesigning step (consumes entitlements.plist)
    - App Store submission (N/A — ad-hoc only per DIST-01 deferral)
tech_stack:
  added: []
  patterns:
    - ad-hoc codesigning entitlements (no Apple Developer account)
    - WKWebView/Wails-compatible entitlement set (JIT + unsigned memory + no sandbox)
key_files:
  created:
    - build/darwin/entitlements.plist
  modified:
    - build/darwin/Info.plist
    - build/darwin/Info.dev.plist
decisions:
  - "Omitted com.apple.security.app-sandbox: incompatible with exec.Command service management architecture"
  - "Set LSMinimumSystemVersion to 12.0.0: Wails v2 WKWebView on Apple Silicon requires Monterey"
  - "Included com.apple.security.automation.apple-events + NSAppleEventsUsageDescription: needed for osascript admin dialogs"
metrics:
  duration: "~5 minutes"
  completed: "2026-04-17"
  tasks_completed: 2
  files_changed: 3
---

# Phase 08 Plan 01: Build Config — Entitlements and Minimum OS Summary

**One-liner:** Ad-hoc codesigning entitlements plist with 6 WKWebView/Wails-compatible keys and LSMinimumSystemVersion bumped to 12.0.0 in both plist templates.

## What Was Built

Created `build/darwin/entitlements.plist` for use by Phase 9's `codesign --entitlements` invocation. Updated `LSMinimumSystemVersion` from `10.13.0` (High Sierra) to `12.0.0` (Monterey) in both `Info.plist` and `Info.dev.plist`.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create build/darwin/entitlements.plist | 9904c1d, 51f4173 | build/darwin/entitlements.plist |
| 2 | Update LSMinimumSystemVersion in both plist templates | e0a9949 | build/darwin/Info.plist, build/darwin/Info.dev.plist |

## Decisions Made

1. **No app-sandbox entitlement** — The app manages external service processes via `exec.Command` calls to binaries in `~/.lamboserver/`. Sandboxing would kill all of these. Accepted risk per v1.4 architectural decision.

2. **LSMinimumSystemVersion 12.0.0** — Wails v2 WKWebView requires macOS Monterey on Apple Silicon in practice. The previous value (10.13.0 / High Sierra, 2017) was misleading and would allow installation on unsupported systems.

3. **NSAppleEventsUsageDescription included** — The `com.apple.security.automation.apple-events` entitlement requires a usage description string for Gatekeeper/TCC to present to the user when `osascript` elevates privileges for service installation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Comment text contained literal "app-sandbox" substring**
- **Found during:** Task 1 verification
- **Issue:** The original comment `<!-- NEVER add: com.apple.security.app-sandbox -->` caused `! grep -q "app-sandbox"` verification to fail, even though the key itself was absent.
- **Fix:** Rewrote comment to `<!-- Sandboxing is intentionally omitted: it breaks exec.Command calls to service binaries -->` — preserves intent without triggering the grep check.
- **Files modified:** build/darwin/entitlements.plist
- **Commit:** 51f4173

## Verification Results

```
1. test -f build/darwin/entitlements.plist           PASS
2. grep -c "com.apple.security" entitlements.plist   6 (correct)
3. ! grep -q "app-sandbox" entitlements.plist        PASS
4. grep "12.0.0" Info.plist Info.dev.plist           PASS (both)
5. grep "dev.lamboserver.app" Info.plist             PASS
6. grep "NSAllowsLocalNetworking" Info.dev.plist     PASS
```

## Known Stubs

None.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| T-08-01 accepted | build/darwin/entitlements.plist | App runs unsandboxed by design — exec.Command to service binaries is the core architecture. Risk accepted, documented in threat register. |

## Self-Check: PASSED

- `build/darwin/entitlements.plist` exists and contains all 6 required entitlement keys
- `build/darwin/Info.plist` contains `12.0.0`, no `10.13.0`
- `build/darwin/Info.dev.plist` contains `12.0.0`, `NSAllowsLocalNetworking` preserved, no `10.13.0`
- Commits verified: 9904c1d, 51f4173, e0a9949
