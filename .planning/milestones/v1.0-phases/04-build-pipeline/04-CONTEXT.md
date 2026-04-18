# Phase 4: Build Pipeline - Context

**Gathered:** 2026-04-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Update the build pipeline (CI and local script) to support CGO compilation required by the custom tray package (`internal/tray/`). Verify that Info.plist does not need tray-specific keys. Ensure the CI workflow produces a valid signed universal DMG with Cocoa framework linkage.

</domain>

<decisions>
## Implementation Decisions

### Info.plist
- **D-01:** No `LSUIElement` key added. The Dock icon stays visible alongside the tray icon. Dynamic Dock hiding is deferred to v2 (TRAY-06).
- **D-02:** No new Info.plist keys are needed for tray functionality. NSStatusBar works at runtime via CGO — no plist configuration required. The existing `build/darwin/Info.plist` template is sufficient as-is.
- **D-03:** TRAY-04 (post-build plist patching) is satisfied by verifying the existing plist is correct, not by adding new keys.

### CI Pipeline
- **D-04:** Set `CGO_ENABLED=1` as an environment variable in the GitHub Actions workflow. Use the existing `wails build -platform darwin/universal` command — Wails handles universal binary cross-compilation with CGO internally.
- **D-05:** Add a basic verification step after the build: use `otool -L` and/or `lipo -info` on the built binary to confirm Cocoa.framework linkage and universal architecture. Quick sanity check, no runtime test.
- **D-06:** If `CGO_ENABLED=1` + `darwin/universal` fails in Wails, fall back to separate arm64 + amd64 builds combined with `lipo -create`. Try the simple approach first.

### Local Build Script
- **D-07:** Claude's discretion on how to update `scripts/build-dmg.sh`. At minimum, add `CGO_ENABLED=1` to the wails build command. Claude may add verification or parity with CI as appropriate.

### Claude's Discretion
- Level of CI/local script parity (verification steps in local script)
- Whether to extract version from `wails.json` instead of hardcoding in `scripts/build-dmg.sh`
- Any additional Xcode SDK or CC/CXX environment variables needed for universal CGO builds
- Ordering of CI steps around the CGO build

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Build Configuration
- `build/darwin/Info.plist` — Wails Go template for production builds. No changes needed for tray (D-01, D-02).
- `build/darwin/Info.dev.plist` — Dev-mode plist. Same as production but with `NSAppTransportSecurity` for local networking.
- `build/darwin/entitlements.plist` — macOS entitlements (JIT, unsigned memory, network, Apple Events). No changes needed.
- `wails.json` — Wails project configuration. Contains `productVersion` that could replace hardcoded version in build script.

### CI Pipeline
- `.github/workflows/release.yml` — Current release workflow. Triggers on `v*` tags. Builds universal .app, ad-hoc signs, creates DMG, publishes GitHub Release. Must add `CGO_ENABLED=1` and verification step.

### Local Build Script
- `scripts/build-dmg.sh` — Local DMG build script. Currently calls bare `wails build` without CGO flags. Hardcodes `VERSION="1.0.0"`.

### CGO Tray Package
- `internal/tray/controller_darwin.go` — CGO directives: `#cgo CFLAGS: -x objective-c`, `#cgo LDFLAGS: -framework Cocoa`. These require `CGO_ENABLED=1` at build time.
- `internal/tray/tray_darwin.m` — ObjC implementation. Compiled via CGO.
- `internal/tray/tray_darwin.h` — C header for CGO bridge.

### Phase 1 Context
- `.planning/phases/01-tray-foundation-window-lifecycle/01-CONTEXT.md` — D-01 (custom CGO package decision), D-03 (GCD dispatch).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.github/workflows/release.yml` — Complete CI pipeline already in place. Only needs `CGO_ENABLED=1` env var and a verification step added.
- `scripts/build-dmg.sh` — Local build script with 4-step flow (build, sign, package, done). Step 1 needs CGO flag.
- `build/darwin/entitlements.plist` — Entitlements already include all permissions needed (JIT, unsigned memory, network, Apple Events).

### Established Patterns
- **Ad-hoc signing:** `codesign --force --deep -s -` with entitlements. No Apple Developer identity needed.
- **Universal binary:** `wails build -platform darwin/universal` produces fat binary for arm64+amd64.
- **DMG packaging:** hdiutil-based DMG creation with Applications symlink.

### Integration Points
- `.github/workflows/release.yml` line 31 — `wails build` command: prepend `CGO_ENABLED=1` env var.
- `.github/workflows/release.yml` after line 31 — Insert verification step (otool/lipo check).
- `scripts/build-dmg.sh` line 12 — `wails build` command: prepend `CGO_ENABLED=1`.

</code_context>

<specifics>
## Specific Ideas

- Keep it simple — the tray CGO package is already built and working locally, this phase is about making CI reproduce that
- Verification step is a safety net, not a blocker — if otool shows Cocoa.framework linked and lipo shows universal, the build is valid
- No LSUIElement means no behavioral change to the app bundle — just the build process changes

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-build-pipeline*
*Context gathered: 2026-04-18*
