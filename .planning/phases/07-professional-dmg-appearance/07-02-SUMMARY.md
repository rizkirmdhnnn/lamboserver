---
phase: 07-professional-dmg-appearance
plan: 02
subsystem: infra
tags: [create-dmg, dmg, build-script, homebrew, macos-packaging, hdiutil]

# Dependency graph
requires:
  - phase: 07-professional-dmg-appearance
    provides: "Plan 07-01 — build/darwin/dmg-background.png (@1x + @2x) reachable at the hardcoded path used by --background"
  - phase: 06-ci-pipeline-hardening
    provides: "D-06 filename contract (LamboServer-<VERSION>.dmg + .sha256 sidecar) that this plan must preserve byte-identically"
  - phase: 05-signing-correctness
    provides: "Ad-hoc signed .app emerging from step [5/7]; D-02 prohibition on re-signing the DMG container"
provides:
  - "scripts/build-dmg.sh step [6/7] rewritten to use create-dmg with locked layout flags"
  - "Preflight guard that fails fast with brew install instructions when create-dmg is missing"
  - "Defensive gate asserting the expected DMG filename exists before step [7/7] checksums it"
affects:
  - "07-03-release-workflow-update (Wave 2 sibling): release.yml must install create-dmg via brew before invoking this script"
  - "08-* and beyond: any phase touching packaging must preserve DMG_OUT contract and the single-trap invariant"

# Tech tracking
tech-stack:
  added:
    - "create-dmg (Homebrew) — external build dep, invoked as shell command in step [6/7]"
  patterns:
    - "Preflight tool-availability guard using the existing inline `<cmd> || { echo FAIL: ...; exit 1; }` idiom"
    - "Post-invocation filename gate (`test -f ${DMG_OUT}`) to fail fast before the checksum step touches the wrong artifact"
    - "Locked-flag commenting: each create-dmg flag annotated with its decision ID (D-XX) so future edits cannot silently drift layout"

key-files:
  created: []
  modified:
    - "scripts/build-dmg.sh — step [6/7] body: hdiutil+staging replaced by create-dmg invocation; trap, numbering, DMG_OUT, and [7/7] block untouched"

key-decisions:
  - "D-01: Adopted create-dmg (Homebrew) as the DMG builder — enacted by the step [6/7] rewrite"
  - "D-02: --no-code-sign passed — DMG container is not re-signed; inner .app keeps its step [5/7] ad-hoc signature"
  - "D-03: --sandbox-safe passed — avoids AppleScript paths that hang on headless CI runners"
  - "D-04: Surgical replacement of lines 96-107 only; steps [1/7]..[5/7] and [7/7] untouched"
  - "D-05: Preflight `command -v create-dmg` guard added with canonical FAIL message"
  - "D-06: DMG_OUT=\"LamboServer-${VERSION}.dmg\" and `.sha256` sidecar naming preserved byte-identically (Phase 6 CI upload contract)"
  - "D-13: --window-size 660 400 locked"
  - "D-14: --icon-size 128 locked"
  - "D-15: --icon LamboServer.app 165 220 / --app-drop-link 495 220 locked"
  - "D-17: --volname \"${APP_NAME}\" — stable /Volumes/LamboServer mount point"

patterns-established:
  - "Decision-ID flag commenting: every create-dmg flag is paired with a D-XX reference in a preceding comment block"
  - "Post-command filename gate: after the external tool runs, `test -f` asserts the expected output path exists before downstream steps consume it"
  - "Single-trap preservation: STAGING=\"\" declaration and guarded cleanup line kept even though staging path is now dead, to keep the trap contract identical with Phase 5/6"

requirements-completed: [DMG-01, DMG-02, DMG-03]

# Metrics
duration: ~4min
completed: 2026-04-18
---

# Phase 07 Plan 02: create-dmg Build Script Integration Summary

**scripts/build-dmg.sh step [6/7] now drives create-dmg with locked layout flags (660x400 window, 128px icons, icon slot (165,220), app-drop-link (495,220), custom background, ad-hoc-compatible --no-code-sign + --sandbox-safe), preserving the Phase 6 LamboServer-<VERSION>.dmg filename contract and single-trap invariant.**

## Performance

- **Duration:** ~4 minutes
- **Started:** 2026-04-18T11:17:00Z (approximate — worktree base-reset + file reads)
- **Completed:** 2026-04-18T11:21:37Z
- **Tasks:** 1/1
- **Files modified:** 1

## Accomplishments
- Replaced the hdiutil+STAGING packaging block (old lines 96-107) with a single create-dmg invocation carrying all eight locked flags from the plan's `<locked_flags>` block
- Added a preflight `command -v create-dmg` guard with the canonical `FAIL: create-dmg not installed. Run: brew install create-dmg` message (D-05) using the file's existing inline FAIL idiom
- Added a post-invocation `test -f "${DMG_OUT}"` gate so a silent create-dmg failure cannot propagate into step [7/7] and checksum a stale/missing artifact
- Preserved byte-identically: `DMG_OUT`, the `[7/7]` shasum block, the single `trap cleanup EXIT`, `APP_NAME`, `APP_PATH`, and the final post-build echo lines
- Documented each create-dmg flag with its D-XX decision reference inline, so future edits have to actively override a tracked decision to drift the layout

## Task Commits

Each task was committed atomically:

1. **Task 1: Swap step [6/7] body to create-dmg, preserve trap / step numbering / filename contract** — `428381f` (feat)

**Plan metadata:** (SUMMARY commit follows — orchestrator handles STATE.md/ROADMAP.md post-merge per parallel executor contract)

## Files Created/Modified
- `scripts/build-dmg.sh` — step [6/7] body replaced with create-dmg invocation; preflight guard + post-invocation filename gate added; everything outside lines 96-107 byte-identical to the base version

## Decisions Made
- None new. All decisions were pre-locked in 07-CONTEXT.md (D-01..D-06, D-12..D-17). Execution was mechanical transcription of the locked block.
- Preserved `STAGING=""` declaration (line 6) and its guarded cleanup line even though staging is no longer used — deliberately, to keep the trap contract identical with Phase 5/6 (PATTERNS.md rule: "Never add a second `trap ... EXIT`"). The guard short-circuits harmlessly when STAGING stays empty.

## Deviations from Plan

None - plan executed exactly as written.

One observation worth recording, but NOT a deviation I acted on:

### Out-of-Scope Observation

The plan's `<automated>` verification included `[ "$(grep -cE '\[[1-7]/7\]' scripts/build-dmg.sh)" = "7" ]`. The actual count in the base file (and in the modified file) is **8**, because the `[1/7]` marker appears in both branches of the `if [ -n "${VERSION:-}" ]; then ... else ... fi` block introduced by plan 06 (lines 24 and 37). The plan author specified 7 distinct step IDs (coherent sequence), which is satisfied (1,1,2,3,4,5,6,7 → seven distinct IDs, one branch runs). The grep count of 8 is pre-existing and is not caused by this plan's edits. Per the scope-boundary rule, I did not touch it. Exit 2's parallel executor brief also phrases the requirement as "7 `[N/7]` step markers remain (renumber if needed to keep coherent sequence)" — the sequence is coherent; no renumbering needed.

All other plan verification commands passed:

- `bash -n scripts/build-dmg.sh` → OK
- `grep -q "create-dmg \\"` → OK
- `--window-size 660 400` → OK
- `--icon-size 128` → OK
- `--no-code-sign` → OK
- `--sandbox-safe` → OK
- `--app-drop-link 495 220` → OK
- `"${APP_NAME}.app" 165 220` → OK
- `--background "build/darwin/dmg-background.png"` → OK
- `--volname "${APP_NAME}"` → OK
- `FAIL: create-dmg not installed. Run: brew install create-dmg` → OK
- `! grep -q "hdiutil create"` → OK (removed)
- `! grep -q "mktemp -d"` → OK (staging removed)
- `! grep -q "ln -s /Applications"` → OK (create-dmg handles the symlink)
- `grep -cE '^trap '` → 1 (single-trap invariant preserved)
- `DMG_OUT="${APP_NAME}-${VERSION}.dmg"` → OK (D-06 contract)
- `shasum -a 256 "${DMG_OUT}" > "${DMG_OUT}.sha256"` → OK (D-06 sidecar contract)

---

**Total deviations:** 0 auto-fixed
**Impact on plan:** Clean execution. Surgical edit; no scope expansion. The `--icon-size 128` vs plan's `--icon-size 128` (etc.) match exactly.

## Issues Encountered

- **Worktree base drift:** On agent startup, `git merge-base HEAD <expected>` reported `5fbc532...` but the orchestrator-specified base was `5c70baa...`. Executed the documented hard-reset protocol; HEAD is now exactly `5c70baa95610d07d6e6413d612fb5edaa5f67216`. No impact on deliverable — the reset is the documented recovery path for this exact scenario.
- **Planning directory location:** `.planning/` is gitignored in this project. Per the parallel_execution contract, SUMMARY.md is committed with `git add -f`, matching prior phase convention.

## User Setup Required

None - no external service configuration required at runtime. Note: developers running `scripts/build-dmg.sh` locally need `brew install create-dmg` once (the preflight guard now prints this exact instruction on failure).

## Next Phase Readiness

Wave 2 sibling plan 07-03 (`release.yml` + `README.md`) must ensure the CI job installs `create-dmg` via Homebrew before invoking `scripts/build-dmg.sh`. This plan's preflight guard will fail fast on the CI runner if that install step is missing or runs in the wrong order — so the orchestrator's post-wave validation will catch the integration contract immediately on the next tag push.

## Threat Flags

None. The threat register in `07-02-PLAN.md` already enumerated the relevant surface (T-07-05..T-07-10); this execution introduced no new network endpoints, auth paths, file access patterns, or schema changes beyond what was modeled.

## Self-Check: PASSED

- scripts/build-dmg.sh exists and contains the create-dmg invocation with all 8 locked flags (verified via grep pass)
- Commit `428381f` exists on the worktree branch (verified via `git rev-parse --short HEAD` after commit)
- `bash -n scripts/build-dmg.sh` exits 0 (syntax valid)
- Single `trap cleanup EXIT` preserved (grep count = 1)
- DMG_OUT and shasum sidecar lines byte-identical to base (grep matches the exact variable-expansion strings)

---
*Phase: 07-professional-dmg-appearance*
*Completed: 2026-04-18*
