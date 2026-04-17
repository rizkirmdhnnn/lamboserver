---
phase: 09-code-signing-dmg
plan: "02"
subsystem: build-scripts
tags: [codesign, dmg, hdiutil, release-notes, gatekeeper, macos-sequoia]

# Dependency graph
requires:
  - phase: 09-01
    provides: scripts/build-dmg.sh pipeline script
  - phase: 08-production-build
    provides: build/darwin/entitlements.plist, wails build infrastructure
provides:
  - LamboServer-1.0.0.dmg (signed universal binary DMG, gitignored build artifact)
  - RELEASE_NOTES.md (Gatekeeper workaround for GitHub release)
affects: [github-release-distribution]

# Tech tracking
tech-stack:
  added: []
  patterns: [adhoc-codesign-with-entitlements, udzo-dmg-drag-to-applications]

key-files:
  created:
    - RELEASE_NOTES.md
  modified: []

key-decisions:
  - "RELEASE_NOTES.md goes in project root only (not README.md, not inside DMG) per D-10"
  - "Gatekeeper workaround path: System Settings > Privacy & Security > Open Anyway (macOS Sequoia)"
  - "xattr terminal alternative included for developer users"

patterns-established:
  - "Release notes pattern: separate RELEASE_NOTES.md for GitHub release copy-paste, not embedded in README"

requirements-completed: [DMG-01, DMG-02, DOC-01]

# Metrics
duration: 1min
completed: 2026-04-17
---

# Phase 9 Plan 02: DMG Build & Release Notes Summary

**Full pipeline executed: wails build -> ad-hoc codesign with entitlements -> hdiutil UDZO DMG; RELEASE_NOTES.md with macOS Sequoia Gatekeeper workaround ready for GitHub release.**

## Performance

- **Duration:** ~5 min (dominated by wails build compilation)
- **Started:** 2026-04-17T10:37:13Z
- **Completed:** 2026-04-17T10:38:30Z
- **Tasks:** 2 complete (Task 3 is human-verify checkpoint)
- **Files modified:** 1

## Accomplishments

- `bash scripts/build-dmg.sh` ran to completion without errors; exit code 0
- Signed .app passes `codesign --verify --deep --strict` and contains `allow-jit` entitlement
- DMG mounts at `/Volumes/LamboServer/` with `LamboServer.app` and `Applications` symlink in UDZO compressed read-only format
- RELEASE_NOTES.md created with exact Gatekeeper workaround path for macOS Sequoia: System Settings > Privacy & Security > Open Anyway

## Task Commits

Each task was committed atomically:

1. **Task 1: Run build-dmg.sh and verify pipeline output** - No commit (build artifacts only; LamboServer-1.0.0.dmg is gitignored per .gitignore *.dmg rule from plan 09-01)
2. **Task 2: Create RELEASE_NOTES.md with Gatekeeper workaround** - `9f9a078` (docs)
3. **Task 3: Verify DMG layout and release notes in Finder** - CHECKPOINT (awaiting human verification)

## Files Created/Modified

- `RELEASE_NOTES.md` - GitHub release notes with Gatekeeper workaround instructions for macOS Sequoia; includes System Settings > Privacy & Security path, Open Anyway button, xattr terminal alternative, and "only once per installation" note

## Decisions Made

- RELEASE_NOTES.md in project root per D-10 (GitHub release notes only, not README.md or inside DMG)
- Gatekeeper workaround uses exact macOS Sequoia UI path: System Settings > Privacy & Security
- xattr terminal alternative included for developers who prefer CLI

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - build pipeline ran cleanly on first attempt.

## Known Stubs

None - RELEASE_NOTES.md is complete and accurate for the v1.0.0 release.

## Threat Surface Scan

No new network endpoints, auth paths, or trust boundaries introduced.

T-09-03 (Spoofing - Gatekeeper workaround social engineering) is mitigated: RELEASE_NOTES.md documents the exact System Settings > Privacy & Security > Open Anyway path so users can distinguish the legitimate workaround from phishing.

T-09-06 (Information Disclosure - release notes expose architecture) accepted: file contains only user-facing installation instructions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- v1.4 macOS Installer milestone is complete after Task 3 human verification
- LamboServer-1.0.0.dmg is ready for distribution via GitHub release
- RELEASE_NOTES.md is ready to copy-paste into GitHub release description

## Self-Check

Files exist:
- RELEASE_NOTES.md: FOUND (31 lines, contains Privacy & Security, Open Anyway, xattr command)
- LamboServer-1.0.0.dmg: FOUND (13.5MB, gitignored build artifact)

Commits exist:
- 9f9a078: FOUND (docs(09-02): add RELEASE_NOTES.md with Gatekeeper workaround for macOS Sequoia)

## Self-Check: PASSED

---
*Phase: 09-code-signing-dmg*
*Completed: 2026-04-17*
