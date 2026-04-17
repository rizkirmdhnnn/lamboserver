# Feature Research

**Domain:** macOS installer packaging for a Wails desktop app (no Apple Developer account)
**Researched:** 2026-04-17
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `.app` bundle with correct structure | macOS only runs apps as bundles (MyApp.app/Contents/MacOS/binary). Wails already generates this with `wails build` — it is not a separate task, but it needs correct metadata. | LOW | `build/darwin/Info.plist` template already exists in the repo. Needs `CFBundleIdentifier`, `CFBundleVersion`, `CFBundleIconFile`, and `LSMinimumSystemVersion` populated correctly via `wails.json` `info` section. |
| `Info.plist` with bundle identifier | macOS uses `CFBundleIdentifier` (reverse-domain e.g. `io.lamboserver.app`) to uniquely identify the app in Keychain, permissions, and system registries. Wails generates this from the `info` section in `wails.json`. Missing or using default `com.wails.LamboServer` will cause issues if app ever touches Keychain or is re-installed. | LOW | Set `name` and `info.companyName` in `wails.json` to drive the plist template. The template in `build/darwin/Info.plist` already uses `{{.Name}}` and `{{.Info.ProductVersion}}`. |
| App icon (`.icns`) | macOS shows an icon in Dock, Finder, Launchpad, and Alt-Tab. A missing icon falls back to a default file icon — the app looks unfinished. Wails reads `build/appicon.png` and generates `.icns` automatically on build. | LOW | Source image must be at least 1024x1024 px. Wails handles the `sips`/`iconutil` pipeline. Just replace `build/appicon.png` with a real icon. |
| DMG installer file | macOS convention for distributing an app is a `.dmg` file the user opens and drags the `.app` to `/Applications`. Distributing a raw `.app` folder via zip is acceptable but feels amateur. | MEDIUM | Wails has no built-in DMG creation. Use `create-dmg` (brew) as a post-build step. Script is straightforward but needs to be authored. |
| Ad-hoc code signing | On Apple Silicon, macOS requires all executables to be signed — even ad-hoc. Without any signing the binary will not launch on ARM Macs. Ad-hoc signing uses the `-s -` identity (no certificate needed). | LOW | Single command: `codesign --force --deep -s - LamboServer.app`. No Apple Developer account. No certificate purchase. |
| Production build configuration | `wails build` without flags produces a development binary (larger, not optimized). Production builds use `-trimpath` and `-ldflags "-s -w"` to strip debug info and shrink binary size. | LOW | Add `-ldflags` flags to `wails build` invocation. Can also set `postBuildHooks` in `wails.json` to run the codesign step automatically. |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Universal binary (darwin/universal) | A single `.app` runs natively on both Intel and Apple Silicon. Avoids shipping two separate files or users getting the wrong one. | LOW | `wails build -platform darwin/universal` — Wails supports this natively. Doubles build time. Mildly increases binary size. Strong UX win. |
| DMG background image with arrow | Teaches the user to drag the app to Applications. Reduces support friction. Polished first-impression. | MEDIUM | Requires creating a 600x400 background PNG with an arrow graphic. `create-dmg --background` flag handles placement. Optional but appreciated. |
| Version embedded in binary | `wails.json` `info.productVersion` drives `CFBundleVersion` in `Info.plist` and shows in Finder's Get Info. Provides traceability for support. | LOW | Already supported by Wails plist template. Just requires setting `info.productVersion` in `wails.json`. |
| Custom volume icon for DMG | When the DMG mounts, it shows a custom icon in Finder's sidebar instead of a generic disk image. Polished. | LOW | `create-dmg --volicon` accepts the same `.icns` file. No extra work if the icon is already created. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Apple notarization | Users expect apps not to show Gatekeeper warnings | Requires a paid Apple Developer account ($99/year). Involves uploading the binary to Apple's servers, waiting for approval, and stapling a ticket. Not viable without an account. | Ad-hoc signing + documentation telling users to allow the app via System Settings > Privacy & Security. This is a one-time step per installation. |
| Developer ID signing (full) | Eliminates all Gatekeeper friction | Same as notarization — requires Apple Developer enrollment and certificate management | Ad-hoc signing covers ARM launch requirement. For a personal/internal tool this friction is acceptable. |
| Hardened runtime entitlements | Needed for notarization | Wails apps managing system services (Nginx, PHP, MySQL via Homebrew) need broad entitlements that are difficult to get approved. The `com.apple.security.cs.allow-unsigned-executable-memory` and network server entitlements create notarization review friction. | Skip hardened runtime. Ad-hoc signing with `--deep` is sufficient for local distribution. |
| Sparkle auto-update | Users like automatic updates | Adds significant complexity: requires a separate update server or GitHub Releases feed, certificate signing for update verification, and Sparkle framework integration in a Wails app. Out of scope for this milestone. | Manual update: user downloads new DMG and replaces the app. |
| Mac App Store distribution | Maximum trust, no Gatekeeper | Requires sandboxing. LamboServer manages system services, binds ports, writes to Homebrew paths — none of this is compatible with the App Store sandbox. | Direct DMG distribution only. |

## Feature Dependencies

```
[wails.json info section populated]
    └──drives──> [Info.plist CFBundleIdentifier, CFBundleVersion]
    └──drives──> [Wails production build metadata]

[build/appicon.png (1024x1024)]
    └──drives──> [.icns generation by Wails]
                     └──used by──> [.app bundle icon]
                     └──used by──> [DMG volume icon (optional)]

[wails build -platform darwin/universal]
    └──produces──> [LamboServer.app bundle]
                       └──requires──> [Ad-hoc codesign]
                                          └──enables──> [Launch on Apple Silicon]

[LamboServer.app (signed)]
    └──input to──> [create-dmg script]
                       └──produces──> [LamboServer.dmg]
```

### Dependency Notes

- **`wails.json` info section must be populated before building:** The `Info.plist` template uses `{{.Info.ProductVersion}}`, `{{.Name}}`, and `{{.Info.Copyright}}`. If these are empty, the plist will contain blank fields or defaults.
- **Ad-hoc signing must happen after `wails build` and before DMG creation:** Signing the `.app` first, then packaging into DMG, ensures the DMG contains the correctly signed bundle.
- **`build/appicon.png` is the single source of truth for the icon:** Wails generates `.icns` from this file during build. No manual `iconutil` workflow needed — Wails handles it.
- **Universal build is optional but recommended:** `darwin/universal` produces a fat binary. If skipped, build only for `darwin/amd64` or `darwin/arm64`. For distribution to both Intel and Apple Silicon users, universal is the correct choice.

## MVP Definition

### Launch With (v1.4)

Minimum set to produce a proper installable macOS app:

- [ ] `wails.json` info section populated — `productVersion`, `companyName`, `copyright` set
- [ ] `build/darwin/Info.plist` — verify `CFBundleIdentifier` is set to `io.lamboserver.app` (not the default `com.wails.LamboServer`)
- [ ] App icon — replace `build/appicon.png` with a real 1024x1024 PNG
- [ ] Production build — `wails build -platform darwin/universal -clean -ldflags "-s -w" -trimpath`
- [ ] Ad-hoc code signing — `codesign --force --deep -s - build/bin/LamboServer.app`
- [ ] DMG creation — `create-dmg` script producing `LamboServer-1.0.0.dmg` with drag-to-Applications layout

### Add After Validation (v1.x)

- [ ] DMG background image with arrow graphic — only if users report confusion about installation
- [ ] Custom `.icns` volume icon for DMG — low effort polish if icon asset exists
- [ ] Build script (`Makefile` or `scripts/build-release.sh`) to automate the full pipeline — reduces manual steps

### Future Consideration (v2+)

- [ ] Notarization — only if Apple Developer account is obtained
- [ ] Sparkle auto-update — only after distribution channel is established and user base exists
- [ ] GitHub Actions CI release pipeline — only if releases become frequent

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Info.plist populated correctly | HIGH | LOW | P1 |
| App icon (.icns) | HIGH | LOW | P1 |
| Universal binary build | HIGH | LOW | P1 |
| Ad-hoc code signing | HIGH (required on ARM) | LOW | P1 |
| DMG installer | HIGH | MEDIUM | P1 |
| Production build flags | MEDIUM | LOW | P1 |
| DMG background image | LOW | MEDIUM | P2 |
| Build script automation | MEDIUM | LOW | P2 |
| Notarization | HIGH | HIGH (requires $99/yr account) | P3 |
| Sparkle auto-update | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for this milestone
- P2: Should have, add when possible
- P3: Future milestone

## Gatekeeper Behavior Reference

This is operationally important — users will encounter it:

| Signing Level | Gatekeeper Behavior | User Action Required |
|---------------|--------------------|--------------------|
| Unsigned (no signing) | Blocked outright on Apple Silicon; quarantine warning on Intel | System Settings > Privacy & Security > Open Anyway (macOS Sequoia+) |
| Ad-hoc signed (`-s -`) | Apple Silicon: launches normally (no quarantine issue on local builds). Downloaded via browser: same quarantine warning as unsigned | System Settings > Privacy & Security > Open Anyway (one-time) |
| Developer ID signed | Gatekeeper check passes, no warning | No action |
| Developer ID signed + Notarized | Full trust, no warning, no action | No action |

**Key fact for this milestone:** Ad-hoc signing satisfies the ARM architecture requirement (Apple Silicon Macs require some signature). However, when users download the DMG from the internet, macOS quarantines it. On macOS Sequoia (15+), the old right-click workaround is gone — users must go to System Settings > Privacy & Security > Open Anyway. This is expected behavior for unsigned distribution and should be documented in the release notes.

## Wails Build System Integration

The existing Wails build system handles most packaging automatically:

- `wails build` → creates `build/bin/LamboServer.app` with correct bundle structure
- `build/darwin/Info.plist` → template already exists, needs metadata values populated
- `build/appicon.png` → already present as placeholder; needs replacement with real icon
- `wails.json` → already has `name`, `outputfilename`, `author`; needs `info` section added

The only gap Wails does **not** fill: DMG creation. This requires a separate tool (`create-dmg` via Homebrew) as a post-build step.

## Sources

- Wails v2 macOS signing guide: https://wails.io/docs/guides/signing
- Wails v2 project config reference: https://wails.io/docs/reference/project-config
- Wails v2 CLI build flags: https://wails.io/docs/reference/cli
- Wails Mac App Store guide: https://wails.io/docs/guides/mac-appstore
- create-dmg GitHub: https://github.com/create-dmg/create-dmg
- sindresorhus/create-dmg (simpler wrapper): https://github.com/sindresorhus/create-dmg
- macOS Sequoia Gatekeeper change: https://www.idownloadblog.com/2024/08/07/apple-macos-sequoia-gatekeeper-change-install-unsigned-apps-mac/
- Ad-hoc code signing explanation: https://stories.miln.eu/graham/2024-06-25-ad-hoc-code-signing-a-mac-app/
- Apple Support: Safely open apps on Mac: https://support.apple.com/en-us/102445
- Wails DMG issue (undocumented gap): https://github.com/wailsapp/wails/issues/3926

---
*Feature research for: macOS installer packaging (Wails desktop app, no Apple Developer account)*
*Researched: 2026-04-17*
