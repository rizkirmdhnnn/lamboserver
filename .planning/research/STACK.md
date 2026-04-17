# Stack Research

**Domain:** macOS desktop app packaging — .app bundle, DMG installer, ad-hoc signing, .icns icon
**Researched:** 2026-04-17
**Confidence:** HIGH (macOS tooling is stable; Wails v2 integration verified against v2.12.0 docs)

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Wails v2 build system | v2.12.0 (existing) | Compiles Go + React into `.app` bundle at `build/bin/LamboServer.app` | Already in the project. `wails build -platform darwin/universal` produces a universal binary (Apple Silicon + Intel) with `.app` structure. No additional packaging tool needed for the bundle itself. |
| `codesign` (macOS built-in) | Ships with Xcode CLT | Ad-hoc sign the `.app` so macOS does not block local execution | Built into every Mac with Xcode Command Line Tools. Ad-hoc flag `-s -` requires no Apple account. No installation required. |
| `create-dmg` (shell script) | v1.2.3 | Wrap the signed `.app` in a drag-to-Applications DMG | Homebrew-installable shell script with zero Node.js or npm dependency. Produces professional-looking DMGs (`--app-drop-link`, background image, icon layout) using only stock macOS tools (`hdiutil`, `SetFile`). Actively maintained, latest release Nov 2025. |
| `sips` + `iconutil` (macOS built-in) | Ships with macOS | Generate `.icns` from a source PNG | `sips` resizes the 1024x1024 `build/appicon.png` into the 10 required sizes; `iconutil` compiles the `.iconset` folder into `.icns`. Both ship with macOS — no install required. |

### Supporting Tools

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| Xcode Command Line Tools | Current (via `xcode-select --install`) | Provides `codesign`, `sips`, `iconutil`, `hdiutil` | Must be installed before any signing or icon work. Already present on most dev machines. |
| `create-dmg` (Homebrew) | v1.2.3 | DMG creation with layout | Install once via `brew install create-dmg`. Used in the packaging script. |

### What Wails v2.12.0 Does Automatically

Understanding this prevents double-work:

- `wails build` compiles the Go binary, bundles the React frontend, creates `build/bin/LamboServer.app` with `Contents/MacOS/`, `Contents/Resources/`, `Contents/Info.plist`.
- `build/darwin/Info.plist` is the template; Wails fills in `{{.Info.ProductName}}`, `{{.Info.ProductVersion}}`, `{{.Info.Copyright}}` from `wails.json`'s `info` block.
- `build/appicon.png` (1024x1024, already exists) is used by Wails to generate a `.icns` at `Contents/Resources/iconfile.icns` — BUT this auto-generation is unreliable in v2 (tracked in issue #3860). Explicitly generating the `.icns` before building is safer and ensures correctness.
- Wails does NOT create a DMG. Post-build DMG creation is entirely external.
- Wails does NOT sign the app. Code signing is entirely external.

---

## Configuration Changes Required

### wails.json — Add `info` block

The existing `wails.json` has no `info` block. Add it:

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
    "companyName": "Achmad Rizki Ramadhan",
    "productName": "LamboServer",
    "productVersion": "1.4.0",
    "copyright": "Copyright © 2026 Achmad Rizki Ramadhan",
    "comments": "Local web development infrastructure manager"
  }
}
```

### build/darwin/Info.plist — Fix CFBundleIdentifier

Current template uses `com.wails.{{.Name}}` which resolves to `com.wails.LamboServer`. Change to a proper reverse-domain identifier:

```xml
<key>CFBundleIdentifier</key>
<string>com.lamboserver.app</string>
```

(The template variable `{{.Name}}` should be replaced with a hardcoded identifier to avoid Wails substituting an undesired value.)

---

## Build Script Design

The entire packaging pipeline fits in a shell script `scripts/build-macos.sh`. No Makefile or CI system required for v1.4:

```bash
#!/bin/bash
set -e

APP_NAME="LamboServer"
VERSION="1.4.0"

# 1. Generate .icns (explicit, don't rely on Wails auto-generation)
mkdir -p build/darwin/${APP_NAME}.iconset
sips -z 16 16     build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_16x16.png
sips -z 32 32     build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_16x16@2x.png
sips -z 32 32     build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_32x32.png
sips -z 64 64     build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_32x32@2x.png
sips -z 128 128   build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_128x128.png
sips -z 256 256   build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_128x128@2x.png
sips -z 256 256   build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_256x256.png
sips -z 512 512   build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_256x256@2x.png
sips -z 512 512   build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_512x512.png
sips -z 1024 1024 build/appicon.png --out build/darwin/${APP_NAME}.iconset/icon_512x512@2x.png
iconutil -c icns build/darwin/${APP_NAME}.iconset -o build/darwin/iconfile.icns

# 2. Build universal .app
wails build -platform darwin/universal -clean

# 3. Ad-hoc sign (required for macOS to allow execution)
codesign --force --deep -s - "build/bin/${APP_NAME}.app"

# 4. Create DMG
create-dmg \
  --volname "${APP_NAME}" \
  --volicon "build/darwin/iconfile.icns" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --icon "${APP_NAME}.app" 175 190 \
  --hide-extension "${APP_NAME}.app" \
  --app-drop-link 425 190 \
  "build/bin/${APP_NAME}-${VERSION}.dmg" \
  "build/bin/${APP_NAME}.app"
```

**Note on `--deep` flag in step 3:** Using `--deep` for ad-hoc local signing is acceptable here because LamboServer has no embedded frameworks to sign separately. Apple's guidance against `--deep` applies to Developer ID / notarized distributions where each component needs individual signing. For ad-hoc signing of a simple Wails app, `--deep` is fine.

---

## Ad-Hoc Signing: Honest Limitations

**What ad-hoc signing (`-s -`) actually provides:**
- A checksum-based seal on the bundle — macOS tracks binary integrity
- Required for certain macOS APIs to function (notably on Apple Silicon)
- Works without any Apple account

**What it does NOT provide:**
- Trust on other Macs — Gatekeeper will block the app when received by another user
- Notarization — Apple's malware scan is not performed
- Distribution identity — no developer name shown in security dialogs

**User workaround for other Macs:** Recipients must right-click → Open in Finder (not double-click) to bypass Gatekeeper the first time. This is a one-time action per install. This is acceptable for developer tools distributed to other developers who understand macOS security.

**Bottom line:** Ad-hoc is the correct choice given no Apple Developer account. It is appropriate for this milestone's goal of "packaged app" — it is not appropriate for Mac App Store or seamless end-user distribution.

---

## Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `create-dmg` (shell script) | `sindresorhus/create-dmg` (npm) | Requires Node.js 20+; adds a runtime dependency for what is a shell operation. The shell script version has identical output, no npm dependency. |
| `create-dmg` (shell script) | `appdmg` (npm) | JSON config approach is more complex; Node.js dependency; less actively maintained. |
| `codesign -s -` (ad-hoc) | `gon` (Bearer/tap) | `gon` requires Apple Developer account credentials for notarization and proper signing. Overkill and non-functional without an account. |
| `codesign -s -` (ad-hoc) | No signing at all | Unsigned apps are blocked by macOS immediately on Apple Silicon without workarounds. Ad-hoc at minimum is needed. |
| `sips` + `iconutil` | External icon conversion tools | `sips` and `iconutil` are built into macOS. No additional installs needed. The 1024x1024 `appicon.png` already exists in `build/`. |
| Custom `build/darwin/Info.plist` | Rely on Wails default Info.plist | The existing template uses `com.wails.LamboServer` as bundle identifier. This should be changed to `com.lamboserver.app` for any proper distribution. |

---

## What NOT to Add

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Notarization services / `xcrun notarytool` | Requires Apple Developer account ($99/yr) — explicitly out of scope | Ad-hoc `codesign -s -` |
| `gon` (Bearer/tap) | Designed for Apple Developer ID signing + notarization. Useless without credentials. | Built-in `codesign` |
| `electron-builder` / `electron-installer-dmg` | Electron-specific tooling; brings npm heavy dependencies for a Go app | `create-dmg` shell script |
| DMG Canvas / DropDMG | GUI tools; not scriptable in a repeatable build process | `create-dmg` |
| Hardened runtime (`-o runtime`) | Required for notarization; incompatible with ad-hoc signing flow | Not applicable without Developer ID |

---

## Version Compatibility

| Component | Version | Notes |
|-----------|---------|-------|
| Wails | v2.12.0 | Already in project. `darwin/universal` platform flag available. |
| `create-dmg` | v1.2.3 | Install via `brew install create-dmg`. Requires macOS 10.12+. |
| `sips` | Built-in | Ships with macOS 10.13+. No install needed. |
| `iconutil` | Built-in | Ships with Xcode CLT. `xcode-select --install` if absent. |
| `codesign` | Built-in | Ships with Xcode CLT. |
| `hdiutil` | Built-in | Used internally by `create-dmg`. Ships with macOS. |

---

## Installation

```bash
# Prerequisites: Xcode Command Line Tools (provides codesign, sips, iconutil, hdiutil)
xcode-select --install   # skip if already installed

# DMG creation tool
brew install create-dmg

# No other installs required. All other tools are macOS built-ins.
```

---

## Sources

- Context7 `/wailsapp/wails` — Wails v2.12.0 build flags, Info.plist template, wails.json `info` block structure, macOS signing workflow (HIGH confidence)
- [Wails Code Signing Guide](https://wails.io/docs/guides/signing/) — Official `gon`-based signing docs (HIGH confidence)
- [Wails Mac App Store Guide](https://wails.io/docs/guides/mac-appstore/) — `codesign` command structure and `wails build -platform darwin/universal` (HIGH confidence)
- [create-dmg/create-dmg](https://github.com/create-dmg/create-dmg) — v1.2.3, `--app-drop-link` option, Homebrew install (HIGH confidence)
- [Ad-Hoc Code Signing a Mac App — Miln](https://stories.miln.eu/graham/2024-06-25-ad-hoc-code-signing-a-mac-app/) — `codesign -s -` command, Gatekeeper limitations (MEDIUM confidence — third-party article, consistent with Apple docs)
- [macOS distribution gist — rsms](https://gist.github.com/rsms/929c9c2fec231f0cf843a1a746a416f5) — Ad-hoc trust model, right-click Open workaround (MEDIUM confidence)
- [Wails issue #3860](https://github.com/wailsapp/wails/issues/3860) — Wails v2 `.icns` auto-generation unreliability (MEDIUM confidence — GitHub issue, not official docs)
- [sips + iconutil workflow — GitHub Gist](https://gist.github.com/jamieweavis/b4c394607641e1280d447deed5fc85fc) — iconset naming convention, `iconutil -c icns` command (HIGH confidence — uses Apple-standard tooling)

---

*Stack research for: LamboServer v1.4 macOS Installer packaging*
*Researched: 2026-04-17*
