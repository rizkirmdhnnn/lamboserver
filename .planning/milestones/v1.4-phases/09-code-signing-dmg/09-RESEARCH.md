# Phase 9: Code Signing & DMG - Research

**Researched:** 2026-04-17
**Domain:** macOS ad-hoc code signing, hdiutil DMG creation, Gatekeeper documentation
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Build pipeline is a shell script at `scripts/build-dmg.sh` (matches existing `scripts/generate-icns.sh` pattern)
- **D-02:** Pipeline steps: `wails build -platform darwin/universal -clean` -> `codesign -s -` with entitlements -> `hdiutil` to create DMG
- **D-03:** Script does NOT include icon generation step — `iconfile.icns` is already committed and built. Script starts at wails build.
- **D-04:** Ad-hoc signing with `codesign -s -` (no Apple Developer account — project constraint)
- **D-05:** Entitlements applied from `build/darwin/entitlements.plist` (already exists with JIT, unsigned memory, library validation, Apple Events, network client/server)
- **D-06:** No Hardened Runtime, no App Sandbox (project constraints — blocks exec.Command calls)
- **D-07:** DMG created with plain `hdiutil` (macOS built-in, no external dependencies)
- **D-08:** DMG filename is `LamboServer-1.0.0.dmg` (version from wails.json)
- **D-09:** DMG contents: LamboServer.app + Applications folder alias only. No README, no LICENSE inside the DMG.
- **D-10:** Gatekeeper workaround instructions go in GitHub release notes only (not in README.md or inside DMG)
- **D-11:** Build script only outputs the DMG — no release notes template generation

### Claude's Discretion

- Exact hdiutil flags and volume name for DMG creation
- codesign flags beyond `-s -` and `--entitlements` (e.g., `--deep`, `--force`)
- Whether to verify the signature as part of the build script (codesign --verify)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SIGN-01 | `.app` bundle is ad-hoc signed with `codesign -s -` | codesign command verified available on build machine; exact flags documented in Architecture Patterns |
| SIGN-02 | `entitlements.plist` exists with network entitlements (no sandbox, no Hardened Runtime) | File verified present at `build/darwin/entitlements.plist` with correct keys |
| SIGN-03 | Signed app passes `codesign --verify --deep` | Verified that current build ALREADY passes (exit 0) — signing step must preserve this |
| DMG-01 | DMG installer is created with drag-to-Applications layout | hdiutil two-pass pattern documented; Applications symlink approach confirmed |
| DMG-02 | Build script automates the full pipeline (build, sign, package DMG) | `scripts/build-dmg.sh` pattern documented end-to-end |
| DOC-01 | Release notes include Gatekeeper workaround instructions (Sequoia: System Settings > Privacy & Security > Open Anyway) | Workaround text and xattr command documented; Sequoia behavior verified from prior research |
</phase_requirements>

---

## Summary

Phase 9 is the final phase of the v1.4 milestone. The work is tightly scoped: one shell script (`scripts/build-dmg.sh`) that runs `wails build`, `codesign`, and `hdiutil` in sequence, plus a release notes template. All prerequisite assets already exist from Phases 7 and 8.

**Current state (verified 2026-04-17):** `build/bin/LamboServer.app` is already present as a universal binary (x86_64 + arm64) with an ad-hoc signature (`flags=0x2(adhoc)`) and passes `codesign --verify --deep` with exit 0. The bundle identifier is `dev.lamboserver.app`, version 1.0.0, copyright populated. `build/darwin/entitlements.plist` exists with all required keys. The build-dmg.sh script does NOT yet exist — that is the primary deliverable.

**Key constraint:** The user chose plain `hdiutil` over `create-dmg` (no external dependencies). This requires a two-pass approach: create a writable staging folder, add an `/Applications` symlink, run `hdiutil create -srcfolder` to produce a compressed DMG. No mounting/unmounting loop needed when using `-srcfolder`.

**Primary recommendation:** The `scripts/build-dmg.sh` script should follow the existing `generate-icns.sh` pattern (`set -e`, sequential steps, echo progress). Use `codesign --force --deep -s - --entitlements` for the signing step (force re-signs even if already signed, deep covers the bundle structure). Use `hdiutil create -srcfolder` with a temporary staging directory to produce the final DMG with Applications symlink.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Universal binary build | Build system (wails) | — | `wails build -platform darwin/universal` produces the .app |
| Code signing | Build script (codesign) | — | External to the app; applied post-build |
| DMG packaging | Build script (hdiutil) | — | External to the app; applied post-sign |
| Entitlements | Build config (plist) | Build script (codesign) | Plist exists; script references it |
| Gatekeeper documentation | Release notes (GitHub) | — | Decided D-10: GitHub release notes only |

---

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `codesign` | Built-in (Xcode CLT) | Ad-hoc sign the .app bundle | Ships with every Mac that has Xcode CLT; no account required for `-s -` |
| `hdiutil` | Built-in (macOS) | Create compressed read-only DMG | macOS built-in; zero dependencies; handles all DMG format variants |
| `wails build` | v2.12.0 (existing) | Produce universal .app from Go + React | Already in project; `-platform darwin/universal` confirmed working from Phase 8 |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `ln -s /Applications` | Built-in (shell) | Create Applications folder alias inside DMG staging | Must be created in staging dir before `hdiutil create -srcfolder` |
| `mkdir -p` | Built-in (shell) | Create staging directory for DMG contents | One-time per build run |

### Alternatives Considered (and Rejected per D-07)

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `hdiutil` (built-in) | `create-dmg` (Homebrew) | create-dmg is more ergonomic (single command, layout options) but requires `brew install create-dmg`; user decided no external dependencies |

**Installation:**

No installation required. All tools are macOS built-ins available on any machine with Xcode Command Line Tools.

**Version verification (performed 2026-04-17):**
- `codesign`: `/usr/bin/codesign` — present [VERIFIED: Bash]
- `hdiutil`: `/usr/bin/hdiutil` — present [VERIFIED: Bash]
- `wails`: v2.12.0 — present [VERIFIED: Bash]
- `create-dmg`: NOT FOUND — confirms hdiutil-only approach is required [VERIFIED: Bash]

---

## Architecture Patterns

### System Architecture Diagram

```
[wails build -platform darwin/universal -clean]
        |
        v
[build/bin/LamboServer.app] (universal binary, unentitled ad-hoc or unsigned)
        |
        v
[codesign --force --deep -s - --entitlements build/darwin/entitlements.plist]
        |
        v
[build/bin/LamboServer.app] (ad-hoc signed, entitlements embedded, _CodeSignature/ populated)
        |
        v
[mkdir -p /tmp/LamboServer-dmg-staging/]
[cp -r build/bin/LamboServer.app -> staging/]
[ln -s /Applications -> staging/Applications]
        |
        v
[hdiutil create -volname "LamboServer" -srcfolder staging/ -ov -format UDZO -o LamboServer-1.0.0.dmg]
        |
        v
[LamboServer-1.0.0.dmg] (drag-to-Applications installer, distributable)
        |
        v
[rm -rf staging/] (cleanup)
```

### Recommended Project Structure

```
scripts/
├── generate-icns.sh    # existing — icon generation (not called by build-dmg.sh)
└── build-dmg.sh        # NEW — full build pipeline: wails build -> sign -> DMG

build/
├── darwin/
│   ├── entitlements.plist   # EXISTS — referenced by codesign step
│   ├── Info.plist           # EXISTS — Wails template, rendered at build time
│   └── iconfile.icns        # EXISTS — used by Wails for .app icon
└── bin/
    ├── LamboServer.app      # EXISTS — output of wails build; overwritten each run
    └── (LamboServer-1.0.0.dmg is written to project root, not build/bin)
```

**DMG output location:** Write DMG to project root (or a `dist/` directory) so it is obvious to the user after the script completes. Do NOT put it inside `build/bin/` where it could be cleaned by `wails build -clean`.

### Pattern 1: hdiutil Two-Pass DMG Creation with Staging Directory

**What:** Create a temporary staging folder, populate it with .app and Applications symlink, then use `hdiutil create -srcfolder` to produce a compressed read-only DMG in a single pass. No mount/unmount loop required.

**When to use:** Any time you need a drag-to-Applications DMG without external tools.

**Example:**
```bash
# Source: hdiutil create -help (macOS built-in), cross-referenced with
# https://gist.github.com/jadeatucker/5382343 and
# https://asmaloney.com/2013/07/howto/packaging-a-mac-os-x-application-using-a-dmg/

APP_NAME="LamboServer"
VERSION="1.0.0"
STAGING_DIR="$(mktemp -d)"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"

# Populate staging
cp -r "build/bin/${APP_NAME}.app" "${STAGING_DIR}/"
ln -s /Applications "${STAGING_DIR}/Applications"

# Create compressed DMG from staging folder
hdiutil create \
  -volname "${APP_NAME}" \
  -srcfolder "${STAGING_DIR}" \
  -ov \
  -format UDZO \
  "${DMG_OUT}"

# Cleanup
rm -rf "${STAGING_DIR}"
```

**Flag rationale:**
- `-volname "LamboServer"`: The volume name shown in Finder when DMG is opened (Claude's Discretion area)
- `-srcfolder`: Takes the staging directory as the DMG's contents — simpler than mount/unmount loop
- `-ov`: Overwrite if DMG already exists (idempotent builds)
- `-format UDZO`: Zlib-compressed read-only DMG — the standard distribution format [VERIFIED: hdiutil help output]

### Pattern 2: codesign Ad-Hoc Signing with Entitlements

**What:** Sign the .app bundle in-place with an ad-hoc identity (`-s -`), embedding the entitlements plist.

**When to use:** Always, after `wails build`, before DMG creation.

**Example:**
```bash
# Source: Apple codesign man page; confirmed against current signing state
# of build/bin/LamboServer.app (already ad-hoc signed, exit 0 verify)

codesign \
  --force \
  --deep \
  -s - \
  --entitlements "build/darwin/entitlements.plist" \
  "build/bin/${APP_NAME}.app"
```

**Flag rationale:**
- `--force`: Re-signs even if already signed (required after wails build which may self-sign)
- `--deep`: Signs nested binaries inside the .app (Contents/MacOS/, frameworks) [ASSUMED — acceptable for ad-hoc; Apple's guidance against --deep applies to Developer ID notarization only]
- `-s -`: Ad-hoc identity (dash = no certificate)
- `--entitlements`: Embeds the plist so entitlements are part of the code signature

**Verification step (include in script):**
```bash
codesign --verify --deep --strict "build/bin/${APP_NAME}.app" && echo "Signature valid"
```

### Pattern 3: Full build-dmg.sh Script Structure

```bash
#!/bin/bash
set -e

APP_NAME="LamboServer"
VERSION="1.0.0"
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "[1/4] Building universal .app..."
wails build -platform darwin/universal -clean

echo "[2/4] Ad-hoc signing with entitlements..."
codesign --force --deep -s - --entitlements "${ENTITLEMENTS}" "${APP_PATH}"
codesign --verify --deep "${APP_PATH}" && echo "  Signature OK"

echo "[3/4] Creating DMG..."
STAGING=$(mktemp -d)
cp -r "${APP_PATH}" "${STAGING}/"
ln -s /Applications "${STAGING}/Applications"
hdiutil create -volname "${APP_NAME}" -srcfolder "${STAGING}" -ov -format UDZO "${DMG_OUT}"
rm -rf "${STAGING}"

echo "[4/4] Done: ${DMG_OUT}"
```

### Anti-Patterns to Avoid

- **Signing AFTER DMG creation:** Always sign the .app before packaging into the DMG. The DMG itself is not signed — only the .app inside needs to be.
- **Using `--options runtime` with ad-hoc signing:** This flag enables Hardened Runtime, which is only valid with Developer ID certificates. Including it with `-s -` produces a meaningless signature that may break osascript. [VERIFIED: PITFALLS.md Pitfall 3]
- **Not using `--force`:** Without `--force`, codesign fails if the bundle already has a signature (wails build may produce an ad-hoc pre-signed bundle).
- **Writing DMG inside `build/bin/`:** The `wails build -clean` flag removes the entire `build/bin/` directory before building. A DMG there would be destroyed on the next build run.
- **Using `hdiutil attach` + manual copy + `hdiutil detach` loop:** The `-srcfolder` flag in `hdiutil create` eliminates the need for a mount/unmount loop. The loop approach is error-prone (resource busy errors if Finder opens the DMG).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| .app bundle structure | Custom folder assembly | `wails build` | Wails handles Go binary compilation, frontend embedding, Info.plist template rendering, and .app structure |
| Code signature | Custom checksum scheme | `codesign -s -` | Apple's code signing is cryptographic and format-specific; macOS API calls check for valid CodeDirectory structure |
| DMG creation | Custom archive format | `hdiutil create` | DMG format has specific HFS+ structures; hdiutil is the authoritative tool |
| Entitlements plist | Inline in script | `build/darwin/entitlements.plist` | File already exists and is correct; referencing it is one flag |

**Key insight:** All the heavy lifting (binary compilation, app bundling, DMG format) is handled by existing tools. The `build-dmg.sh` script is purely orchestration — it calls the right tools in the right order with the right flags.

---

## Current State Assessment (Critical for Planning)

**What Phase 8 completed (verified 2026-04-17):**

| Asset | State | Verified By |
|-------|-------|-------------|
| `build/bin/LamboServer.app` | EXISTS — universal binary (x86_64 + arm64) | `lipo -info` |
| `build/bin/LamboServer.app` signing | Ad-hoc signed (`flags=0x2(adhoc)`) | `codesign -dv` |
| `codesign --verify --deep` | PASSES (exit 0) | `codesign --verify --deep` |
| `build/darwin/entitlements.plist` | EXISTS — 6 entitlement keys, no sandbox | `Read` tool |
| `build/darwin/Info.plist` | Template with `CFBundleIdentifier=dev.lamboserver.app` | `Read` tool |
| Built `Info.plist` in .app | Rendered with correct identifier, version 1.0.0, copyright | `cat .app/Contents/Info.plist` |
| `wails.json` info block | EXISTS — productVersion 1.0.0, copyright © 2026 LamboServer | `Read` tool |
| `scripts/generate-icns.sh` | EXISTS — pattern for build-dmg.sh to follow | `Read` tool |

**What does NOT exist yet (Phase 9 creates):**

- `scripts/build-dmg.sh` — the primary deliverable
- `LamboServer-1.0.0.dmg` — produced by the script
- Release notes content for GitHub (DOC-01)

**Key discovery:** The current .app is already ad-hoc signed but WITHOUT entitlements (the `codesign -d --entitlements` output showed no entitlements dict in the current build). The signing step in build-dmg.sh must use `--force` and `--entitlements` to re-sign with the plist. [VERIFIED: Bash — codesign output showed empty entitlements on current .app]

---

## Common Pitfalls

### Pitfall 1: codesign without --entitlements produces empty entitlement dict

**What goes wrong:** Running `codesign -s - app.app` (without `--entitlements`) signs the bundle with zero entitlements. The WKWebView may still work on Intel but will fail on Apple Silicon if Hardened Runtime is ever applied. More critically, `codesign --verify --deep` may pass but the entitlements that osascript and network operations need are not embedded.

**Why it happens:** The entitlements flag is optional; codesign does not look for the plist automatically.

**How to avoid:** Always include `--entitlements build/darwin/entitlements.plist` in the codesign command. Verify with `codesign -d --entitlements - LamboServer.app` to confirm keys are present.

**Warning signs:** `codesign -d --entitlements - app.app` shows empty output or no dict.

### Pitfall 2: DMG written to build/bin/ gets deleted by -clean flag

**What goes wrong:** The script writes `LamboServer-1.0.0.dmg` into `build/bin/`. On the next run, `wails build -platform darwin/universal -clean` deletes `build/bin/` entirely before building.

**Why it happens:** `-clean` removes the entire output directory.

**How to avoid:** Write the DMG to the project root or a `dist/` directory that is outside `build/`. The script should use a path like `"${APP_NAME}-${VERSION}.dmg"` (project root) not `"build/bin/${APP_NAME}-${VERSION}.dmg"`.

### Pitfall 3: Gatekeeper blocks downloaded DMG on macOS Sequoia 15+

**What goes wrong:** Users download the DMG from GitHub Releases. macOS quarantines the file. On macOS Sequoia 15.1+, the old Control-click → Open bypass was removed. Users see "cannot be opened because it is from an unidentified developer" with no obvious recovery.

**Why it happens:** Ad-hoc signing is local-identity only. Gatekeeper rejects it for downloaded files. macOS Sequoia tightened this. [VERIFIED: PITFALLS.md Pitfall 2, multiple sources]

**How to avoid:**
1. Release notes (DOC-01) must include the exact path: System Settings → Privacy & Security → scroll to "LamboServer was blocked" → click "Open Anyway"
2. Include the technical workaround for developers: `xattr -r -d com.apple.quarantine /Applications/LamboServer.app`

**Warning signs:** `spctl --assess --type exec LamboServer.app` exits non-zero after downloading from a URL.

### Pitfall 4: hdiutil "resource busy" when staging dir is on a mounted volume

**What goes wrong:** If the staging directory happens to be inside a mounted DMG or network volume, `hdiutil create -srcfolder` may fail with "resource busy."

**Why it happens:** hdiutil cannot read from an already-attached hdiutil volume.

**How to avoid:** Use `mktemp -d` to create the staging directory in `/tmp/` (always a local filesystem). Do not reuse any existing mounted DMG path.

### Pitfall 5: cp -r copies symlink as symlink, not as resolved path

**What goes wrong:** The Applications symlink (`ln -s /Applications`) created in the staging dir should remain a symlink so Finder shows the Applications alias in the DMG. If `cp -r` is used to copy the staging dir to another location, it may dereference the symlink and fail.

**Why it happens:** `hdiutil create -srcfolder` reads the staging dir directly — no copy of the staging dir is needed. This pitfall applies only if someone adds an extra copy step.

**How to avoid:** Pass the staging dir directly to `-srcfolder`. Do not `cp` the staging dir.

---

## Code Examples

Verified patterns from macOS built-in tools and project conventions:

### Complete build-dmg.sh

```bash
#!/bin/bash
# Source: project convention (scripts/generate-icns.sh pattern)
set -e

APP_NAME="LamboServer"
VERSION="1.0.0"
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "[1/4] Building universal .app..."
wails build -platform darwin/universal -clean

echo "[2/4] Ad-hoc signing with entitlements..."
codesign --force --deep -s - --entitlements "${ENTITLEMENTS}" "${APP_PATH}"
codesign --verify --deep "${APP_PATH}"
echo "  Signature verified"

echo "[3/4] Packaging DMG..."
STAGING=$(mktemp -d)
trap "rm -rf ${STAGING}" EXIT
cp -r "${APP_PATH}" "${STAGING}/"
ln -s /Applications "${STAGING}/Applications"
hdiutil create \
  -volname "${APP_NAME}" \
  -srcfolder "${STAGING}" \
  -ov \
  -format UDZO \
  "${DMG_OUT}"

echo "[4/4] Done. Output: ${DMG_OUT}"
```

**Notes on this script:**
- `trap "rm -rf ${STAGING}" EXIT`: Ensures cleanup even if the script exits early on any error (works with `set -e`)
- `-ov`: Overwrites existing DMG so repeated builds are idempotent
- DMG written to project root (not `build/bin/`) to survive `-clean` on next run

### Verify Signing After Script

```bash
# Confirm entitlements are embedded
codesign -d --entitlements - build/bin/LamboServer.app

# Confirm verify passes
codesign --verify --deep --strict build/bin/LamboServer.app && echo "OK"

# Inspect DMG contents (mount and list)
hdiutil attach LamboServer-1.0.0.dmg
ls /Volumes/LamboServer/
hdiutil detach /Volumes/LamboServer
```

### Gatekeeper Workaround (Release Notes DOC-01)

```markdown
## First Launch on macOS

LamboServer is not notarized by Apple. On first launch, macOS may block the app.

**To open LamboServer:**
1. Open **System Settings** → **Privacy & Security**
2. Scroll down to the Security section
3. You should see "LamboServer was blocked from use because it is not from an identified developer"
4. Click **Open Anyway**
5. Enter your password when prompted

**Alternative (Terminal):**
```bash
xattr -r -d com.apple.quarantine /Applications/LamboServer.app
```

This warning appears only once per installation.
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Control-click → Open to bypass Gatekeeper | System Settings → Privacy & Security → Open Anyway | macOS Sequoia 15.1 (Nov 2024) | Release notes must use the new path |
| `--deep` flag for all codesign operations | Use `--deep` only for ad-hoc; individual signing for Developer ID | Apple guidance (still valid) | For ad-hoc, `--deep` is acceptable and simpler |
| `hdiutil` mount/unmount loop | `hdiutil create -srcfolder` (single pass) | macOS 10.x | Simpler and avoids "resource busy" errors |

**Deprecated/outdated:**
- `create-dmg` npm package (sindresorhus): Superseded by the shell script version; user decided against external tools anyway
- `gon` tool: Requires Apple Developer credentials; not applicable

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `--deep` is acceptable for ad-hoc signing of a Wails app with no embedded frameworks | Code Examples | Low — if wrong, signing still succeeds; `--deep` is conservative/safe for ad-hoc |
| A2 | `-format UDZO` produces the correct standard DMG format for distribution | Code Examples | Low — UDZO is confirmed as zlib-compressed from hdiutil help; widely used |
| A3 | The staging dir approach with `ln -s /Applications` produces the correct Finder "drag here" alias inside the DMG | Architecture Patterns | Medium — the Applications symlink resolves correctly when the DMG is mounted on another Mac; tested pattern on macOS but not verified on macOS 26.4.1 in this session |
| A4 | `codesign --verify --deep` succeeds after re-signing with `--force --entitlements` | Code Examples | Low — current build already passes verify; adding entitlements should not break it |

---

## Open Questions

1. **DMG output location: project root vs. `dist/` directory**
   - What we know: DMG must not be in `build/bin/` (cleared by `-clean`)
   - What's unclear: Whether a `dist/` directory should be created or the project root is acceptable
   - Recommendation: Project root is simplest and matches what users expect to find after running a build script. If the project later adds CI, `dist/` is easy to add. Default to project root for now.

2. **Volume name for DMG**
   - What we know: This is Claude's Discretion per CONTEXT.md
   - Recommendation: Use `"LamboServer"` as the volume name — it's the product name and matches what Finder will display as the mounted volume name. Avoids confusion from version-suffixed volume names.

3. **Whether to add `.gitignore` entry for the DMG**
   - What we know: DMG will be written to project root
   - What's unclear: Current `.gitignore` content
   - Recommendation: Check `.gitignore` and add `*.dmg` if not present — DMG files are binary build artifacts and should not be committed.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `codesign` | SIGN-01, SIGN-03 | ✓ | `/usr/bin/codesign` (Xcode CLT) | None — required for signing |
| `hdiutil` | DMG-01, DMG-02 | ✓ | `/usr/bin/hdiutil` (macOS built-in) | None — required for DMG |
| `wails` | DMG-02 | ✓ | v2.12.0 | None — required for build |
| `build/bin/LamboServer.app` | SIGN-01 | ✓ | Universal binary (x86_64 + arm64) | Run `wails build -platform darwin/universal -clean` |
| `build/darwin/entitlements.plist` | SIGN-02 | ✓ | Committed, 6 keys | None — already exists |
| `create-dmg` | — | ✗ | Not installed | Uses `hdiutil` per D-07 |

**Missing dependencies with no fallback:** None — all required tools present.

**Missing dependencies with fallback:** `create-dmg` not needed (hdiutil used per decision D-07).

---

## Validation Architecture

> `workflow.nyquist_validation` key is absent from `.planning/config.json` — treated as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Shell assertions + `codesign` / `hdiutil` verification commands |
| Config file | None — shell-based verification in the script itself |
| Quick run command | `codesign --verify --deep build/bin/LamboServer.app` |
| Full suite command | `bash scripts/build-dmg.sh && codesign --verify --deep build/bin/LamboServer.app && hdiutil attach LamboServer-1.0.0.dmg && ls /Volumes/LamboServer/ && hdiutil detach /Volumes/LamboServer` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | Notes |
|--------|----------|-----------|-------------------|-------|
| SIGN-01 | .app is ad-hoc signed with `codesign -s -` | smoke | `codesign -dv build/bin/LamboServer.app 2>&1 \| grep 'Signature=adhoc'` | ❌ Wave 0 — add after script created |
| SIGN-02 | entitlements.plist exists with required keys | unit | `plutil -lint build/darwin/entitlements.plist && codesign -d --entitlements - build/bin/LamboServer.app 2>&1 \| grep 'allow-jit'` | ✅ File exists |
| SIGN-03 | Signed app passes codesign --verify --deep | smoke | `codesign --verify --deep build/bin/LamboServer.app` | ✅ Already passes |
| DMG-01 | DMG has drag-to-Applications layout | manual | Mount DMG, verify Applications alias visible | Manual verification required |
| DMG-02 | build-dmg.sh runs full pipeline | integration | `bash scripts/build-dmg.sh` (exit 0 = pass) | ❌ Wave 0 — script doesn't exist yet |
| DOC-01 | Release notes include Gatekeeper workaround | manual | Review GitHub release notes content | Manual review |

### Sampling Rate

- **Per task commit:** `codesign --verify --deep build/bin/LamboServer.app`
- **Per wave merge:** `bash scripts/build-dmg.sh && codesign --verify --deep build/bin/LamboServer.app`
- **Phase gate:** Full pipeline (build, sign, DMG) green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] No automated test for SIGN-01 (ad-hoc flag check) — add inline verification to build-dmg.sh
- [ ] `scripts/build-dmg.sh` does not exist — Wave 0 creates it
- [ ] DMG-01 (layout) and DOC-01 (release notes) require manual verification — note in plan

---

## Security Domain

> `security_enforcement` not explicitly set to false — included.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Not applicable — no auth in build/sign pipeline |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No | Not applicable |
| V5 Input Validation | No | Build script has no user input |
| V6 Cryptography | Yes (partial) | Ad-hoc signing uses SHA-256 CodeDirectory hash — no hand-rolling |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Tampered .app after signing | Tampering | `codesign --verify --deep` catches modifications to the bundle after signing |
| Helper script privilege escalation | Elevation of Privilege | Not in scope for Phase 9; documented in PITFALLS.md Pitfall 4 for future |
| Quarantine bypass misleads users | Spoofing | Document exact System Settings path so users know the legitimate workaround |
| Distributing unsigned entitlements-free app | Tampering | `--entitlements` flag must be included; verify with `codesign -d --entitlements -` |

---

## Sources

### Primary (HIGH confidence)

- Bash verification of `/usr/bin/codesign`, `/usr/bin/hdiutil`, `wails version` — tool availability [VERIFIED: Bash]
- `codesign -dv --verbose=4 build/bin/LamboServer.app` — current signing state [VERIFIED: Bash]
- `lipo -info ...LamboServer` — universal binary confirmed [VERIFIED: Bash]
- `cat build/darwin/entitlements.plist` — entitlements content [VERIFIED: Read tool]
- `hdiutil create -help` — all hdiutil flags documented [VERIFIED: Bash]
- `build/bin/LamboServer.app/Contents/Info.plist` — rendered bundle metadata [VERIFIED: Bash]
- `.planning/research/PITFALLS.md` — signing and DMG pitfalls from prior research [VERIFIED: Read tool]
- `.planning/research/STACK.md` — stack decisions from prior research [VERIFIED: Read tool]
- `09-CONTEXT.md` — locked decisions and discretion areas [VERIFIED: Read tool]

### Secondary (MEDIUM confidence)

- [Packaging a Mac OS X Application Using a DMG](https://asmaloney.com/2013/07/howto/packaging-a-mac-os-x-application-using-a-dmg/) — hdiutil two-pass pattern [CITED: asmaloney.com]
- [GitHub Gist: DMG creation with hdiutil](https://gist.github.com/jadeatucker/5382343) — Applications symlink pattern [CITED: gist.github.com]
- macOS Sequoia Gatekeeper changes (from PITFALLS.md) — Privacy & Security workaround path [CITED: eclecticlight.co via PITFALLS.md]

### Tertiary (LOW confidence)

None — all findings verified via tools or cited sources.

---

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — all tools verified present on build machine; versions confirmed
- Architecture: HIGH — hdiutil flags verified against help output; signing verified against current .app
- Pitfalls: HIGH — from prior project research, cross-referenced with live tool verification
- DMG layout pattern: MEDIUM — Applications symlink approach is well-documented but not tested on macOS 26.4.1 in this session

**Research date:** 2026-04-17
**Valid until:** 2026-07-17 (macOS built-in tools are stable; Wails 2.x API stable)
