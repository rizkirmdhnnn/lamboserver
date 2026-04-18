---
phase: 07-professional-dmg-appearance
plan: 01
subsystem: infra
tags: [dmg, pillow, python, assets, packaging, macos]

requires:
  - phase: 06-ci-pipeline-hardening
    provides: CI release pipeline stable; this plan's PNG outputs are committed artifacts consumed by plan 07-02.
provides:
  - Deterministic Pillow-based generator `scripts/gen-dmg-background.py` for DMG background artwork
  - Committed `build/darwin/dmg-background.png` (660×400 RGBA) — the @1x DMG installer background
  - Committed `build/darwin/dmg-background@2x.png` (1320×800 RGBA) — the Retina variant auto-picked by create-dmg
affects:
  - 07-professional-dmg-appearance plan 02 (create-dmg integration — consumes the committed PNGs)
  - 07-professional-dmg-appearance plan 03 (docs/README updates — references the generator)

tech-stack:
  added: [Pillow (dev-only, D-08)]
  patterns:
    - "Pillow generator mirrors scripts/gen_icon.py idioms: no shebang, module docstring, hex constants, REPO_ROOT-derived paths, RGBA canvas, Unicode box-drawing section headers, system-font fallback list, textbbox centering, os.makedirs+save+print"
    - "Script-to-asset split: generator in scripts/, committed PNG artifacts under build/darwin/; build-dmg.sh consumes committed artifacts and does NOT re-run the generator"
    - "Scale-parameterized render(): single render(scale) function renders both @1x and @2x at native resolution (no resize) per D-12"

key-files:
  created:
    - scripts/gen-dmg-background.py
    - build/darwin/dmg-background.png
    - build/darwin/dmg-background@2x.png
  modified: []

key-decisions:
  - "Reused #FF7A00 brand orange from scripts/gen_icon.py as ACCENT for visual continuity (D-11 'subtle accent auto-derived from the app icon's dominant color' — orange IS the dominant color)"
  - "Implemented D-11 'light/neutral palette' as a subtle vertical gradient #FFFFFF → #F5F5F7 rendered by horizontal 1px strips (deterministic, no random state)"
  - "Rendered @2x natively at 1320×800 via a shared render(scale) function instead of resizing the @1x — keeps Retina variant crisp per D-12"
  - "Kept caption 'Drag to Applications' exact — canonical copy per CONTEXT §Specific Ideas"
  - "Omitted shebang and argparse / CLI flags to mirror scripts/gen_icon.py exactly (hardcoded output paths derived from REPO_ROOT via __file__)"
  - "Chose #8E8E93 (system mid-gray) for the chevron arrow — 'subtle' per D-11 without competing with the brand orange"

patterns-established:
  - "DMG asset generator template: any future DMG background redesign edits scripts/gen-dmg-background.py and re-runs it; committed PNGs are the review surface (diffable via git image-diff or regenerate-and-compare)"
  - "Scale-parameterized composition: render(scale: int) -> Image.Image is the reusable shape for any future 1x+2x asset generator in this repo"

requirements-completed: [DMG-01, DMG-03]

duration: 2m 10s
completed: 2026-04-18
---

# Phase 07 Plan 01: Generate DMG Background PNGs Summary

**Pillow-based deterministic generator plus committed 660×400 and 1320×800 RGBA PNGs for the polished DMG installer backdrop, replicating scripts/gen_icon.py conventions.**

## Performance

- **Duration:** 2m 10s
- **Started:** 2026-04-18T11:15:23Z
- **Completed:** 2026-04-18T11:17:33Z
- **Tasks:** 2
- **Files modified:** 3 (all newly created)

## Accomplishments

- `scripts/gen-dmg-background.py` created — deterministic Pillow generator that reads `build/appicon.png` (D-09) and emits both `dmg-background.png` (660×400) and `dmg-background@2x.png` (1320×800).
- Generator mirrors every structural idiom of `scripts/gen_icon.py`: no shebang, module docstring documenting design/palette, hex color constants, `REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))` path resolution, RGBA canvas, Unicode box-drawing section headers, system-font fallback list, textbbox-centered caption, and `os.makedirs(..., exist_ok=True)` + `img.save(..., "PNG")` + `print(f"Saved: ...")` save block.
- Both committed PNGs have the exact required dimensions and 8-bit RGBA format (confirmed via `file` and Pillow `.size`).
- Determinism verified: regenerating produced byte-identical SHA-256 hashes for both outputs.

## Task Commits

Each task was committed atomically:

1. **Task 1: Create gen-dmg-background.py with composition logic for 1x and 2x** — `c1566c5` (feat)
2. **Task 2: Run generator and commit both PNG assets** — `5407748` (feat)

_Plan metadata commit is handled by the orchestrator after worktree merge per the parallel-executor contract._

## Files Created/Modified

- `scripts/gen-dmg-background.py` — 127-line Pillow generator with `render(scale: int) -> Image.Image` function and unguarded top-level save loop. Writes both `build/darwin/dmg-background.png` and `build/darwin/dmg-background@2x.png`.
- `build/darwin/dmg-background.png` — 660×400 8-bit RGBA PNG (~20 KB). SHA-256: `a541ff66a33d7e1cd38c8c3f2c97f4df17596ba97dba90f64b2cf61dbf84f659`.
- `build/darwin/dmg-background@2x.png` — 1320×800 8-bit RGBA PNG (~46 KB). SHA-256: `8f58b8a974b0a22fb0b65673b61795a0e15ba08a1b015fb0038765be8768e014`.

## Decisions Made

- **Brand orange (#FF7A00) as ACCENT:** Reused the value from `scripts/gen_icon.py` to honor D-11's "subtle accent auto-derived from the app icon's dominant color" — keeps palette constants centralized conceptually across both generators.
- **Subtle vertical gradient (#FFFFFF → #F5F5F7):** Drawn by filling 1px horizontal strips in a deterministic `for y in range(H)` loop. Satisfies D-11 "off-white or very light gradient" and remains safe under both Light and Dark Finder.
- **Native @2x rendering via `render(scale)`:** Rather than rendering @1x and resizing to @2x, the single `render(scale)` function scales every coordinate and line width by `scale`, producing crisp Retina artwork per D-12.
- **Graceful font fallback:** If none of the four macOS system sans-serif paths resolve, the caption is skipped (not a crash). Matches `gen_icon.py`'s fallback pattern but simplified because Pillow's default bitmap font would look jarring at 28px; skipping is the safer default.
- **Unguarded top-level save loop:** Kept the save block at module level (not inside `if __name__ == "__main__":`) to match `scripts/gen_icon.py` exactly. Running `python3 scripts/gen-dmg-background.py` performs the save, mirroring the established invocation pattern.

## Deviations from Plan

None — plan executed exactly as written. All eight plan-level acceptance criteria for Task 1 and all five for Task 2 passed on first run. Pillow was already installed locally (v11.1.0), so the `pip install --user Pillow` fallback was not needed.

## Issues Encountered

- Initial Write tool call accidentally wrapped the save loop in `if __name__ == "__main__":`, diverging from the `scripts/gen_icon.py` analog. Corrected via Edit to restore the unguarded top-level save loop before any commit. No impact on functionality or commit history.

## User Setup Required

None — Pillow remains a dev-only dependency (D-08). CI does not need Pillow installed because `build-dmg.sh` reads the committed PNGs directly.

## Next Phase Readiness

- Plan 07-02 (create-dmg integration) can now pass `--background "build/darwin/dmg-background.png"` to create-dmg; the Retina sibling at `dmg-background@2x.png` will be auto-discovered.
- Both PNGs are reproducible byte-for-byte — reviewers can verify integrity by running `python3 scripts/gen-dmg-background.py` and comparing hashes.

## Self-Check

Verification before reporting completion:

- `scripts/gen-dmg-background.py` — FOUND (127 lines, syntax-valid, contains `def render`, references `appicon.png`, `dmg-background.png`, `dmg-background@2x.png`, `Drag to Applications`, `RGBA`, no shebang, `from PIL import` present).
- `build/darwin/dmg-background.png` — FOUND (660×400 PNG, 8-bit RGBA).
- `build/darwin/dmg-background@2x.png` — FOUND (1320×800 PNG, 8-bit RGBA).
- Commit `c1566c5` — FOUND in git log.
- Commit `5407748` — FOUND in git log.
- Determinism — CONFIRMED (byte-identical SHA-256 on re-run).

## Self-Check: PASSED

---
*Phase: 07-professional-dmg-appearance*
*Completed: 2026-04-18*
