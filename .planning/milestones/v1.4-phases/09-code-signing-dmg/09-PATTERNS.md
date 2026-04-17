# Phase 9: Code Signing & DMG - Pattern Map

**Mapped:** 2026-04-17
**Files analyzed:** 2 new files (scripts/build-dmg.sh, .gitignore update)
**Analogs found:** 2 / 2

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `scripts/build-dmg.sh` | build script / utility | batch (sequential pipeline) | `scripts/generate-icns.sh` | role-match (same script role, same project, same shell conventions) |
| `.gitignore` (append `*.dmg`) | config | — | existing `.gitignore` | exact (append one line) |

---

## Pattern Assignments

### `scripts/build-dmg.sh` (build script, batch pipeline)

**Analog:** `scripts/generate-icns.sh`

**Shebang + error-exit pattern** (`scripts/generate-icns.sh` lines 1-2):
```bash
#!/bin/bash
set -e
```
Copy this exactly. `set -e` means any non-zero exit aborts the script immediately — matching the existing project convention.

**Variable declaration pattern** (`scripts/generate-icns.sh` lines 4-6):
```bash
SRC="build/appicon.png"
ICONSET="build/darwin/LamboServer.iconset"
ICNS_OUT="build/darwin/iconfile.icns"
```
Declare all path/name variables at the top of the script before any commands. No `export`, no subshells — plain assignment. Apply this same pattern in build-dmg.sh with:
```bash
APP_NAME="LamboServer"
VERSION="1.0.0"
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"
```

**Progress echo pattern** (`scripts/generate-icns.sh` line 22):
```bash
echo "Generated: $ICNS_OUT"
```
One `echo` at the end reporting what was produced. In build-dmg.sh, extend this to numbered progress steps since the pipeline has 4 stages:
```bash
echo "[1/4] Building universal .app..."
# ... step ...
echo "[2/4] Ad-hoc signing with entitlements..."
# ... step ...
echo "[3/4] Packaging DMG..."
# ... step ...
echo "[4/4] Done. Output: ${DMG_OUT}"
```

**Sequential commands (no functions, no subshells)** (`scripts/generate-icns.sh` lines 8-21):
```bash
mkdir -p "$ICONSET"
sips -z 16 16 "$SRC" --out "$ICONSET/icon_16x16.png"
# ... more sips calls ...
iconutil -c icns "$ICONSET" -o "$ICNS_OUT"
```
No functions, no conditionals, no loops — plain sequential tool invocations. `set -e` handles error propagation. Replicate this flat structure in build-dmg.sh.

**Core pipeline pattern** (from RESEARCH.md Architecture Patterns — no codebase analog exists; use research pattern directly):

Step 1 — Build:
```bash
wails build -platform darwin/universal -clean
```

Step 2 — Ad-hoc sign with entitlements (codesign pattern from RESEARCH.md Pattern 2):
```bash
codesign \
  --force \
  --deep \
  -s - \
  --entitlements "${ENTITLEMENTS}" \
  "${APP_PATH}"
codesign --verify --deep "${APP_PATH}"
echo "  Signature verified"
```
- `--force`: Required — wails build produces a pre-signed bundle; without `--force`, codesign exits non-zero on an already-signed bundle
- `--deep`: Correct for ad-hoc signing; signs nested binaries inside Contents/MacOS/
- `-s -`: Ad-hoc identity (dash = no certificate; no Apple Developer account required)
- `--entitlements`: Must be included — without it, entitlements are not embedded (current build verified to have empty entitlements dict)

Step 3 — DMG staging + hdiutil (RESEARCH.md Pattern 1):
```bash
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
```
- `mktemp -d`: Always creates staging in `/tmp/` — avoids "resource busy" if staging were inside a mounted volume (PITFALLS.md Pitfall 4)
- `trap "rm -rf ${STAGING}" EXIT`: Guarantees cleanup even when `set -e` exits early on error
- `ln -s /Applications`: Creates the drag-to-Applications alias Finder displays inside the DMG
- `-ov`: Overwrites existing DMG — idempotent, safe for repeated builds
- `-format UDZO`: Zlib-compressed read-only DMG — standard distribution format
- `"${DMG_OUT}"` resolves to project root, NOT `build/bin/` — `wails build -clean` deletes `build/bin/` on next run (PITFALLS.md Pitfall 2)

**Complete assembled script** (for planner reference):
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

---

### `.gitignore` (config, append)

**Analog:** existing `/Users/rizkirmdhn/Documents/Code/lamboserver/.gitignore` (lines 1-6):
```
build/bin/
node_modules
frontend/dist
Raw*/
*.folder/
.DS_Store
```

**Pattern:** One artifact pattern per line, no comments, no blank lines between entries. Append `*.dmg` as a new line to exclude DMG build artifacts from version control. DMG files are binary build artifacts (same category as `build/bin/`) and must not be committed.

**Append target** (after line 6):
```
*.dmg
```

---

## Shared Patterns

### set -e Error Propagation
**Source:** `scripts/generate-icns.sh` line 2
**Apply to:** `scripts/build-dmg.sh`
```bash
set -e
```
Every command in the pipeline is a gate. If any step fails (wails build error, codesign error, hdiutil error), the script exits immediately with the failing command's exit code. No explicit error handling needed per step.

### trap for Cleanup on Error
**Source:** RESEARCH.md Code Examples (no codebase analog — pattern is additive over generate-icns.sh)
**Apply to:** `scripts/build-dmg.sh` — the staging directory creation block only
```bash
STAGING=$(mktemp -d)
trap "rm -rf ${STAGING}" EXIT
```
Place `trap` immediately after `mktemp -d`. This ensures the temporary staging directory is deleted whether the script exits cleanly or via `set -e` on a failed command. `generate-icns.sh` does not need this because `mkdir -p` creates a tracked path, not a temp dir — `build-dmg.sh` requires it because `mktemp -d` creates an untracked path.

### Quoted Variable References
**Source:** `scripts/generate-icns.sh` lines 8-22 (all path vars quoted)
**Apply to:** `scripts/build-dmg.sh`
```bash
sips -z 16 16 "$SRC" --out "$ICONSET/icon_16x16.png"
```
All variable references in command arguments use double quotes. Apply consistently throughout build-dmg.sh for all path variables (`"${APP_PATH}"`, `"${ENTITLEMENTS}"`, `"${STAGING}"`, `"${DMG_OUT}"`).

---

## No Analog Found

Files with no close match in the codebase — planner uses RESEARCH.md patterns directly:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| (codesign invocation) | build step | — | No codesign usage exists anywhere in the codebase; RESEARCH.md Pattern 2 is the reference |
| (hdiutil invocation) | build step | — | No hdiutil usage exists anywhere in the codebase; RESEARCH.md Pattern 1 is the reference |

Both of these are sub-steps within `scripts/build-dmg.sh`, not separate files. The overall script structure follows `generate-icns.sh`; only the individual tool calls are novel to this project.

---

## Anti-Patterns to Avoid (from RESEARCH.md)

These are documented in RESEARCH.md and must NOT appear in the implementation:

| Anti-Pattern | Correct Pattern |
|--------------|-----------------|
| Sign AFTER DMG creation | Sign .app before `hdiutil create` |
| `codesign` without `--entitlements` | Always include `--entitlements build/darwin/entitlements.plist` |
| `codesign` without `--force` | Always include `--force` — wails build pre-signs the bundle |
| `--options runtime` with `-s -` | Omit `--options runtime` — Hardened Runtime requires Developer ID |
| DMG output to `build/bin/` | Output to project root — `build/bin/` is deleted by `-clean` |
| `hdiutil attach` + manual copy loop | Use `hdiutil create -srcfolder` directly |

---

## Metadata

**Analog search scope:** `scripts/`, `build/darwin/`, `.gitignore`, `wails.json`
**Files scanned:** `scripts/generate-icns.sh`, `build/darwin/entitlements.plist`, `build/darwin/Info.plist`, `wails.json`, `.gitignore`
**Pattern extraction date:** 2026-04-17
