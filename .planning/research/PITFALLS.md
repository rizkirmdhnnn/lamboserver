# Pitfalls Research

**Domain:** macOS .app bundling, DMG packaging, and ad-hoc code signing for an existing Wails v2 desktop app that manages system services via exec.Command and osascript
**Researched:** 2026-04-17
**Confidence:** HIGH (multiple authoritative sources: Apple docs, Wails issues, macOS ecosystem post-mortems)

---

## Critical Pitfalls

### Pitfall 1: GUI App PATH Starvation — Homebrew Binaries Not Found at Runtime

**What goes wrong:**
`exec.Command("brew")`, `exec.Command("nginx")`, `exec.Command("php")`, etc. fail with "no such file or directory" or "executable not found" when the bundled `.app` is launched from Finder, Spotlight, or double-click. The same code works in `wails dev` or when running the binary directly from Terminal.

**Why it happens:**
macOS GUI applications launched outside of a terminal session do not inherit the user's shell `PATH`. Homebrew on Apple Silicon installs to `/opt/homebrew/bin`; on Intel to `/usr/local/bin`. Neither is in the default `PATH` that `launchd` gives to GUI processes. The user's `.zshrc` (which adds `/opt/homebrew/bin`) is never sourced. This is a silent failure — no user-facing error unless the app surfaces the underlying `exec` error.

This is a confirmed live issue in the Wails tracker (issue #2507: "node: No such file or directory" in prod mode; issue #3558: bundle works only if you run the inner binary from Terminal).

**How to avoid:**
Never rely on `exec.LookPath` or bare command names (e.g., `exec.Command("brew", ...)`) in production. Use explicit absolute path resolution at startup that checks the known Homebrew prefix locations:
```
/opt/homebrew/bin  (Apple Silicon)
/usr/local/bin     (Intel)
/usr/bin           (system tools)
```
Implement a `resolveBinaryPath(name string) string` function that probes these locations in order before falling back to `LookPath`. The existing `BinaryLocator` in `internal/system/binary.go` already has the right skeleton — extend it to hard-code the Homebrew prefixes. Alternatively, set `PATH` explicitly on each `exec.Cmd` via `cmd.Env = append(os.Environ(), "PATH=/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin")` before calling `cmd.Run()`.

**Warning signs:**
- App works in `wails dev` but fails in `wails build` output
- Log shows "exit status 127" or "no such file or directory" for service binaries
- `exec.LookPath("brew")` returns empty string inside the bundled app

**Phase to address:**
Phase: App Bundle and Production Build — validate all `exec.Command` calls use resolved absolute paths before signing the bundle.

---

### Pitfall 2: Quarantine Attribute Blocks Ad-Hoc Signed Apps on macOS Sequoia 15+

**What goes wrong:**
When users download the DMG from a website, GitHub Releases, or receive it via AirDrop, macOS sets the `com.apple.quarantine` extended attribute on the `.app`. An ad-hoc signed (or unsigned) app with this attribute is blocked by Gatekeeper on macOS Sequoia 15+. The user sees "LamboServer cannot be opened because it is from an unidentified developer." The old Finder workaround (Control-click → Open) was removed in Sequoia 15. Users must now go to System Settings → Privacy & Security → "Open Anyway".

**Why it happens:**
Ad-hoc signing (`codesign -s -`) produces a valid local signature but does not contain a Developer ID certificate. Gatekeeper's quarantine check specifically looks for Developer ID or App Store signatures on downloaded files. macOS Sequoia 15.1 tightened this further, removing the contextual-menu bypass that previously made it easy to override.

**How to avoid:**
Two-pronged approach:
1. Document the manual bypass clearly in README / first-run UI: Settings → Privacy & Security → Open Anyway (single-step workaround for the user).
2. Provide a distribution script that strips quarantine after extracting the DMG, for technically savvy users:
   ```bash
   xattr -r -d com.apple.quarantine /Applications/LamboServer.app
   ```
3. Long-term: obtain a free Apple Developer account ($0 for non-App-Store distribution enrollment). Developer ID signing + notarization removes this problem entirely.

Ad-hoc signing is sufficient for distributing to users who know and trust the source and are willing to bypass Gatekeeper once. For wider distribution, treat notarization as a future milestone requirement.

**Warning signs:**
- Users on Sequoia report "cannot be opened" on first launch
- `spctl --assess --type exec /Applications/LamboServer.app` exits non-zero
- `xattr -l LamboServer.app` shows `com.apple.quarantine` attribute

**Phase to address:**
Phase: DMG Packaging and Distribution — include a `Makefile` or `build.sh` target that codesigns the `.app`, creates the DMG, and documents the Gatekeeper bypass. The release notes must include the Privacy & Security workaround instruction.

---

### Pitfall 3: osascript "do shell script with administrator privileges" Blocked by Hardened Runtime

**What goes wrong:**
The existing `RunWithAdminPrivileges` function in `internal/system/darwin.go` uses `osascript -e 'do shell script "..." with administrator privileges'`. When a Hardened Runtime is enabled (required for notarization, common in production builds), this AppleScript call is blocked unless the app has the `com.apple.security.automation.apple-events` entitlement. Without it, osascript silently fails or produces "Not authorized to send Apple events to System Events."

**Why it happens:**
Hardened Runtime restricts apps from sending Apple Events to other processes by default. `osascript` runs as a separate process; the app is sending it Apple Events. Without the entitlement, this is blocked. Apps built with `wails build` without explicit Hardened Runtime flags may or may not have this issue depending on how the signing is done — ad-hoc signing without `--options runtime` does not enable Hardened Runtime, so this pitfall only activates if you later add `--options runtime` (required for notarization).

**How to avoid:**
For ad-hoc signing (current milestone), this is NOT an issue because ad-hoc signing does not enable Hardened Runtime. Document this as a future constraint: if notarization is ever added, the `entitlements.plist` must include:
```xml
<key>com.apple.security.automation.apple-events</key>
<true/>
<key>NSAppleEventsUsageDescription</key>
<string>LamboServer uses administrator dialogs to install system services.</string>
```
For the current milestone: ensure the codesign command does NOT include `--options runtime` since we are not notarizing.

**Warning signs:**
- Admin password dialogs stop appearing after switching to notarized builds
- `osascript` returns error -1743 "Not authorized to send Apple events"
- Any `RunWithAdminPrivileges` call returns "admin command failed"

**Phase to address:**
Phase: Ad-Hoc Code Signing — explicitly document that `--options runtime` must be omitted. Annotate the entitlements gap for a future notarization phase.

---

### Pitfall 4: sudoers Helper Script Path Breaks After Moving App to /Applications

**What goes wrong:**
The `lambo-helper` script is installed to `~/.lamboserver/bin/lambo-helper`, and the sudoers entry references its full path (e.g., `/Users/alice/.lamboserver/bin/lambo-helper`). If the user reinstalls the app (creates a new version, changes username, or migrates to a new Mac), the sudoers entry points to the old path, breaking all privileged operations silently.

Additionally, the helper script path is hardcoded with `h.paths.Home` at install time. If the user's home directory changes (unlikely but possible) or the app is used by multiple users, the sudoers entry is wrong for the new user.

**Why it happens:**
The sudoers file contains an absolute path to the helper script generated at install time. The `Uninstall` + reinstall flow regenerates the script but the sudoers entry may be stale if the path changes. Worse, the sudoers entry grants `NOPASSWD` root to a script in a user-writable directory — if that script is replaced (even accidentally), the replacement runs as root. This is a known attack vector (similar to the Laravel Valet privilege escalation).

**How to avoid:**
- Keep helper script path stable and under a fixed location (e.g., `/usr/local/bin/lambo-helper` or `/Library/Application Support/LamboServer/bin/lambo-helper`) rather than `~/.lamboserver/bin/`. User-writable paths with `NOPASSWD` sudoers entries are a privilege escalation risk.
- Set script permissions to `0755` owned by `root:wheel` after installing, so the user cannot overwrite it without root. Use `chown root:wheel` and `chmod 755` in the `RunWithAdminPrivileges` install command.
- During app startup, call `Helper.IsInstalled()` and `Helper.UpdateScript()` to detect stale installations and prompt reinstall if the sudoers file points to a path that no longer exists.

**Warning signs:**
- `sudo -n lambo-helper status` exits non-zero after app reinstall
- Privileged operations (daemon install, nginx reload) fail silently
- `/etc/sudoers.d/lamboserver` exists but references a path that `stat` says is missing

**Phase to address:**
Phase: App Bundle and Production Build — validate the helper installation path and permissions before signing. Phase: DMG Packaging — document the stale-sudoers recovery in the uninstall flow.

---

### Pitfall 5: Missing or Wrong entitlements.plist Causes Wails WebView to Crash on Apple Silicon

**What goes wrong:**
The Wails WebView (WKWebView) executes JavaScript. On Apple Silicon, WKWebView requires the `com.apple.security.cs.allow-jit` entitlement to function when Hardened Runtime is enabled. Without it, the WebView silently fails to render content, or crashes on arm64. This affects all Wails apps since the entire UI is WebView-based.

**Why it happens:**
Apple Silicon CPUs enforce JIT restrictions more strictly. WKWebView uses JavaScriptCore, which requires executable+writable memory (JIT). Under Hardened Runtime, the `com.apple.security.cs.allow-jit` entitlement is required to permit this. Ad-hoc signing without `--options runtime` avoids this, but any future attempt to add Hardened Runtime without this entitlement will break the entire UI.

**How to avoid:**
Create `build/darwin/entitlements.plist` now (it does not exist yet in this project) with the minimal required set for a non-sandboxed Wails app that uses osascript and runs shell commands:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <key>com.apple.security.automation.apple-events</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>LamboServer uses administrator dialogs to install system services.</string>
</dict>
</plist>
```
This file is referenced by the `codesign` command during signing. Without it, the app is signed without entitlements, which is fine for ad-hoc now but will silently break WebView if Hardened Runtime is ever added.

**Warning signs:**
- Blank white window on Apple Silicon after production build
- WKWebView loads but shows nothing; no JavaScript execution
- `codesign -dv --entitlements - LamboServer.app` shows empty entitlements

**Phase to address:**
Phase: Ad-Hoc Code Signing — create `entitlements.plist` before running `codesign`.

---

### Pitfall 6: wails.json Missing Required Info Fields — Info.plist Templated with Blanks

**What goes wrong:**
The current `wails.json` has no `info` block (no `productVersion`, `companyName`, `productName`, etc.). Wails uses Go templating to populate `Info.plist` at build time. Missing fields result in a bundle with a blank `CFBundleVersion`, empty `NSHumanReadableCopyright`, and `CFBundleIdentifier` defaulting to `com.wails.lamboserver` (the `com.wails.*` namespace). Gatekeeper and macOS services key on the bundle identifier — a generic identifier may conflict with other Wails apps on the same machine and complicates future notarization.

**Why it happens:**
`wails.json` is a new-project default. Most developers add application-specific info only when they encounter a problem. The `Info.plist` template uses `{{.Info.ProductName}}` etc., which silently renders as empty strings when the fields are absent.

**How to avoid:**
Add an `info` block to `wails.json` before the production build:
```json
"info": {
  "companyName": "Achmad Rizki Ramadhan",
  "productName": "LamboServer",
  "productVersion": "1.4.0",
  "copyright": "Copyright © 2026 Achmad Rizki Ramadhan",
  "comments": "Local web development environment manager"
}
```
Set `CFBundleIdentifier` in `build/darwin/Info.plist` to a reverse-DNS identifier owned by the developer, not the `com.wails.*` default:
```xml
<key>CFBundleIdentifier</key>
<string>dev.rizkirmdhn.lamboserver</string>
```

**Warning signs:**
- `wails build` output shows blank version in About dialog
- `CFBundleIdentifier` is `com.wails.lamboserver` in the built bundle
- macOS Dock or About menu shows empty version string

**Phase to address:**
Phase: App Bundle and Production Build — update `wails.json` and `Info.plist` before any signing or DMG work.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Ad-hoc signing only | No Apple Developer account needed, works today | Gatekeeper blocks downloads on Sequoia+; users must manually override | Acceptable for v1.4 internal/known-user distribution |
| Skip `entitlements.plist` | One less file to create | WebView breaks on arm64 under Hardened Runtime; ad-hoc re-sign is non-trivial | Never — create the file now, even if not used with `--options runtime` yet |
| Bare `exec.Command` names without PATH resolution | Works in dev | Silent failure in bundled app; hard to debug | Never in production code paths |
| Helper script in `~/.lamboserver/bin` | Simple user-space path | Privilege escalation vector; stale sudoers after reinstall | Acceptable for now; fix before wide distribution |
| Single-arch build (arm64 only) | Simpler build step | Intel Mac users cannot run the app | Acceptable if target audience is Apple Silicon only |
| `lipo` universal binary without testing both arches | One DMG for all users | Silent failure on one arch if build flags wrong | Acceptable if CI runs on both arches |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `osascript` admin dialogs | Using Hardened Runtime + `--options runtime` without `com.apple.security.automation.apple-events` entitlement | Add entitlement to `entitlements.plist`; do not use `--options runtime` for ad-hoc-only builds |
| launchd plist install via helper | Helper script path in sudoers becomes stale after app update | Keep helper at a fixed, root-owned path; validate on startup |
| Homebrew service binaries | Calling by name (`brew`, `nginx`, `php`) without resolving PATH | Hard-code Homebrew prefix probing: `/opt/homebrew/bin`, `/usr/local/bin` |
| DMG creation with `create-dmg` | Running `hdiutil create` on a mounted image causes "Resource busy" error | Use `create-dmg --hdiutil-retries 5`; unmount with `hdiutil detach`, not `umount` |
| `codesign` on `.app` with embedded binaries | Signing the `.app` without deep-signing embedded tools first | Use `codesign --deep` or sign each embedded binary individually before signing the bundle |
| pgweb binary inside app | Bundled third-party binary may have its own quarantine attribute | Re-sign all embedded binaries as part of the signing step |

---

## Performance Traps

Not applicable for a desktop app at this scale. The only relevant concern:

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Starting all services on app startup | App takes 10+ seconds to open; user sees blank window | Use lazy service status checks; only detect running state, do not start services at launch | Immediately if auto-start is added |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Helper script in user-writable dir with NOPASSWD sudoers | Any process (or the user) can replace the script and gain root | Install helper to `/usr/local/bin/lambo-helper`, chown `root:wheel`, chmod `755` after install |
| Shipping `entitlements.plist` with `com.apple.security.cs.allow-unsigned-executable-memory` when not needed | Opens the app to code injection attacks | Only include entitlements that are actually required; remove if Hardened Runtime is not used |
| Storing Homebrew paths as hardcoded `/usr/local` only | App silently breaks on Apple Silicon Macs | Always check both `/opt/homebrew/bin` (arm64) and `/usr/local/bin` (x86_64) |
| Distributing a DMG without stripping quarantine in the build script | Users get blocked by Gatekeeper with no clear recovery path | Document xattr removal command in release notes and README |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No explanation when app is quarantine-blocked | User sees cryptic Apple dialog; believes app is broken | Include a first-run helper page or README section: "First Launch on macOS: go to System Settings → Privacy & Security → Open Anyway" |
| Blank app icon (missing or wrong `.icns`) | App shows generic document icon in Dock and Finder; looks unfinished | Generate all 10 required icon sizes from a 1024x1024 source using `sips` + `iconutil` before the build |
| Missing `NSHumanReadableCopyright` | About dialog shows blank copyright | Always populate the `info` block in `wails.json` |
| Version string "0.0.0" or empty in About dialog | App looks like a prototype | Set `productVersion` in `wails.json` to the milestone version before building |

---

## "Looks Done But Isn't" Checklist

- [ ] **Code signing:** `codesign -dv --verbose=4 LamboServer.app` shows the expected signing identity and entitlements, not "ad-hoc" with empty entitlements dict
- [ ] **Bundle identifier:** `defaults read LamboServer.app/Contents/Info.plist CFBundleIdentifier` shows a non-`com.wails.*` value
- [ ] **PATH resolution:** Run the production `.app` by double-clicking (not from Terminal) and verify all service status checks work — Homebrew binaries are found
- [ ] **Icon:** `LamboServer.app` shows the correct icon in Finder and Dock (not the Wails default rocket or generic document icon)
- [ ] **Admin dialog:** Launching the app and triggering a first-time setup shows the macOS password prompt (osascript admin dialog works in bundled mode)
- [ ] **DMG layout:** Opening the DMG shows the `.app` and an Applications folder alias — the drag-to-install UX works
- [ ] **Quarantine test:** Download the DMG from a web URL (not `cp`), open it, and verify what the user actually sees — does Gatekeeper block it?
- [ ] **Entitlements file:** `build/darwin/entitlements.plist` exists and is referenced in the `codesign` command
- [ ] **Universal binary (if targeted):** `lipo -info LamboServer.app/Contents/MacOS/LamboServer` shows both `arm64` and `x86_64`
- [ ] **Version string:** About dialog and `Info.plist` show the correct milestone version, not empty or "0.0.0"

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| PATH starvation discovered post-release | MEDIUM | Add explicit PATH resolution in `internal/system/binary.go`; rebuild and re-release DMG |
| Gatekeeper blocks quarantined DMG | LOW | Document `xattr -r -d com.apple.quarantine` workaround; no code change needed |
| osascript blocked by Hardened Runtime (future) | MEDIUM | Add `com.apple.security.automation.apple-events` to `entitlements.plist`; re-sign and rebuild |
| Stale sudoers path after reinstall | LOW | User runs `sudo rm /etc/sudoers.d/lamboserver`; app re-installs on next launch |
| Wrong bundle identifier shipped | HIGH | Changing bundle identifier invalidates stored preferences and any future notarization history; fix before first public release |
| Missing `entitlements.plist` under Hardened Runtime | HIGH | WebView is non-functional; requires complete rebuild; catch in local testing before distribution |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| GUI app PATH starvation | App Bundle & Production Build | Launch `.app` by double-click; confirm service status reads succeed |
| Quarantine blocks ad-hoc DMG on Sequoia | DMG Packaging & Distribution | Download DMG from URL, open on Sequoia 15+ — verify exact UX and document workaround |
| osascript blocked by Hardened Runtime | Ad-Hoc Code Signing | Ensure `--options runtime` is absent from `codesign` command; annotate for future notarization |
| Stale sudoers helper path | App Bundle & Production Build | Run `Helper.IsInstalled()` in startup; test reinstall flow |
| Missing entitlements crashes WebView (arm64) | Ad-Hoc Code Signing | Create `build/darwin/entitlements.plist` and reference it; test on Apple Silicon |
| wails.json missing Info fields | App Bundle & Production Build | Verify `Info.plist` in built bundle has non-empty version, identifier, and copyright |

---

## Sources

- Wails issue #2507: `exec.Command` production PATH failure — https://github.com/wailsapp/wails/issues/2507
- Wails issue #3558: Bundle file does not work, only inner binary does — https://github.com/wailsapp/wails/issues/3558
- macOS Sequoia Gatekeeper changes (August 2024) — https://eclecticlight.co/2024/08/10/gatekeeper-and-notarization-in-sequoia/
- Apple Sequoia 15.1 removes Gatekeeper bypass — https://hackaday.com/2024/11/01/apple-forces-the-signing-of-applications-in-macos-sequoia-15-1/
- Sequoia removes Control-click Gatekeeper override — https://www.idownloadblog.com/2024/08/07/apple-macos-sequoia-gatekeeper-change-install-unsigned-apps-mac/
- Apple Developer: Hardened Runtime entitlements — https://developer.apple.com/documentation/security/hardened-runtime
- Apple Developer: com.apple.security.cs.allow-jit — https://developer.apple.com/documentation/BundleResources/Entitlements/com.apple.security.cs.allow-jit
- Sending AppleScript events from hardened app — https://ishaangandhi.medium.com/sending-applescript-events-from-electron-app-18dc1b7d7a51
- macOS distribution gist (quarantine, signing, notarization) — https://gist.github.com/rsms/929c9c2fec231f0cf843a1a446a416f5
- Wails Code Signing guide — https://wails.io/docs/guides/signing/
- Wails Mac App Store guide (entitlements reference) — https://wails.io/docs/v2.9.0/guides/mac-appstore/
- create-dmg tool (hdiutil pitfalls) — https://github.com/create-dmg/create-dmg
- Laravel Valet privilege escalation / helper script security — https://jozefcipa.com/blog/how-to-use-sudo-without-a-password-in-your-programs/
- macOS GUI app PATH problem — https://www.bounga.org/tips/2020/04/07/instructs-mac-os-gui-apps-about-path-environment-variable/

---
*Pitfalls research for: macOS .app bundling, DMG packaging, ad-hoc signing — Wails v2 desktop app*
*Researched: 2026-04-17*
