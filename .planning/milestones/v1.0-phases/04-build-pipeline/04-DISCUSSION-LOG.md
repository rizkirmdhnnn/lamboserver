# Phase 4: Build Pipeline - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-18
**Phase:** 04-build-pipeline
**Areas discussed:** Info.plist patching strategy, CI CGO enablement, Local build script

---

## Info.plist Patching Strategy

### Dock Icon Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Keep Dock icon visible | App shows in both Dock and tray. Matches user expectation for GUI app. TRAY-06 deferred to v2. | ✓ |
| Hide Dock icon (LSUIElement) | App only in tray/menu bar. No Dock icon, no app menu bar. | |
| Add LSUIElement=false | Explicitly declare key as false. Same behavior as omitting, documents intent. | |

**User's choice:** Keep Dock icon visible (Recommended)
**Notes:** Aligns with TRAY-06 being deferred to v2.

### Plist Keys Needed

| Option | Description | Selected |
|--------|-------------|----------|
| No new keys needed | Tray works via CGO/NSStatusBar at runtime — no Info.plist keys required. | ✓ |
| Add NSStatusBarUsageDescription | Usage description for status bar. Not required by macOS. | |
| You decide | Claude decides based on research. | |

**User's choice:** No new keys needed (Recommended)
**Notes:** Simplifies the phase — TRAY-04 becomes verification rather than patching.

---

## CI CGO Enablement

### Build Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| CGO_ENABLED=1 + darwin/universal | Set env var, let Wails handle cross-compilation. Likely works out of the box. | ✓ |
| Separate arm64 + amd64, then lipo | Build each arch separately, combine. More control but more complexity. | |
| arm64 only | Drop universal, ship arm64 only. Drops Intel support. | |

**User's choice:** CGO_ENABLED=1 + darwin/universal (Recommended)
**Notes:** Simple approach first, fall back to separate builds only if needed.

### Build Verification

| Option | Description | Selected |
|--------|-------------|----------|
| Basic verification | otool -L and lipo -info check after build. Quick sanity check. | ✓ |
| No verification | Trust the build, keep CI simple. | |
| You decide | Claude decides approach. | |

**User's choice:** Basic verification (Recommended)
**Notes:** None.

---

## Local Build Script

### CGO Update Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Add CGO_ENABLED=1 env var | Simple: prepend to wails build command. | |
| Full parity with CI | Mirror CI exactly: CGO, verification, same env vars. | |
| You decide | Claude decides level of alignment. | ✓ |

**User's choice:** You decide
**Notes:** Claude has discretion on local script updates.

---

## Claude's Discretion

- Level of CI/local script parity
- Version extraction from wails.json vs hardcoded
- Additional Xcode SDK env vars for universal CGO
- Local build script verification steps

## Deferred Ideas

None — discussion stayed within phase scope
