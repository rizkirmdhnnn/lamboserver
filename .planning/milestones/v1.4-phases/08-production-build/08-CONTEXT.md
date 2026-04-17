# Phase 8: Production Build - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Produce a universal binary .app via `wails build -platform darwin/universal` and validate it launches correctly from build/bin/. This phase ensures the .app is a working production artifact — code signing and DMG packaging happen in Phase 9.

</domain>

<decisions>
## Implementation Decisions

### PATH Fix
- **D-01:** PATH fix for Homebrew binaries is NOT NEEDED — all service binaries (nginx, dnsmasq, mysql, postgres, php, node, pgweb) are self-contained in `~/.lamboserver/` and located via `BinaryLocator.Find()` which checks `LocalPath` first. No Homebrew dependency exists.

### Build Configuration
- **D-02:** Build target is `darwin/universal` (Intel + Apple Silicon universal binary)
- **D-03:** Build command is `wails build -platform darwin/universal -clean`
- **D-04:** No build script needed for this phase — the single wails command is sufficient. A full build pipeline script will be created in Phase 9 (build + sign + DMG).

### Testing
- **D-05:** Validation is build + manual launch test — run `wails build`, then open the .app from `build/bin/LamboServer.app` and verify it launches with the correct icon and metadata
- **D-06:** Verify Info.plist inside the built .app contains correct values from Phase 7 (productName, version 1.0.0, bundle ID dev.lamboserver.app, copyright)

### Claude's Discretion
- Whether to add a Makefile target or keep the build as a direct wails command
- Whether LSMinimumSystemVersion needs adjustment based on build output

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Build Configuration (from Phase 7)
- `wails.json` — Info block with productName, productVersion, copyright (configured in Phase 7)
- `build/darwin/Info.plist` — Bundle ID fixed to `dev.lamboserver.app` (Phase 7)
- `build/appicon.png` — Custom icon (Phase 7)

### Binary Resolution
- `internal/system/binary.go` — BinaryLocator checks LocalPath (~/.lamboserver/) first, then system PATH. No Homebrew paths needed.

### Research
- `.planning/research/ARCHITECTURE.md` — Wails build pipeline details
- `.planning/research/PITFALLS.md` — Build pitfalls (sandboxing, entitlements)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `BinaryLocator` in `internal/system/binary.go` — already handles binary discovery via local paths, no changes needed
- `wails.json` — fully configured with info block from Phase 7

### Established Patterns
- All services use `BinaryLocator` with `LocalPath` pointing to `~/.lamboserver/` subdirectories
- No Homebrew dependency anywhere in the codebase

### Integration Points
- `wails build` reads `wails.json` → produces `.app` in `build/bin/`
- `build/appicon.png` → `.app/Contents/Resources/iconfile.icns` during build
- `build/darwin/Info.plist` template → `.app/Contents/Info.plist` rendered at build time

</code_context>

<specifics>
## Specific Ideas

- The build should be a simple `wails build -platform darwin/universal -clean` command
- Verification: open the .app, check icon in Dock, check Get Info for version/copyright
- Phase 9 will create the full build pipeline script (build + sign + DMG)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-production-build*
*Context gathered: 2026-04-17*
