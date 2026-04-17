# Architecture Research

**Domain:** macOS installer packaging for Wails v2 desktop app
**Researched:** 2026-04-17
**Confidence:** HIGH (Wails build system), MEDIUM (DMG tooling — community-driven, not Wails-native)

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Source (existing)                           │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │  Go backend │  │ React/TS UI  │  │  build/darwin/       │   │
│  │  (app.go,   │  │ frontend/    │  │  Info.plist (tmpl)   │   │
│  │   pkg/, ...) │  │ dist/        │  │  appicon.png         │   │
│  └──────┬──────┘  └──────┬───────┘  └──────────┬───────────┘   │
│         │                │                      │               │
└─────────┴────────────────┴──────────────────────┴───────────────┘
                           │
                    wails build
                    -platform darwin/universal
                    -clean
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Wails Build Step                            │
│  1. npm install && npm run build  (frontend)                    │
│  2. go build -tags desktop,production -ldflags "-w -s"          │
│  3. Assemble .app bundle:                                        │
│     - Renders Info.plist template → Contents/Info.plist         │
│     - Converts appicon.png → iconfile.icns (iconutil)           │
│     - Places binary at Contents/MacOS/LamboServer               │
│  Output: build/bin/LamboServer.app                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                      (new work begins here)
                              │
          ┌───────────────────┴──────────────────┐
          │                                      │
          ▼                                      ▼
┌──────────────────────┐              ┌──────────────────────────┐
│   Ad-hoc Signing     │              │    DMG Creation          │
│                      │              │                          │
│  codesign -s "-"     │              │  create-dmg (brew)       │
│  --deep              │              │  --volname "LamboServer" │
│  --force             │              │  --icon at 200,190       │
│  --options runtime   │              │  --app-drop-link 600,185 │
│  LamboServer.app     │              │  LamboServer.dmg         │
│                      │              │  staging/                │
│  (no Apple account   │              │                          │
│   required; works    │              │  Output: LamboServer.dmg │
│   locally + internal │              │                          │
│   distribution)      │              └──────────────────────────┘
└──────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Status |
|-----------|----------------|--------|
| `wails.json` | Project metadata, build hooks, frontend commands | EXISTS — needs `info` block filled |
| `build/darwin/Info.plist` | .app bundle identity (bundle ID, version, icon ref) | EXISTS — needs real bundle ID + version |
| `build/darwin/entitlements.plist` | macOS security entitlements for codesign | MISSING — must create |
| `build/appicon.png` | Source icon (currently Wails default "W" logo) | EXISTS — must replace |
| `scripts/build-dmg.sh` | Orchestrates sign + DMG steps post-`wails build` | MISSING — must create |
| `build/bin/LamboServer.app` | Output of `wails build` | NOT YET BUILT |
| `LamboServer.dmg` | Final distributable installer | NOT YET BUILT |

## Recommended Project Structure

```
lamboserver/
├── wails.json                     # MODIFY: add info block (version, copyright)
├── build/
│   ├── appicon.png                # REPLACE: design LamboServer icon (1024x1024)
│   ├── darwin/
│   │   ├── Info.plist             # MODIFY: set real bundle ID, version placeholders
│   │   ├── Info.dev.plist         # MODIFY: same as Info.plist for dev
│   │   └── entitlements.plist     # CREATE: network + subprocess entitlements
│   └── bin/
│       └── LamboServer.app        # OUTPUT of wails build (gitignored)
├── scripts/
│   └── build-dmg.sh              # CREATE: full build + sign + DMG pipeline
└── LamboServer.dmg               # OUTPUT artifact (gitignored)
```

### Structure Rationale

- **`build/darwin/`:** Wails already owns this directory for macOS build assets. All new macOS-specific files go here to stay within Wails conventions.
- **`scripts/build-dmg.sh`:** Isolated from the Go/Wails source tree. This is a developer tool, not app code. Keeps the Makefile-style orchestration separate.
- **`wails.json` `info` block:** Wails templates Info.plist at build time using this data. Setting `productVersion` here propagates to `CFBundleVersion` automatically — no manual plist editing needed.

## Architectural Patterns

### Pattern 1: Wails Template Rendering for Info.plist

**What:** Wails renders `build/darwin/Info.plist` as a Go template at build time. Fields like `{{.Info.ProductName}}`, `{{.Info.ProductVersion}}`, and `{{.OutputFilename}}` are injected from `wails.json`. The existing `Info.plist` already uses this pattern — it just has the wrong bundle identifier (`com.wails.LamboServer`).

**When to use:** Always. Do not hardcode values directly in Info.plist. Use template variables so `wails.json` is the single source of truth.

**Trade-offs:** Template errors are silent (Wails will still build); wrong bundle ID only surfaces when signing.

**What to set in `wails.json`:**
```json
{
  "info": {
    "productName": "LamboServer",
    "productVersion": "1.4.0",
    "copyright": "Copyright © 2026 Achmad Rizki Ramadhan",
    "comments": "Local development environment manager for macOS"
  }
}
```

**What to set in `Info.plist` (bundle identifier key):**
```xml
<key>CFBundleIdentifier</key>
<string>com.rizkirmdhn.lamboserver</string>
```

Note: Bundle identifier is NOT a template variable in Wails v2 — it must be hardcoded in the plist. Confirm this against `build/darwin/Info.plist` before editing.

### Pattern 2: Ad-hoc Code Signing (No Apple Developer Account)

**What:** Sign the `.app` with `-s "-"` (dash = ad-hoc identity). This embeds a checksum, satisfies Gatekeeper's requirement that the binary is signed, and prevents "damaged app" errors on Apple Silicon. It does NOT provide notarization or Developer ID trust.

**When to use:** Internal distribution, developer-to-developer sharing, or any case where Apple Developer Program ($99/year) is not in play. This milestone targets ad-hoc specifically.

**Limitation:** Recipients on other Macs may get a Gatekeeper warning ("can't be opened because Apple cannot check it for malicious software"). They must right-click → Open to bypass it once. This is expected and acceptable for a developer tool.

**Command:**
```bash
codesign --sign "-" \
  --deep \
  --force \
  --options runtime \
  --entitlements ./build/darwin/entitlements.plist \
  ./build/bin/LamboServer.app
```

**Trade-offs:** Simple, no secrets required. Cannot be notarized. Gatekeeper friction for end users.

### Pattern 3: create-dmg for Drag-to-Applications Installer

**What:** `create-dmg` (Homebrew: `brew install create-dmg`) creates a styled `.dmg` with a drag-to-Applications shortcut. It is a shell script wrapper around `hdiutil` and `osascript` that handles window layout, icon positions, and background images.

**When to use:** Whenever producing a distributable DMG. `hdiutil` alone produces a functional but visually bare disk image. `create-dmg` adds the standard UX in 5 lines.

**Trade-offs:** Requires Homebrew. DMG creation takes a few seconds. No background image complexity required — the minimal invocation is sufficient.

**Command:**
```bash
create-dmg \
  --volname "LamboServer" \
  --volicon "./build/appicon.icns" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "LamboServer.app" 200 190 \
  --hide-extension "LamboServer.app" \
  --app-drop-link 600 185 \
  "LamboServer.dmg" \
  "build/bin/"
```

## Data Flow

### Build Pipeline (Full Sequence)

```
Developer runs: ./scripts/build-dmg.sh
      │
      ▼
[Step 1] wails build -platform darwin/universal -clean
      │  - Installs frontend deps (npm install)
      │  - Builds React frontend (npm run build)
      │  - Compiles Go binary (darwin/amd64 + darwin/arm64)
      │  - Renders Info.plist from template + wails.json
      │  - Converts appicon.png → iconfile.icns (via iconutil)
      │  - Assembles LamboServer.app bundle
      │  Output: build/bin/LamboServer.app
      │
      ▼
[Step 2] codesign --sign "-" --deep --force --options runtime \
                  --entitlements build/darwin/entitlements.plist \
                  build/bin/LamboServer.app
      │  - Embeds ad-hoc signature in all binary components
      │  Output: build/bin/LamboServer.app (signed in-place)
      │
      ▼
[Step 3] create-dmg ... LamboServer.dmg build/bin/
      │  - Creates staged disk image with .app + Applications link
      │  - Sets window size, icon positions
      │  Output: LamboServer.dmg (ready to distribute)
      │
      ▼
[Done] Distribute LamboServer.dmg
```

### Key Data Flows

1. **Icon pipeline:** `build/appicon.png` (1024x1024 PNG) → Wails converts to `.icns` via `sips` + `iconutil` at build time → embedded in .app bundle as `iconfile.icns`. No manual `.icns` creation needed IF using `wails build`.

2. **Metadata pipeline:** `wails.json` `info` block → Wails template engine → `build/darwin/Info.plist` rendered → written to `LamboServer.app/Contents/Info.plist`. Version, copyright, and product name flow from a single source.

3. **Entitlements pipeline:** `build/darwin/entitlements.plist` → referenced by `codesign` → embedded in signed binary headers. Entitlements must match what the app actually does (network access, subprocess spawning) or macOS may kill the process at runtime.

## Integration Points

### Wails Build Integration

| Integration Point | What Changes | Notes |
|-------------------|--------------|-------|
| `wails.json` | Add `info` block with version, copyright | Template source for Info.plist |
| `build/darwin/Info.plist` | Set real `CFBundleIdentifier` | Change from `com.wails.LamboServer` to `com.rizkirmdhn.lamboserver` |
| `build/darwin/entitlements.plist` | Create from scratch | Required by codesign; Wails does not generate this |
| `build/appicon.png` | Replace with actual LamboServer icon | Wails auto-converts to .icns during build |
| `wails build` command | Use `-platform darwin/universal` flag | Produces fat binary for Intel + Apple Silicon |

### New vs Modified — Explicit List

**NEW files to create:**
- `build/darwin/entitlements.plist` — codesign entitlements (network, subprocess)
- `scripts/build-dmg.sh` — pipeline script (build + sign + DMG)

**MODIFIED files (existing, need edits):**
- `wails.json` — add `info` block (`productVersion`, `copyright`, `comments`)
- `build/darwin/Info.plist` — change `CFBundleIdentifier` from `com.wails.LamboServer` to project-specific value; add `LSApplicationCategoryType` for App Store categorization (optional but good practice)
- `build/appicon.png` — replace Wails default "W" logo with actual LamboServer icon

**NO changes needed to:**
- `main.go` — Mac options already set; About info already present
- `app.go` — no packaging concerns
- Frontend code — packaging is entirely a build-layer concern
- `build/darwin/Info.dev.plist` — optional; only affects `wails dev` display name

### Entitlements Required by LamboServer

LamboServer spawns child processes (Nginx, PHP, MySQL, PostgreSQL, dnsmasq, pgweb) via `exec.Command` and makes outgoing network connections. The entitlements file must declare:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Required: Wails WebView uses network -->
    <key>com.apple.security.network.client</key>
    <true/>
    <!-- Required: pgweb, phpMyAdmin server processes bind local ports -->
    <key>com.apple.security.network.server</key>
    <true/>
    <!-- Do NOT enable app-sandbox: exec.Command for Homebrew services -->
    <!-- will fail if sandbox is active. LamboServer is a dev tool,    -->
    <!-- not a sandboxed consumer app.                                  -->
</dict>
</plist>
```

**Critical:** Do NOT add `com.apple.security.app-sandbox`. Sandboxing prevents `exec.Command` from spawning Homebrew-managed services (Nginx, PHP-FPM, etc.). The Wails signing guide shows a sandboxed example — that pattern is for Mac App Store, not developer tools.

## Anti-Patterns

### Anti-Pattern 1: Sandboxing a Service Manager

**What people do:** Copy the Wails signing guide's `entitlements.plist` verbatim, which includes `com.apple.security.app-sandbox = true`.

**Why it's wrong:** macOS sandbox blocks `exec.Command` calls to binaries outside the app bundle. LamboServer's entire backend is `exec.Command` calls to Homebrew-installed binaries (`/opt/homebrew/bin/nginx`, etc.). With sandbox enabled, every service start/stop call will silently fail or crash.

**Do this instead:** Use entitlements without sandbox (network client + server only). Ad-hoc signed apps do not require sandboxing.

### Anti-Pattern 2: Hardcoding wails.json Values in Info.plist

**What people do:** Directly edit `build/bin/LamboServer.app/Contents/Info.plist` after `wails build` to change the version or name.

**Why it's wrong:** `wails build` regenerates the .app bundle on every run, overwriting manual changes. Changes to the rendered output are lost.

**Do this instead:** Edit `wails.json` `info` block and `build/darwin/Info.plist` (the template). Wails renders from the template on every build.

### Anti-Pattern 3: Using hdiutil Directly Instead of create-dmg

**What people do:** Call `hdiutil create -volname "LamboServer" -srcfolder build/bin/LamboServer.app -format UDZO LamboServer.dmg` directly.

**Why it's wrong:** This produces a functional but bare DMG with no Applications shortcut. Users don't know to drag-to-Applications. The installer feels unfinished.

**Do this instead:** Use `create-dmg` with `--app-drop-link`. The tool wraps hdiutil and adds the drag-to-install UX in a single command.

### Anti-Pattern 4: Building for Only One Architecture

**What people do:** Run `wails build` without a `-platform` flag, which builds for the host machine's architecture only.

**Why it's wrong:** A build from an Apple Silicon Mac produces an arm64-only binary that won't run on Intel Macs (and vice versa). The DMG would be architecture-specific with no indication to users.

**Do this instead:** Use `wails build -platform darwin/universal`. Wails produces a universal binary (fat binary containing both arm64 and amd64 slices) that runs natively on both chip families.

## Scaling Considerations

This is a macOS desktop app packager. Scaling is not a runtime concern. The relevant "scale" questions are build-time:

| Concern | Current Approach | Future Path |
|---------|-----------------|-------------|
| One developer building locally | `scripts/build-dmg.sh` run manually | Sufficient |
| Multiple developers building | Same script on each machine | Add to Makefile or GitHub Actions |
| Verified distribution (notarization) | Not in this milestone; ad-hoc only | Requires Apple Developer Program ($99/year) + `gon` or `notarytool` |
| Auto-update | Not in scope | Would require Sparkle framework integration |

## Sources

- Wails v2 Code Signing Guide (via Context7): [https://wails.io/docs/guides/signing](https://wails.io/docs/guides/signing)
- Wails v2 CLI Reference (via Context7): [https://wails.io/docs/reference/cli](https://wails.io/docs/reference/cli)
- Wails v2 Project Config (via Context7): [https://wails.io/docs/reference/project-config](https://wails.io/docs/reference/project-config)
- create-dmg GitHub: [https://github.com/create-dmg/create-dmg](https://github.com/create-dmg/create-dmg)
- macOS distribution reference (ad-hoc signing behavior): [https://gist.github.com/rsms/929c9c2fec231f0cf843a1a746a416f5](https://gist.github.com/rsms/929c9c2fec231f0cf843a1a746a416f5)

---
*Architecture research for: macOS installer packaging (Wails v2)*
*Researched: 2026-04-17*
