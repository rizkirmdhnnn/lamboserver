---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: Build Pipeline Overhaul
status: phase_complete
stopped_at: Phase 7 complete — human visual DMG check pending
last_updated: "2026-04-18T11:31:01.440Z"
last_activity: 2026-04-18 -- Phase 7 complete (3/3 plans merged, verified)
progress:
  total_phases: 3
  completed_phases: 3
  total_plans: 11
  completed_plans: 11
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-18)

**Core value:** Users can control their local dev services instantly from the system tray without opening the full application window.
**Current focus:** Phase 7 complete — milestone v1.1 done pending human DMG visual check

## Current Position

Phase: 7 (Professional DMG Appearance) — COMPLETE (3/3 plans merged, verified at HEAD 4c11234)
Plan: 3 of 3
Status: Phase 7 complete (human visual DMG smoke test pending)
Last activity: 2026-04-18 -- Phase 7 complete (create-dmg packaging wired end-to-end)

Progress: [██████████] 100%

## Performance Metrics

**Velocity:**

- Total plans completed: 3 (this milestone)
- Average duration: —
- Total execution time: —

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 6 | 3 | - | - |

*Updated after each plan completion*
| Phase 05-signing-correctness P01 | 2min | 2 tasks | 2 files |
| Phase 05-signing-correctness P02 | 1min | 1 tasks | 1 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Research: Do NOT use `spctl --assess` as a build gate — always fails for ad-hoc (not a bug)
- Research: Do NOT sign inner binary separately before bundle — creates CodeResources conflicts on Sequoia
- Research: Use `--sandbox-safe` flag with `create-dmg` for headless CI compatibility
- [Phase 05-signing-correctness]: Removed two-step signing (sign binary then bundle) — replaced with single bundle-level codesign to prevent CodeResources conflict on macOS Sequoia
- [Phase 05-signing-correctness]: codesignBinary failure in ExtractBinary is non-fatal (D-08): warn to stderr, return nil regardless
- [Phase 05-signing-correctness]: Removed 5 inline CI steps; single bash scripts/build-dmg.sh call is the only build step — signing fixes from Plan 01 now apply to CI (D-01, D-02)

### Pending Todos

None yet.

### Blockers/Concerns

- None. Phase 7 resolved the `build/darwin/dmg-background.png` asset blocker via scripts/gen-dmg-background.py (committed PNG + @2x variant).

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| v2 | TRAY-06: Dynamic Dock icon hiding | Deferred | v1.0 Init |
| v2 | TRAY-07: Tray icon state on service failure | Deferred | v1.0 Init |
| v2 | TRAY-08: Auto-start at login | Deferred | v1.0 Init |
| v2 | TRAY-09: Per-site PHP version switching from tray | Deferred | v1.0 Init |
| Future | DIST-01: Apple Developer ID cert / notarization | Deferred | v1.1 Init |
| Future | DIST-02: Homebrew Cask tap | Deferred | v1.1 Init |
| Future | DIST-03: Auto-update mechanism | Deferred | v1.1 Init |

## Session Continuity

Last session: 2026-04-18T10:56:11.982Z
Stopped at: Phase 7 context gathered
Resume file: .planning/phases/07-professional-dmg-appearance/07-CONTEXT.md
