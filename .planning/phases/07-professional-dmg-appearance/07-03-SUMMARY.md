---
phase: 07-professional-dmg-appearance
plan: 03
subsystem: build-and-docs
tags: [ci, release, homebrew, create-dmg, docs]
requires:
  - scripts/build-dmg.sh (Plan 07-02) exposes a D-05 preflight that needs create-dmg on PATH
  - .github/workflows/release.yml (Phase 06) bootstrap step cluster
provides:
  - CI runner has create-dmg installed before scripts/build-dmg.sh executes
  - Local-dev README docs a single `brew install create-dmg` prerequisite
affects:
  - .github/workflows/release.yml (Release DMG workflow, macos-15 job)
  - README.md §Build
tech-stack:
  added: []
  patterns:
    - CI single-line bootstrap step (mirrors "Install Wails" shape)
    - README comment-header `# <Verb phrase>` inside fenced bash blocks
key-files:
  created: []
  modified:
    - .github/workflows/release.yml
    - README.md
decisions:
  - D-18 (CI install step between "Get version from tag" and "Build, sign, and package")
  - D-19 (no Homebrew caching — accept ~5–15s install vs cache desync risk)
  - D-20 (README one-liner; CI installs automatically, local dev runs once)
requirements:
  - DMG-01 (supports the create-dmg bootstrap chain)
metrics:
  duration_seconds: 62
  completed: 2026-04-18
  tasks_completed: 2
  files_modified: 2
---

# Phase 7 Plan 3: CI + Docs for create-dmg Homebrew Install Summary

One-liner: Wired the `create-dmg` Homebrew install into the release CI (new dedicated step before `bash scripts/build-dmg.sh`) and documented the same `brew install create-dmg` prerequisite inside README §Build, satisfying D-18, D-19, and D-20.

## Tasks Completed

| Task | Name                                                     | Commit   | Files                              |
| ---- | -------------------------------------------------------- | -------- | ---------------------------------- |
| 1    | Add "Install create-dmg" step to release.yml             | 2d7e3a4  | .github/workflows/release.yml      |
| 2    | Document `brew install create-dmg` in README §Build      | d9ab195  | README.md                          |

## What Changed

### `.github/workflows/release.yml`
Inserted a new step between "Get version from tag" and "Build, sign, and package":

```yaml
      - name: Install create-dmg
        run: brew install create-dmg
```

Shape mirrors the existing "Install Wails" step (single `- name:` + single-line `run:`). No `with:`, no `env:`, no `actions/cache`. Runner pin `runs-on: macos-15` unchanged. Final step order:

1. `actions/checkout@v4`
2. Setup Go
3. Setup Node.js
4. Install Wails
5. Install frontend dependencies
6. Get version from tag
7. **Install create-dmg** (new)
8. Build, sign, and package
9. Compose release notes
10. Create GitHub Release

### `README.md`
Added the `brew install create-dmg` prerequisite as the first command inside the existing §Build fenced bash block:

```bash
# Install DMG packaging tool (one-time, local dev only; CI installs this automatically)
brew install create-dmg

# Build .app
wails build -platform darwin/universal -clean

# Build DMG (ad-hoc signed)
./scripts/build-dmg.sh
```

Preserved the §Build heading (still a single occurrence), the existing `# <Verb phrase>` comment-header style, the `wails build` → `./scripts/build-dmg.sh` ordering, the blank-line separator cadence, and the trailing "The built app will be at `build/bin/LamboServer.app`." sentence. All other README sections (§Download, §Features, §Setup, §Architecture, §Tech Stack, §Contributing, §License, §Author) untouched.

## Verification

All six plan-level checks pass:

1. `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"` — valid YAML
2. Step-order assertion — `Get version from tag` (index 5) < `Install create-dmg` (index 6) < `Build, sign, and package` (index 7)
3. `grep -q "brew install create-dmg" .github/workflows/release.yml` — present
4. `grep -q "brew install create-dmg" README.md` — present
5. README §Build ordering assertion — `brew install create-dmg` appears before `wails build`, which appears before `./scripts/build-dmg.sh`
6. `! grep -qE "cache:\s*(homebrew|brew)" .github/workflows/release.yml` — no Homebrew caching introduced (D-19 satisfied)

Additional structural checks:
- `grep -q "runs-on: macos-15" .github/workflows/release.yml` — runner pin unchanged
- `grep -cE "^### Build$" README.md` = 1 — heading not duplicated

## Deviations from Plan

None — plan executed exactly as written. Both edits were surgical inserts (3 new lines apiece) with no scope creep, no Rule 1–3 auto-fixes, and no architectural questions.

## Authentication Gates

None.

## Decisions Made

No new decisions — Plan 07-03 executed against pre-locked phase decisions D-18, D-19, and D-20. All three decisions were honored verbatim.

## Known Stubs

None.

## Commits

- `2d7e3a4` ci(07-03): add Install create-dmg step before build-dmg.sh
- `d9ab195` docs(07-03): document brew install create-dmg in README §Build

## Self-Check: PASSED

- `.github/workflows/release.yml`: FOUND (modified, 74 lines, parses as valid YAML)
- `README.md`: FOUND (modified, §Build updated, all adjacent sections preserved)
- Commit `2d7e3a4`: FOUND in `git log`
- Commit `d9ab195`: FOUND in `git log`
