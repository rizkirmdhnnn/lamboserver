---
phase: 09-code-signing-dmg
plan: "01"
subsystem: build-scripts
tags: [codesign, dmg, hdiutil, bash, build-pipeline]
dependency_graph:
  requires: [phase-08-production-build]
  provides: [scripts/build-dmg.sh, dmg-pipeline]
  affects: [.gitignore]
tech_stack:
  added: []
  patterns: [set-e-sequential-script, mktemp-staging-trap-cleanup, adhoc-codesign-entitlements]
key_files:
  created:
    - scripts/build-dmg.sh
  modified:
    - .gitignore
decisions:
  - "Ad-hoc signing (-s -) with --force --deep to override wails pre-sign, then verify before DMG packaging"
  - "mktemp -d staging dir with trap EXIT cleanup prevents hdiutil resource-busy errors"
  - "DMG output at project root (LamboServer-1.0.0.dmg), not build/bin/ (deleted by -clean)"
metrics:
  duration: "~5 minutes"
  completed: "2026-04-17T10:35:23Z"
  tasks_completed: 2
  files_changed: 2
---

# Phase 9 Plan 01: Build DMG Pipeline Summary

**One-liner:** Shell script automating wails build -> ad-hoc codesign with entitlements -> hdiutil UDZO DMG with Applications alias staging dir.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create scripts/build-dmg.sh build pipeline | 7e5bed8 | scripts/build-dmg.sh (created, executable) |
| 2 | Add *.dmg to .gitignore | a841cc4 | .gitignore (appended *.dmg) |

## What Was Built

`scripts/build-dmg.sh` — a single executable shell script that runs the full 4-step pipeline:

1. `wails build -platform darwin/universal -clean` — builds universal binary .app
2. `codesign --force --deep -s - --entitlements build/darwin/entitlements.plist build/bin/LamboServer.app` — ad-hoc signs with all required entitlements; `--force` required because wails pre-signs the bundle
3. `codesign --verify --deep` — validates signature before proceeding to DMG
4. `hdiutil create -volname LamboServer -srcfolder <staging> -ov -format UDZO LamboServer-1.0.0.dmg` — packages signed .app + Applications symlink into compressed read-only DMG

Staging directory pattern (`mktemp -d` + `trap "rm -rf ${STAGING}" EXIT`) avoids "resource busy" hdiutil errors and ensures cleanup on both success and failure.

`.gitignore` updated to exclude `*.dmg` (binary build artifacts).

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — script is complete and self-contained.

## Threat Surface Scan

No new network endpoints, auth paths, or trust boundaries introduced. Threat T-09-01 (tampering of .app after signing) is mitigated: `codesign --verify --deep` validates signature before DMG packaging, per the threat register.

## Self-Check

Files exist:
- scripts/build-dmg.sh: FOUND (executable, 30 lines)
- .gitignore: FOUND (7 lines, *.dmg on line 7)

Commits exist:
- 7e5bed8: FOUND (feat(09-01): add build-dmg.sh pipeline)
- a841cc4: FOUND (chore(09-01): add *.dmg to .gitignore)

## Self-Check: PASSED
