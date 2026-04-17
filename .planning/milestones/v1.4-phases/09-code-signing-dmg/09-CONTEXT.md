# Phase 9: Code Signing & DMG - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Ad-hoc sign the LamboServer.app bundle, package it into a DMG installer with drag-to-Applications layout, and document the Gatekeeper workaround for macOS Sequoia users. This is the final phase of the v1.4 macOS Installer milestone.

</domain>

<decisions>
## Implementation Decisions

### Build Script
- **D-01:** Build pipeline is a shell script at `scripts/build-dmg.sh` (matches existing `scripts/generate-icns.sh` pattern)
- **D-02:** Pipeline steps: `wails build -platform darwin/universal -clean` -> `codesign -s -` with entitlements -> `hdiutil` to create DMG
- **D-03:** Script does NOT include icon generation step -- iconfile.icns is already committed and built. Script starts at wails build.

### Code Signing
- **D-04:** Ad-hoc signing with `codesign -s -` (no Apple Developer account -- project constraint)
- **D-05:** Entitlements applied from `build/darwin/entitlements.plist` (already exists from Phase 8 with JIT, unsigned memory, library validation, Apple Events, network client/server)
- **D-06:** No Hardened Runtime, no App Sandbox (project constraints -- blocks exec.Command calls)

### DMG Packaging
- **D-07:** DMG created with plain `hdiutil` (macOS built-in, no external dependencies)
- **D-08:** DMG filename is `LamboServer-1.0.0.dmg` (version from wails.json)
- **D-09:** DMG contents: LamboServer.app + Applications folder alias only. No README, no LICENSE inside the DMG.

### Documentation
- **D-10:** Gatekeeper workaround instructions go in GitHub release notes only (not in README.md or inside DMG)
- **D-11:** Build script only outputs the DMG -- no release notes template generation

### Claude's Discretion
- Exact hdiutil flags and volume name for DMG creation
- codesign flags beyond `-s -` and `--entitlements` (e.g., `--deep`, `--force`)
- Whether to verify the signature as part of the build script (codesign --verify)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Build Configuration (from Phase 7 & 8)
- `wails.json` -- Info block with productName, productVersion 1.0.0, copyright, bundle ID
- `build/darwin/Info.plist` -- Bundle ID `dev.lamboserver.app` (configured in Phase 7)
- `build/darwin/entitlements.plist` -- Entitlements for ad-hoc signing (JIT, unsigned memory, library validation, Apple Events, network)

### Existing Scripts
- `scripts/generate-icns.sh` -- Pattern for shell scripts in this project (set -e, simple sequential steps)

### Requirements
- `.planning/REQUIREMENTS.md` -- SIGN-01, SIGN-02, SIGN-03, DMG-01, DMG-02, DOC-01

### Research
- `.planning/research/PITFALLS.md` -- Build and signing pitfalls
- `.planning/research/STACK.md` -- Tool versions and build approach

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `build/darwin/entitlements.plist` -- Complete entitlements file ready for codesign --entitlements
- `scripts/generate-icns.sh` -- Shell script pattern to follow for build-dmg.sh
- `build/bin/LamboServer.app` -- Built .app from Phase 8, ready for signing

### Established Patterns
- Shell scripts in `scripts/` directory with `set -e` error handling
- Build output goes to `build/bin/`
- Version 1.0.0 set in wails.json info.productVersion

### Integration Points
- `wails build` produces `.app` in `build/bin/LamboServer.app`
- `codesign` signs the `.app` in place
- `hdiutil` creates DMG from a staging folder containing .app + Applications alias
- DMG output alongside .app or in project root

</code_context>

<specifics>
## Specific Ideas

- DMG should have a clean, standard macOS layout -- just the app and Applications shortcut
- Version number in DMG filename (LamboServer-1.0.0.dmg) for clear release identification
- Single script runs the entire pipeline with no manual steps

</specifics>

<deferred>
## Deferred Ideas

None -- discussion stayed within phase scope

</deferred>

---

*Phase: 09-code-signing-dmg*
*Context gathered: 2026-04-17*
