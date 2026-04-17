# Phase 7: App Bundle & Icon - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Configure wails.json metadata (productName, productVersion, copyright, bundle identifier) and replace the default Wails "W" icon with a custom LamboServer icon in .icns format. This phase produces a correctly identified .app bundle — production build validation and signing happen in Phase 8 and 9.

</domain>

<decisions>
## Implementation Decisions

### Bundle Identity
- **D-01:** Bundle identifier is `dev.lamboserver.app` (product-focused namespace, not personal)
- **D-02:** Copyright text is `© 2026 LamboServer`
- **D-03:** Product name in Info.plist is `LamboServer`

### App Icon
- **D-04:** Icon style is Lamborghini-inspired — bull or shield motif referencing the "Lambo" name
- **D-05:** Color scheme is monochrome dark — black/dark gray silhouette on white background, clean macOS dock aesthetic
- **D-06:** Icon must be generated as .icns using sips + iconutil from a 1024x1024 source PNG (do not rely on Wails auto-generation)

### Version
- **D-07:** First distributable version is `1.0.0` (fresh start, not matching milestone v1.4 numbering)
- **D-08:** Version number goes in wails.json info.productVersion, which templates into Info.plist CFBundleVersion and CFBundleShortVersionString

### Claude's Discretion
- Comments field in wails.json info block — Claude can choose appropriate descriptive text
- LSMinimumSystemVersion — Claude can adjust from current 10.13.0 if Wails v2 requires higher

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Build Configuration
- `wails.json` — Wails project config; needs `info` block added (currently missing)
- `build/darwin/Info.plist` — macOS plist template; `CFBundleIdentifier` must change from `com.wails.{{.Name}}` to `dev.lamboserver.app`
- `build/darwin/Info.dev.plist` — Dev plist template; same bundle ID fix needed

### Current Icon Asset
- `build/appicon.png` — Current Wails "W" placeholder (1024x1024); must be replaced with LamboServer icon

### Research
- `.planning/research/STACK.md` — Tool versions, .icns generation approach (sips + iconutil)
- `.planning/research/ARCHITECTURE.md` — Build pipeline integration, file changes needed
- `.planning/research/PITFALLS.md` — Wails icon auto-gen unreliability (issue #3860)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `build/appicon.png` — 1024x1024 source slot for icon; Wails reads this during build
- `build/darwin/Info.plist` — Template with Go template variables; already wired for `{{.Info.ProductName}}`, `{{.Info.ProductVersion}}`, `{{.Info.Copyright}}`

### Established Patterns
- `wails.json` drives build configuration — the `info` block populates Info.plist template variables
- Wails build places compiled .app in `build/bin/LamboServer.app/`

### Integration Points
- `wails.json` → `Info.plist` template rendering during `wails build`
- `build/appicon.png` → `build/bin/LamboServer.app/Contents/Resources/iconfile.icns` during build
- Bundle identifier appears in Info.plist `CFBundleIdentifier` key

</code_context>

<specifics>
## Specific Ideas

- Icon is Lamborghini-inspired — bull or shield motif, tying to the "Lambo" in LamboServer
- Monochrome dark aesthetic — black/dark gray on white, designed to look clean in the macOS Dock
- Version 1.0.0 signals first official distributable release, independent of internal milestone numbering

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 07-app-bundle-icon*
*Context gathered: 2026-04-17*
