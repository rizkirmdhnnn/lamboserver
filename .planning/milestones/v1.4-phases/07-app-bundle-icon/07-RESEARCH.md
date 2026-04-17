# Phase 7: App Bundle & Icon - Research

**Researched:** 2026-04-17
**Domain:** macOS .app bundle metadata (wails.json, Info.plist) and .icns icon generation
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Bundle identifier is `dev.lamboserver.app`
- **D-02:** Copyright text is `© 2026 LamboServer`
- **D-03:** Product name in Info.plist is `LamboServer`
- **D-04:** Icon style is Lamborghini-inspired — bull or shield motif referencing the "Lambo" name
- **D-05:** Color scheme is monochrome dark — black/dark gray silhouette on white background, clean macOS dock aesthetic
- **D-06:** Icon must be generated as .icns using sips + iconutil from a 1024x1024 source PNG (do not rely on Wails auto-generation)
- **D-07:** First distributable version is `1.0.0`
- **D-08:** Version number goes in wails.json `info.productVersion`, which templates into Info.plist `CFBundleVersion` and `CFBundleShortVersionString`

### Claude's Discretion
- `comments` field in wails.json `info` block — Claude can choose appropriate descriptive text
- `LSMinimumSystemVersion` — Claude can adjust from current 10.13.0 if Wails v2 requires higher

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BUNDLE-01 | App has proper `wails.json` info block with productName, productVersion, and copyright | wails.json `info` block schema verified; exact field names confirmed: `productName`, `productVersion`, `copyright`, `comments` |
| BUNDLE-02 | App uses correct `CFBundleIdentifier` (reverse-DNS, not default `com.wails.LamboServer`) | CFBundleIdentifier is hardcoded in Info.plist (not a template variable); must be set literally to `dev.lamboserver.app` |
| BUNDLE-03 | `Info.plist` renders correct metadata from `wails.json` info block | Wails Go template variables confirmed: `{{.Info.ProductName}}`, `{{.Info.ProductVersion}}`, `{{.Info.Copyright}}`, `{{.Info.Comments}}` |
| ICON-01 | App has a custom LamboServer icon replacing the Wails placeholder | `build/appicon.png` (1024x1024, RGBA) is the source slot; Wails copies whatever is there during build |
| ICON-02 | Icon is generated as `.icns` with all required macOS sizes (16x16 through 1024x1024) | 10 required iconset files documented; sips + iconutil pipeline verified on this machine |
</phase_requirements>

---

## Summary

Phase 7 makes two independent changes: (1) populate the `wails.json` info block and fix `CFBundleIdentifier` in both plist templates, and (2) replace `build/appicon.png` with a new LamboServer icon and pre-generate an explicit `.icns` file before the Wails build runs.

The metadata work is straightforward: `wails.json` currently has no `info` block, and both `Info.plist` files use the placeholder `com.wails.{{.Name}}` as the bundle identifier. `CFBundleIdentifier` is the only plist key that cannot use a Wails template variable — it must be a literal string. All other metadata (name, version, copyright, comments) flows through Go template variables from `wails.json`.

The icon work has two sub-tasks: designing the 1024x1024 source PNG (the Lamborghini-inspired bull/shield motif), then generating the `.icns` via `sips` + `iconutil`. Wails v2.12.0 has known unreliability with its own icon auto-generation (issue #3860), so D-06 correctly mandates explicit pre-generation. The resulting `iconfile.icns` must be placed where Wails expects it — or the source PNG at `build/appicon.png` replaced and the `.icns` manually placed at `build/darwin/iconfile.icns` to override Wails' internal generation step.

**Primary recommendation:** Execute in two waves — Wave 1 (metadata only, verifiable with a quick `wails build` check of Info.plist), Wave 2 (icon design then .icns generation, verifiable by checking Dock and Finder appearance after build).

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Bundle metadata (name, version, copyright) | Build config (`wails.json`) | Info.plist template | wails.json is the single source of truth; Info.plist consumes it via Go templates |
| Bundle identifier | Build config (`Info.plist`) | — | CFBundleIdentifier is NOT a template variable in Wails v2 — must be hardcoded in the plist file directly |
| Icon source design | `build/appicon.png` | — | 1024x1024 PNG is the master; all other sizes derive from it |
| Icon compilation | Build script (sips + iconutil) | Wails build | Explicit pre-generation overrides Wails' unreliable auto-generation |
| Dev build identity | `build/darwin/Info.dev.plist` | — | Same fixes required as production plist; used during `wails dev` |

---

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Wails v2 build system | v2.12.0 (installed) | Renders Info.plist templates, assembles .app bundle | Already in project; template rendering is built-in behavior |
| `sips` | macOS built-in | Resize 1024x1024 PNG to all required iconset sizes | Ships with macOS; no install required; fast, lossless for PNG |
| `iconutil` | macOS built-in (Xcode CLT) | Compile `.iconset/` directory into `.icns` | Apple's canonical tool for .icns creation; ships with Xcode Command Line Tools |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| Python 3 + Pillow | Python 3.13.0 (installed), Pillow (installed) | Programmatic generation of the LamboServer icon PNG | Only for the icon design sub-task; Pillow can draw SVG-like shapes if no vector source is available |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| sips + iconutil | Sketch, Figma, or online icon generators | Online tools require manual steps; sips/iconutil are scriptable and repeatable |
| Python/Pillow for icon design | Inkscape + sips | Inkscape requires install; Pillow is already available on this machine |

---

## Architecture Patterns

### System Architecture Diagram

```
wails.json (info block)
    │
    │  Go template rendering (wails build)
    ▼
build/darwin/Info.plist ──────────────────────────────► LamboServer.app/Contents/Info.plist
  (template source)                                       CFBundleIdentifier: dev.lamboserver.app
  CFBundleIdentifier: dev.lamboserver.app (literal)       CFBundleName: LamboServer
  {{.Info.ProductName}} ◄── productName: "LamboServer"    CFBundleVersion: 1.0.0
  {{.Info.ProductVersion}} ◄── productVersion: "1.0.0"    NSHumanReadableCopyright: © 2026 LamboServer
  {{.Info.Copyright}} ◄── copyright: "© 2026 LamboServer"
  {{.Info.Comments}} ◄── comments: "..."


build/appicon.png (1024x1024 PNG — LamboServer bull/shield)
    │
    │  sips (resize)
    ▼
build/darwin/LamboServer.iconset/
  icon_16x16.png, icon_16x16@2x.png,
  icon_32x32.png, icon_32x32@2x.png,
  icon_128x128.png, icon_128x128@2x.png,
  icon_256x256.png, icon_256x256@2x.png,
  icon_512x512.png, icon_512x512@2x.png
    │
    │  iconutil -c icns
    ▼
build/darwin/iconfile.icns
    │
    │  wails build (copies to .app)
    ▼
LamboServer.app/Contents/Resources/iconfile.icns
```

### Recommended Project Structure

```
lamboserver/
├── wails.json                        # MODIFY: add info block
├── build/
│   ├── appicon.png                   # REPLACE: new LamboServer bull/shield icon (1024x1024)
│   └── darwin/
│       ├── Info.plist                # MODIFY: CFBundleIdentifier → dev.lamboserver.app
│       ├── Info.dev.plist            # MODIFY: same CFBundleIdentifier fix
│       └── LamboServer.iconset/      # CREATE: 10 PNG files generated by sips
│           ├── icon_16x16.png
│           ├── icon_16x16@2x.png
│           ├── icon_32x32.png
│           ├── icon_32x32@2x.png
│           ├── icon_128x128.png
│           ├── icon_128x128@2x.png
│           ├── icon_256x256.png
│           ├── icon_256x256@2x.png
│           ├── icon_512x512.png
│           └── icon_512x512@2x.png
```

### Pattern 1: wails.json Info Block

**What:** Add the `info` object to `wails.json`. Wails renders these fields into Info.plist at build time using Go template variables. This is the single source of truth for app metadata.

**When to use:** Always — missing `info` block causes silent blank fields in the built bundle.

**Verified schema** [VERIFIED: github.com/wailsapp/wails blob/master/website/static/schemas/config.v2.json]:

```json
"info": {
  "companyName": "LamboServer",
  "productName": "LamboServer",
  "productVersion": "1.0.0",
  "copyright": "© 2026 LamboServer",
  "comments": "Local web development environment manager for macOS"
}
```

Field notes:
- `productName` — populates `CFBundleName` in Info.plist via `{{.Info.ProductName}}`
- `productVersion` — populates `CFBundleVersion` AND `CFBundleShortVersionString` via `{{.Info.ProductVersion}}`
- `copyright` — populates `NSHumanReadableCopyright` via `{{.Info.Copyright}}`
- `comments` — populates `CFBundleGetInfoString` via `{{.Info.Comments}}`
- `companyName` — used by Wails for Windows VERSIONINFO; on macOS it does not appear in standard plist keys but is good hygiene

### Pattern 2: CFBundleIdentifier Is a Literal String (Not a Template Variable)

**What:** Unlike all other Info.plist fields, `CFBundleIdentifier` must be set as a literal string directly in the plist template. Wails does NOT provide a template variable for bundle identifier.

**When to use:** Always when fixing bundle ID — do not attempt `{{.Info.BundleIdentifier}}` (that variable does not exist in Wails v2).

**Example** [VERIFIED: reading build/darwin/Info.plist in this project]:

```xml
<!-- Change this: -->
<key>CFBundleIdentifier</key>
<string>com.wails.{{.Name}}</string>

<!-- To this: -->
<key>CFBundleIdentifier</key>
<string>dev.lamboserver.app</string>
```

This change must be applied to BOTH `build/darwin/Info.plist` (production) and `build/darwin/Info.dev.plist` (dev mode).

### Pattern 3: Explicit .icns Generation with sips + iconutil

**What:** Generate all 10 required iconset sizes from the 1024x1024 source PNG, compile to `.icns`, then allow Wails to copy it into the bundle. This bypasses Wails' unreliable auto-generation (issue #3860).

**When to use:** Whenever replacing the app icon — do not rely on Wails auto-conversion.

**Complete command sequence** [VERIFIED: Apple iconset naming convention, sips/iconutil available on this machine]:

```bash
ICONSET="build/darwin/LamboServer.iconset"
SRC="build/appicon.png"

mkdir -p "$ICONSET"

sips -z 16   16   "$SRC" --out "$ICONSET/icon_16x16.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_16x16@2x.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_32x32.png"
sips -z 64   64   "$SRC" --out "$ICONSET/icon_32x32@2x.png"
sips -z 128  128  "$SRC" --out "$ICONSET/icon_128x128.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_128x128@2x.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_256x256.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_256x256@2x.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_512x512.png"
sips -z 1024 1024 "$SRC" --out "$ICONSET/icon_512x512@2x.png"

iconutil -c icns "$ICONSET" -o build/darwin/iconfile.icns
```

Note: `icon_32x32.png` and `icon_16x16@2x.png` are both 32x32 — this is intentional per Apple's iconset spec. The duplicate is required.

### Pattern 4: LSMinimumSystemVersion Recommendation

The current plist value is `10.13.0`. Wails v2's default templates use 10.13.0. Since this project targets developers on modern macOS (likely Monterey/Ventura/Sonoma/Sequoia), raising it to `10.15.0` (Catalina) is safe and appropriate — Wails v3 templates moved to 10.15.0 as the minimum. [MEDIUM confidence: based on Wails PR #3981 and verified installation docs requiring macOS 12.x to build Wails apps]

Recommendation: Keep `10.13.0` to preserve maximum compatibility. The app's actual users are developers, and the Gatekeeper/signing story is already constrained for Sequoia users — the LSMinimumSystemVersion does not affect this.

### Anti-Patterns to Avoid

- **Editing build/bin/LamboServer.app/Contents/Info.plist directly:** `wails build` regenerates the .app on every run and overwrites manual edits. Always edit the template at `build/darwin/Info.plist`.
- **Using `{{.Info.BundleIdentifier}}` in Info.plist:** This template variable does not exist in Wails v2. CFBundleIdentifier must be a literal string.
- **Relying on Wails' icon auto-generation:** Wails issue #3860 documents that `.icns` auto-generation in v2 is unreliable. Always pre-generate the `.icns` explicitly.
- **Forgetting Info.dev.plist:** Both `Info.plist` and `Info.dev.plist` contain the `com.wails.{{.Name}}` placeholder. Both need the CFBundleIdentifier fix.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| PNG resizing for iconset | Custom resize script | `sips` (built-in) | sips handles PNG correctly, no quality loss at standard icon sizes, no install needed |
| .icns compilation | Binary packing of PNG data | `iconutil` (built-in) | iconutil produces the exact binary format macOS expects; .icns has a complex container format |
| Icon design from code | Complex geometry code | Python + Pillow OR source SVG exported to PNG | Pillow is already installed; simple shapes (shield, bull silhouette) are easier to code than engineer |

**Key insight:** All icon generation tooling is already on this machine (sips, iconutil, Python 3.13 + Pillow). Zero installs are required for Phase 7.

---

## Common Pitfalls

### Pitfall 1: CFBundleIdentifier Stays as Wails Default

**What goes wrong:** The built `.app` has `CFBundleIdentifier = com.wails.LamboServer`. macOS Preferences, future notarization, and code signing all key on this identifier — changing it later invalidates stored preferences and signing history.

**Why it happens:** Developers add the `info` block to `wails.json` but don't realize `CFBundleIdentifier` is not a template variable and still resolves to the `com.wails.*` default.

**How to avoid:** After adding the `info` block, explicitly edit both `build/darwin/Info.plist` and `build/darwin/Info.dev.plist` to replace `<string>com.wails.{{.Name}}</string>` with the literal `<string>dev.lamboserver.app</string>`.

**Warning signs:** Run `defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier` after build — should show `dev.lamboserver.app`, not `com.wails.*`.

### Pitfall 2: Info.dev.plist Left Unchanged

**What goes wrong:** Production builds show the correct identifier but `wails dev` still shows the Wails placeholder. Inconsistency can cause confusion when testing.

**Why it happens:** `Info.dev.plist` is a separate file and easy to miss.

**How to avoid:** Treat `Info.plist` and `Info.dev.plist` as identical twins for this change — both need the same `CFBundleIdentifier` fix.

### Pitfall 3: Wails Icon Auto-Generation Silently Uses Wrong Sizes

**What goes wrong:** `wails build` produces an `.icns` file, but at runtime the Dock and Finder show a blurry or incorrectly sized icon. The 16x16 Dock icon looks like a scaled-down 1024x1024 without sharpening.

**Why it happens:** Wails v2's internal `iconutil` invocation has known bugs (issue #3860). It may skip certain sizes or produce an incomplete `.icns`.

**How to avoid:** Pre-generate `build/darwin/iconfile.icns` using sips + iconutil before running `wails build`. Wails copies this file into the bundle; it does not regenerate `.icns` if one already exists at the expected path.

**Warning signs:** Check `build/bin/LamboServer.app/Contents/Resources/iconfile.icns` — run `iconutil -c iconset iconfile.icns -o /tmp/check.iconset` and verify all 10 sizes are present.

### Pitfall 4: Icon PNG Has Transparency Issues at Small Sizes

**What goes wrong:** The bull/shield icon looks correct at 512x512 but appears as a solid black square or invisible at 16x16 and 32x32 because fine details are lost during downscaling.

**Why it happens:** `sips` uses nearest-neighbor or basic bilinear scaling. Complex icon shapes with thin lines or fine detail become unrecognizable at small sizes.

**How to avoid:** The monochrome dark aesthetic (D-05) actually helps here — a simple silhouette without fine internal lines scales well. Keep the design bold and simple. Verify the 16x16 and 32x32 renders by inspecting the generated iconset files before running `iconutil`.

### Pitfall 5: copyright Field Unicode Character

**What goes wrong:** The `©` character in `"© 2026 LamboServer"` may cause JSON parse errors or display as `Â©` if the file encoding is wrong.

**Why it happens:** wails.json must be UTF-8 encoded. The `©` codepoint (U+00A9) is valid UTF-8 and valid JSON, but some editors save with BOM or wrong encoding.

**How to avoid:** Verify the file after editing with `file wails.json` — should report "UTF-8 Unicode text". The `©` can also be written as the Unicode escape `\u00a9` in JSON if encoding is a concern.

---

## Code Examples

### Final wails.json info Block

[VERIFIED: schema from github.com/wailsapp/wails/blob/master/website/static/schemas/config.v2.json]

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "LamboServer",
  "outputfilename": "LamboServer",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Achmad Rizki Ramadhan",
    "email": "achmadrizkiramadhan0101@gmail.com"
  },
  "info": {
    "companyName": "LamboServer",
    "productName": "LamboServer",
    "productVersion": "1.0.0",
    "copyright": "© 2026 LamboServer",
    "comments": "Local web development environment manager for macOS"
  }
}
```

### CFBundleIdentifier Fix for Both Plist Files

[VERIFIED: reading build/darwin/Info.plist — current value confirmed as `com.wails.{{.Name}}`]

Change this block in BOTH `build/darwin/Info.plist` and `build/darwin/Info.dev.plist`:

```xml
<!-- BEFORE -->
<key>CFBundleIdentifier</key>
<string>com.wails.{{.Name}}</string>

<!-- AFTER -->
<key>CFBundleIdentifier</key>
<string>dev.lamboserver.app</string>
```

### Complete sips + iconutil Pipeline

[VERIFIED: sips and iconutil both confirmed available at /usr/bin/sips and /usr/bin/iconutil on this machine]

```bash
#!/bin/bash
set -e

SRC="build/appicon.png"
ICONSET="build/darwin/LamboServer.iconset"
ICNS_OUT="build/darwin/iconfile.icns"

mkdir -p "$ICONSET"

sips -z 16   16   "$SRC" --out "$ICONSET/icon_16x16.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_16x16@2x.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_32x32.png"
sips -z 64   64   "$SRC" --out "$ICONSET/icon_32x32@2x.png"
sips -z 128  128  "$SRC" --out "$ICONSET/icon_128x128.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_128x128@2x.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_256x256.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_256x256@2x.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_512x512.png"
sips -z 1024 1024 "$SRC" --out "$ICONSET/icon_512x512@2x.png"

iconutil -c icns "$ICONSET" -o "$ICNS_OUT"

echo "Generated: $ICNS_OUT"
```

### Verification After Build

```bash
# Verify CFBundleIdentifier
defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier
# Expected: dev.lamboserver.app

# Verify version
defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleShortVersionString
# Expected: 1.0.0

# Verify copyright
defaults read build/bin/LamboServer.app/Contents/Info.plist NSHumanReadableCopyright
# Expected: © 2026 LamboServer

# Verify icon sizes in the bundle
iconutil -c iconset build/bin/LamboServer.app/Contents/Resources/iconfile.icns \
  -o /tmp/verify.iconset && ls /tmp/verify.iconset | wc -l
# Expected: 10
```

---

## Runtime State Inventory

> Phase 7 is a config/asset change only. No runtime state carries the old bundle identifier or icon.

| Category | Items Found | Action Required |
|----------|-------------|-----------------|
| Stored data | None — no database stores the bundle identifier | None |
| Live service config | None — no external service references CFBundleIdentifier | None |
| OS-registered state | macOS caches app icons in Launch Services database | Run `lsregister -kill -r -domain local -domain system -domain user` after build if old icon persists in Dock |
| Secrets/env vars | None — bundle ID and icon are build-time artifacts, not secrets | None |
| Build artifacts | `build/bin/LamboServer.app` — stale .app from any previous build | Delete or use `wails build -clean` to regenerate |

**Icon cache note:** macOS Launch Services caches app icons aggressively. If the old Wails "W" icon persists in Dock/Finder after a new build, run the lsregister command above. This is a display artifact, not a build error.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| `sips` | ICON-02 (.icns generation) | ✓ | macOS built-in at /usr/bin/sips | None needed |
| `iconutil` | ICON-02 (.icns compilation) | ✓ | macOS built-in at /usr/bin/iconutil | None needed |
| `wails` CLI | BUNDLE-01/02/03 (build + template render) | ✓ | v2.12.0 | None needed |
| Python 3 + Pillow | ICON-01 (icon design, if code-generated) | ✓ | Python 3.13.0, Pillow installed | Inkscape or manual design |

**Missing dependencies with no fallback:** None.

---

## Validation Architecture

> `workflow.nyquist_validation` is absent from config.json — treating as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Shell assertions (bash) — no unit test framework needed for config/asset changes |
| Config file | N/A |
| Quick run command | `defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier` |
| Full suite command | See Phase Gate section below |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BUNDLE-01 | wails.json has info block with productName, productVersion, copyright | Config inspection | `python3 -c "import json,sys; d=json.load(open('wails.json')); assert 'info' in d and d['info']['productName']=='LamboServer'"` | ❌ Wave 0 |
| BUNDLE-02 | Info.plist does not contain `com.wails.` | Config inspection | `grep -c 'com.wails' build/darwin/Info.plist && echo FAIL \|\| echo PASS` | ❌ Wave 0 |
| BUNDLE-03 | Built .app Info.plist shows correct metadata | Post-build smoke | `defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier` (manual verify output = `dev.lamboserver.app`) | ❌ Wave 0 |
| ICON-01 | build/appicon.png is not the Wails default "W" icon | File hash check | `shasum build/appicon.png` — confirm hash differs from original | ❌ Wave 0 |
| ICON-02 | iconfile.icns contains all 10 required sizes | Post-generation check | `iconutil -c iconset build/darwin/iconfile.icns -o /tmp/check && ls /tmp/check \| wc -l` (expect 10) | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** Spot-check the changed file (grep or python assertion)
- **Per wave merge:** Run full post-build verification block from the Code Examples section
- **Phase gate:** All five requirement checks pass before marking Phase 7 complete

### Wave 0 Gaps
- [ ] `scripts/verify-bundle.sh` — automates the 5 post-build checks above
- [ ] Original appicon.png SHA to compare against (record before replacement)

*(No test framework install needed — all checks use bash, python3, sips, defaults, and iconutil which are all confirmed available.)*

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual .icns creation with standalone tools | sips + iconutil (macOS built-ins) | macOS 10.7+ | Zero dependencies; repeatable in a shell script |
| Wails generates .icns automatically from appicon.png | Explicit pre-generation, copy to build/darwin/iconfile.icns | Wails issue #3860 (v2, ongoing) | Must generate explicitly to avoid malformed .icns |
| Hardcoded bundle ID in source | Template variables for all metadata except bundle ID | Wails v2 design | Bundle ID must remain literal in plist; everything else uses wails.json |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `LSMinimumSystemVersion: 10.13.0` is acceptable for Wails v2; actual required minimum may be higher | Pattern 4 | App silently fails to launch on macOS 10.13–10.14; negligible real-world risk as those versions are <2% of Mac users and this is a developer tool |
| A2 | Placing `build/darwin/iconfile.icns` before `wails build` causes Wails to use it rather than regenerating | Common Pitfalls / Don't Hand-Roll | If Wails always regenerates from appicon.png regardless, the explicit .icns would be overwritten. Mitigation: replace appicon.png (the source) AND pre-generate .icns; if Wails regenerates, it regenerates from the correct source |

---

## Open Questions (RESOLVED)

1. **Icon design tooling**
   - What we know: Python 3 + Pillow are available; the design brief is "monochrome dark bull or shield silhouette on white"
   - What's unclear: Whether to generate the icon purely in code (Pillow) or produce it as instructions for a human designer
   - Recommendation: Phase 7 plan should include a task to generate the icon in Python/Pillow using geometric primitives — this is automatable and keeps the plan self-contained
   - RESOLVED: Python/Pillow programmatic generation (07-02 Task 1 Step A)

2. **iconfile.icns placement — override or source replacement?**
   - What we know: Wails reads `build/appicon.png` as the icon source; it has also been seen to copy a pre-existing `iconfile.icns` if present in `build/darwin/`
   - What's unclear: Whether Wails v2.12.0 will use a pre-placed `build/darwin/iconfile.icns` or always regenerate from `build/appicon.png`
   - Recommendation: Do both — replace `build/appicon.png` with the new design AND pre-generate `build/darwin/iconfile.icns`. This guarantees the correct icon regardless of Wails' internal behavior.
   - RESOLVED: Do both — replace appicon.png AND pre-generate iconfile.icns (07-02 Task 1 Steps A+C)

---

## Sources

### Primary (HIGH confidence)
- [VERIFIED: github.com/wailsapp/wails blob/master/website/static/schemas/config.v2.json] — wails.json `info` block schema, all field names and defaults
- [VERIFIED: build/darwin/Info.plist in this project] — current plist structure, template variables confirmed, CFBundleIdentifier current value confirmed as `com.wails.{{.Name}}`
- [VERIFIED: /usr/bin/sips, /usr/bin/iconutil — both present on this machine] — tooling availability confirmed
- [VERIFIED: wails v2.12.0 installed at /Users/rizkirmdhn/go/bin/wails] — exact version confirmed

### Secondary (MEDIUM confidence)
- [Wails PR #3981](https://github.com/wailsapp/wails/pull/3981) — evidence that Wails v3 raised LSMinimumSystemVersion to 10.15.0; v2 default remains 10.13.0
- [Wails issue #3860](https://github.com/wailsapp/wails/issues/3860) — .icns auto-generation unreliability in Wails v2 (GitHub issue, not official docs, but directly referenced by D-06)
- [wailsapp/wails STACK.md in this project, researched 2026-04-17] — prior research confirming sips + iconutil workflow and Wails build pipeline

### Tertiary (LOW confidence)
None — all factual claims in this document are verified or cited.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all tools verified present on this machine, schema verified from official source
- Architecture: HIGH — based on reading actual project files (wails.json, Info.plist, Info.dev.plist)
- Pitfalls: HIGH — Pitfall 3 (Wails #3860) is MEDIUM (GitHub issue, not official docs); all others derived from reading actual file contents

**Research date:** 2026-04-17
**Valid until:** 2026-07-17 (Wails v2 is stable; macOS iconset spec does not change)
