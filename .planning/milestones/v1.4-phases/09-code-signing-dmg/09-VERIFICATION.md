---
phase: 09-code-signing-dmg
verified: 2026-04-17T11:00:00Z
status: human_needed
score: 6/8 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Run bash scripts/build-dmg.sh and confirm exit code 0; check codesign -dv build/bin/LamboServer.app shows Signature=adhoc"
    expected: "Script runs to completion without errors; codesign output contains Signature=adhoc"
    why_human: "Cannot run wails build (multi-minute compilation) or codesign verification without the live toolchain and a build environment"
  - test: "Run codesign --verify --deep --strict build/bin/LamboServer.app and codesign -d --entitlements - build/bin/LamboServer.app"
    expected: "First command exits 0; second shows allow-jit in output confirming entitlements are embedded"
    why_human: "Requires a signed .app built by the script; the DMG is gitignored so this must be confirmed after running the build pipeline"
  - test: "Double-click LamboServer-1.0.0.dmg in Finder; verify window shows LamboServer.app with custom icon and an Applications folder shortcut with alias arrow"
    expected: "DMG mounts at /Volumes/LamboServer/ with LamboServer.app and Applications (symlink) visible side-by-side; drag to Applications works"
    why_human: "Visual Finder layout and DMG mount behaviour cannot be verified programmatically; Task 3 in Plan 02 was a blocking human-verify checkpoint"
---

# Phase 9: Code Signing & DMG Verification Report

**Phase Goal:** LamboServer is distributed as a signed DMG that users can install by dragging to Applications
**Verified:** 2026-04-17T11:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | scripts/build-dmg.sh exists and is executable | VERIFIED | File present at `scripts/build-dmg.sh`; `test -x` passes; 30 lines |
| 2 | Script runs wails build, codesign with entitlements, hdiutil create in sequence | VERIFIED | All three commands confirmed in order: `wails build -platform darwin/universal -clean`, `codesign --force --deep -s - --entitlements`, `hdiutil create -format UDZO` |
| 3 | Script produces LamboServer-1.0.0.dmg in project root | VERIFIED | `DMG_OUT="${APP_NAME}-${VERSION}.dmg"` expands to `LamboServer-1.0.0.dmg` with no path prefix (project root output confirmed) |
| 4 | DMG files are excluded from git via .gitignore | VERIFIED | `.gitignore` line 7 is `*.dmg`; all 6 original lines retained; file has 7 lines total |
| 5 | Release notes contain Gatekeeper workaround for macOS Sequoia | VERIFIED | `RELEASE_NOTES.md` contains: "System Settings", "Privacy & Security", "Open Anyway", `xattr -r -d com.apple.quarantine`, "appears only once per installation", "LamboServer v1.0.0" header |
| 6 | Single build script runs full pipeline without manual steps | VERIFIED | `scripts/build-dmg.sh` is a self-contained 4-step script; no manual intervention required between steps; trap-based cleanup on exit |
| 7 | build-dmg.sh runs to completion and produces LamboServer-1.0.0.dmg | HUMAN NEEDED | Summary claims exit code 0 and DMG at 13.5MB; cannot verify without running the build pipeline (DMG is gitignored) |
| 8 | Signed .app inside DMG passes codesign --verify --deep; entitlements embedded | HUMAN NEEDED | Script contains the verify step (`codesign --verify --deep`); entitlements.plist exists with all required keys; but runtime signature cannot be verified without running the build |
| 9 | DMG opens to show LamboServer.app and Applications folder alias | HUMAN NEEDED | Script creates the correct staging structure (`ln -s /Applications`); but Finder DMG layout is visual behaviour requiring human confirmation (Plan 02 Task 3 was a blocking human-verify checkpoint that has not been recorded as completed) |

**Score:** 6/8 truths verified (2 truths require human verification; 1 additional human-needed for codesign runtime)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `scripts/build-dmg.sh` | Full build pipeline: wails build -> codesign -> hdiutil DMG | VERIFIED | 30 lines; executable; shebang `#!/bin/bash`; `set -e`; all 4 pipeline steps present; no anti-patterns (`--options runtime` absent; `build/bin/*.dmg` output absent) |
| `.gitignore` | Excludes DMG build artifacts from version control | VERIFIED | 7 lines total; `*.dmg` on line 7; all original 6 lines intact |
| `RELEASE_NOTES.md` | Gatekeeper workaround instructions for GitHub release | VERIFIED | 31 lines; all required content present: Privacy & Security path, Open Anyway button, xattr terminal alternative, once-only note |
| `build/darwin/entitlements.plist` | Network entitlements (no sandbox, no Hardened Runtime) | VERIFIED | Contains `allow-jit`, `allow-unsigned-executable-memory`, `disable-library-validation`, `apple-events`, `network.client`, `network.server`; no sandbox key present |
| `LamboServer-1.0.0.dmg` | Distributable DMG installer | HUMAN NEEDED | Build artifact; gitignored per plan; Summary reports 13.5MB file was produced; runtime verification required |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `scripts/build-dmg.sh` | `build/darwin/entitlements.plist` | `--entitlements` flag | VERIFIED | Line 14: `codesign --force --deep -s - --entitlements "${ENTITLEMENTS}" "${APP_PATH}"` where `ENTITLEMENTS="build/darwin/entitlements.plist"` |
| `scripts/build-dmg.sh` | `build/bin/LamboServer.app` | wails build output consumed by codesign | VERIFIED | `APP_PATH="build/bin/${APP_NAME}.app"` used in codesign and cp steps |
| `scripts/build-dmg.sh` | `LamboServer-1.0.0.dmg` | hdiutil create output | VERIFIED | `hdiutil create ... "${DMG_OUT}"` where `DMG_OUT="${APP_NAME}-${VERSION}.dmg"` = `LamboServer-1.0.0.dmg` at project root |
| `RELEASE_NOTES.md` | System Settings (Gatekeeper workaround) | Exact UI path documented | VERIFIED | Contains: "Open **System Settings** > **Privacy & Security**" and "Click **Open Anyway**" |

### Data-Flow Trace (Level 4)

Not applicable — this phase produces shell scripts and documentation, not components that render dynamic data.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| build-dmg.sh has correct codesign syntax | `grep 'codesign --force --deep -s - --entitlements' scripts/build-dmg.sh` | Match found | PASS |
| build-dmg.sh has hdiutil UDZO format | `grep 'UDZO' scripts/build-dmg.sh` | Match found | PASS |
| build-dmg.sh creates Applications symlink | `grep 'ln -s /Applications' scripts/build-dmg.sh` | Match found | PASS |
| RELEASE_NOTES.md has Privacy & Security path | `grep 'Privacy & Security' RELEASE_NOTES.md` | Match found | PASS |
| Full pipeline run to completion | `bash scripts/build-dmg.sh` | SKIP — requires wails toolchain and ~5 min compilation; live DMG not in repo | SKIP |
| codesign verify on signed .app | `codesign --verify --deep build/bin/LamboServer.app` | SKIP — requires a prior successful build run | SKIP |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| SIGN-01 | 09-01, 09-02 | .app bundle is ad-hoc signed with `codesign -s -` | VERIFIED (static) / HUMAN NEEDED (runtime) | Script contains `codesign --force --deep -s - --entitlements`; runtime execution verified by Summary claim only |
| SIGN-02 | 09-01 | `entitlements.plist` exists with network entitlements (no sandbox, no Hardened Runtime) | VERIFIED | `build/darwin/entitlements.plist` present with 6 required keys; no sandbox key; no Hardened Runtime |
| SIGN-03 | 09-01, 09-02 | Signed app passes `codesign --verify --deep` | VERIFIED (static) / HUMAN NEEDED (runtime) | Script contains `codesign --verify --deep "${APP_PATH}"` after signing; runtime result requires human confirmation |
| DMG-01 | 09-02 | DMG installer created with drag-to-Applications layout | VERIFIED (static) / HUMAN NEEDED (visual) | Script creates correct staging: `cp -r .app`, `ln -s /Applications`, `hdiutil create -format UDZO`; Finder layout requires human confirmation |
| DMG-02 | 09-01 | Build script automates full pipeline (build, sign, package DMG) | VERIFIED | `scripts/build-dmg.sh` is a complete 4-step pipeline; no manual steps required between stages |
| DOC-01 | 09-02 | Release notes include Gatekeeper workaround instructions (System Settings > Privacy & Security > Open Anyway) | VERIFIED | `RELEASE_NOTES.md` contains all required elements per DOC-01 |

All 6 required requirement IDs (SIGN-01, SIGN-02, SIGN-03, DMG-01, DMG-02, DOC-01) are claimed across Plans 09-01 and 09-02 and are accounted for above. No orphaned requirements for Phase 9 exist in REQUIREMENTS.md.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | None found | — | No placeholders, stubs, or TODO comments present in any phase 9 artifact |

Anti-pattern scan confirmed:
- No `TODO/FIXME/XXX/HACK/PLACEHOLDER` in `scripts/build-dmg.sh` or `RELEASE_NOTES.md`
- No `--options runtime` (Hardened Runtime anti-pattern) in `scripts/build-dmg.sh`
- No DMG output routed to `build/bin/` (which `-clean` deletes)
- No `hdiutil attach` mount loop pattern (uses `-srcfolder` correctly)
- `set -e` present; `trap ... EXIT` cleanup present

### Human Verification Required

#### 1. Full Pipeline Execution

**Test:** From the project root, run `bash scripts/build-dmg.sh` and observe it run to completion.
**Expected:** 4 progress lines printed (`[1/4]` through `[4/4]`); exit code 0; `LamboServer-1.0.0.dmg` appears in project root (13–15 MB).
**Why human:** Requires the wails toolchain, Go compiler, and ~5 minutes of compilation time; cannot be run programmatically in the verifier.

#### 2. Code Signature Verification

**Test:** After the build completes, run:
```
codesign -dv build/bin/LamboServer.app 2>&1 | grep "Signature=adhoc"
codesign --verify --deep --strict build/bin/LamboServer.app
codesign -d --entitlements - build/bin/LamboServer.app 2>&1 | grep "allow-jit"
```
**Expected:** First command shows `Signature=adhoc`; second exits 0; third shows `allow-jit` confirming entitlements are embedded (not empty).
**Why human:** Requires a built .app produced by the pipeline; the build artifact is gitignored and not present in the repository at verification time.

#### 3. DMG Layout in Finder (Task 3 — blocking checkpoint from Plan 02)

**Test:** Open Finder, navigate to the project root, and double-click `LamboServer-1.0.0.dmg` to mount it.
**Expected:** DMG window opens showing (a) `LamboServer.app` with the custom icon from Phase 7, and (b) an `Applications` folder shortcut with a visible alias arrow. Dragging the app to the Applications shortcut completes installation. Eject the volume afterwards.
**Why human:** DMG window layout is visual behaviour in Finder; the drag-to-Applications interaction cannot be verified programmatically. This was designated as a blocking human-verify checkpoint in Plan 02 Task 3.

### Gaps Summary

No blocking gaps found. All 6 phase 9 artifacts exist, are substantive (non-stub), and are wired correctly. The 3 items routed to human verification are runtime/visual checks that are architecturally correct in the code but require executing the build pipeline to confirm end-to-end.

The build script's static analysis is complete — all pipeline steps, flags, and output paths match the plan requirements exactly. Human verification confirms whether the assembled pipeline produces a working distributable.

---

_Verified: 2026-04-17T11:00:00Z_
_Verifier: Claude (gsd-verifier)_
