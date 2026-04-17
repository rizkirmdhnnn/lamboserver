---
phase: 07-app-bundle-icon
plan: "02"
subsystem: build/icons
tags: [icon, macOS, icns, design]
dependency_graph:
  requires: []
  provides: [build/appicon.png, build/darwin/iconfile.icns]
  affects: [wails build, app bundle]
tech_stack:
  added: [Pillow (Python), sips, iconutil]
  patterns: [macOS squircle background, orange wireframe on dark, low-poly polygon car]
key_files:
  created:
    - scripts/gen_icon.py
    - scripts/generate-icns.sh
    - build/darwin/LamboServer.iconset/ (10 PNGs)
    - build/darwin/iconfile.icns
  modified:
    - build/appicon.png
decisions:
  - D-ICON-01: Dark theme with orange wireframe chosen (user feedback: reference image style)
  - D-ICON-02: Low-poly side-profile Lamborghini silhouette — simplified ~12 polygon segments
  - D-ICON-03: </> code tag on windshield to signal developer tool identity
metrics:
  duration: "~5 minutes"
  completed: "2026-04-17"
  tasks_completed: 1
  files_changed: 14
---

# Phase 07 Plan 02: LamboServer Icon — Geometric Lambo Wireframe (Iteration 2) Summary

**One-liner:** Redesigned app icon using orange (#FF7A00) geometric Lamborghini wireframe on dark macOS squircle background, with </> code tag on windshield.

## What Was Built

Per user feedback referencing a dark-theme wireframe car icon, the original monochrome bull/shield design was replaced with:

1. **`scripts/gen_icon.py`** — Python/Pillow script that generates `build/appicon.png` (1024x1024 RGBA):
   - Dark gray rounded-rect background (`#2D2D2D`, corner radius 220px, macOS squircle proportion)
   - Darker inner circle (`#1A1A1A`) behind the car
   - Simplified low-poly Lamborghini side profile — 12 polygon segments forming a wedge silhouette: angular hood, flat roofline, rear drop, wheel arches as semicircular arcs
   - Two filled circle wheels in orange outline
   - `</>` code tag rendered in orange on the windshield area (Menlo font, 64px, fallback to manual line segments)
   - Line width 14px at 1024px scale (scales to ~0.4px at 32x32, thick enough to remain legible)

2. **`scripts/generate-icns.sh`** — Repeatable shell script using `sips` + `iconutil` to produce all 10 macOS iconset sizes from `build/appicon.png` and compile them to `build/darwin/iconfile.icns`.

3. **`build/darwin/LamboServer.iconset/`** — 10 PNG files at all required sizes (16, 32, 64, 128, 256, 512, 1024).

4. **`build/darwin/iconfile.icns`** — Compiled macOS icon file ready for wails build.

## Deviations from Plan

### Auto-fixed Issues

None — this is an iteration continuation. The original plan called for a monochrome bull/shield icon; this iteration implements user's preferred dark-theme wireframe aesthetic per direct feedback.

## Verification Results

- `sips -g pixelWidth -g pixelHeight build/appicon.png` → 1024 x 1024
- `shasum build/appicon.png` → `00eb88c030622cc715af8fec837af10ad4ee3ad3` (not the Wails placeholder)
- `ls build/darwin/LamboServer.iconset/ | wc -l` → 10
- `test -f build/darwin/iconfile.icns` → EXISTS
- `file build/appicon.png` → PNG image data, 1024 x 1024, 8-bit/color RGBA, non-interlaced

## Known Stubs

None — icon is fully realized. The `</>` text uses Menlo font on macOS with manual line-segment fallback.

## Self-Check: PASSED

- build/appicon.png: EXISTS (1024x1024 PNG)
- scripts/generate-icns.sh: EXISTS, executable
- build/darwin/iconfile.icns: EXISTS
- build/darwin/LamboServer.iconset/: EXISTS with 10 PNGs
- Commit f939d2c: EXISTS
