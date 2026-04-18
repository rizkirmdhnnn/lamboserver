---
phase: 05-signing-correctness
plan: 02
subsystem: infra
tags: [ci, release, workflow, codesign, build-script]

# Dependency graph
requires:
  - 05-01 (build-dmg.sh refactored with single-step signing and verification gates)
provides:
  - CI pipeline delegates entire build-sign-package to build-dmg.sh (SIGN-01, SIGN-04 via D-01/D-02)
affects: [.github/workflows/release.yml]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "CI as thin wrapper: setup tools only, delegate build+sign+package to shared script"
    - "Single bash scripts/build-dmg.sh call replaces 5 inline CI steps"
    - "CGO_ENABLED=1 set as step env var for CI environment propagation"

key-files:
  created: []
  modified:
    - .github/workflows/release.yml

key-decisions:
  - "Removed 5 inline CI steps in favor of single bash scripts/build-dmg.sh call — signing fixes from Plan 01 now automatically apply to CI (D-01, D-02)"
  - "CGO_ENABLED=1 set as env var on the script step for CI environment safety"
  - "Version step (id: version) preserved — release upload still uses steps.version.outputs.VERSION for DMG filename; script uses wails.json version independently (Phase 6 CI-05 will sync these)"

requirements-completed: [SIGN-01, SIGN-04]

# Metrics
duration: 1min
completed: 2026-04-18
---

# Phase 5 Plan 02: Signing Correctness — CI Pipeline Simplification Summary

**CI pipeline delegating entire build-sign-package pipeline to build-dmg.sh, eliminating 5 inline steps and ensuring signing fixes from Plan 01 automatically apply to CI**

## Performance

- **Duration:** ~1 min
- **Started:** 2026-04-18T08:55:17Z
- **Completed:** 2026-04-18T08:55:58Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments

- Replaced 5 inline CI steps (Build universal .app, Verify CGO tray linkage, Clear extended attributes, Ad-hoc codesign, Create DMG) with a single `bash scripts/build-dmg.sh` call
- CI now uses the same build+sign+package script as local builds — no more drift between environments
- Signing fixes from Plan 01 (single-step bundle codesign with --options runtime, Gate 1/Gate 2 hard-failure verification) now automatically apply to CI without any additional CI changes
- All tool setup steps and the GitHub Release upload step are preserved unchanged

## Task Commits

Each task was committed atomically:

1. **Task 1: Replace inline CI build/sign/DMG steps with single build-dmg.sh call** - `57274f3` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `.github/workflows/release.yml` - 5 inline build/sign/verify/DMG steps replaced with single `bash scripts/build-dmg.sh` step; all setup and release steps preserved

## Decisions Made

- Removed 5 inline CI steps; single `bash scripts/build-dmg.sh` call is now the only build step — eliminates CI drift (D-01, D-02)
- `CGO_ENABLED: "1"` set as env var on the script step to ensure propagation in CI environment (script also sets it inline in wails build command as belt-and-suspenders)
- Version step `id: version` kept — GitHub Release step uses `steps.version.outputs.VERSION` for DMG filename; script derives version from wails.json independently; both will be synced in Phase 6 (CI-05)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 5 is complete — all 2 plans executed
- Phase 6 (CI pipeline improvements) can proceed: pinning macos-15 runner (CI-01), pinning Wails version (CI-02), version sync between wails.json and git tag (CI-05)

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes introduced. Change is CI workflow configuration only.

---
*Phase: 05-signing-correctness*
*Completed: 2026-04-18*
