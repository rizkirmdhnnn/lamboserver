#!/bin/bash
set -e

APP_NAME="LamboServer"
ENTS_DUMP=""
STAGING=""

# Consolidated EXIT handler: one trap registration for all cleanup concerns.
# NOTE: bash's `trap` does NOT accumulate — a second `trap ... EXIT` would
# silently replace this one (Pitfall 1). All future cleanup goes in this
# function, never in a second `trap` call.
cleanup() {
  [ -n "${STAGING}" ] && rm -rf "${STAGING}"
  [ -n "${ENTS_DUMP}" ] && rm -f "${ENTS_DUMP}"
  [ -f "wails.json.bak" ] && mv wails.json.bak wails.json
}
trap cleanup EXIT

# Step 1: Sync productVersion from VERSION env (CI path) or read from wails.json (local path).
# D-03/D-04/D-05: when VERSION is set, rewrite wails.json in place with backup;
# trap above restores wails.json on any exit. When VERSION is unset, fall back
# to the committed productVersion (current local-dev behavior).
if [ -n "${VERSION:-}" ]; then
  echo "[1/7] Syncing version from environment: ${VERSION}"
  cp wails.json wails.json.bak
  python3 -c "
import json, sys
with open('wails.json') as f:
    data = json.load(f)
data['info']['productVersion'] = sys.argv[1]
with open('wails.json', 'w') as f:
    json.dump(data, f, indent=2)
" "${VERSION}"
  echo "  wails.json productVersion set to ${VERSION}"
else
  VERSION=$(grep -o '"productVersion": *"[^"]*"' wails.json | grep -o '[0-9][0-9.]*')
  echo "[1/7] Using version from wails.json: ${VERSION}"
fi

APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "[2/7] Building universal .app..."
CGO_ENABLED=1 wails build -platform darwin/universal -clean

echo "[3/7] Verifying CGO tray linkage..."
otool -L "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "Cocoa.framework" || { echo "FAIL: Cocoa.framework not linked"; exit 1; }
echo "  Cocoa.framework linked"
lipo -info "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "x86_64" || { echo "FAIL: missing x86_64 arch"; exit 1; }
lipo -info "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "arm64" || { echo "FAIL: missing arm64 arch"; exit 1; }
echo "  Universal binary (arm64 + x86_64)"

echo "[4/7] Clearing quarantine and provenance attributes..."
xattr -cr "${APP_PATH}"
echo "  Extended attributes cleared"

echo "[5/7] Ad-hoc signing with entitlements..."
# NOTE: --options runtime (hardened runtime) is intentionally omitted. Hardened
# runtime is designed for Developer ID + notarization; combined with ad-hoc
# signing it triggers SIGKILL on launch via AMFI on macOS 26.x (error 163,
# "Launchd job spawn failed"). Signature + entitlements alone are sufficient.
codesign \
  --force \
  --sign - \
  --entitlements "${ENTITLEMENTS}" \
  "${APP_PATH}"
echo "  Signed: ${APP_PATH}"

# Gate 1: Signature integrity (SIGN-04)
codesign --verify --deep --strict "${APP_PATH}" || {
  echo "FAIL: codesign signature verification failed"
  exit 1
}
echo "  Gate 1 passed: signature integrity OK"

# Gate 2: Entitlements presence (SIGN-02)
ENTS_DUMP=$(mktemp)
codesign -d --entitlements "${ENTS_DUMP}" "${APP_PATH}" 2>/dev/null
REQUIRED_KEYS=(
  "com.apple.security.cs.allow-jit"
  "com.apple.security.cs.allow-unsigned-executable-memory"
  "com.apple.security.cs.disable-library-validation"
  "com.apple.security.automation.apple-events"
  "com.apple.security.network.client"
  "com.apple.security.network.server"
)
for key in "${REQUIRED_KEYS[@]}"; do
  grep -q "${key}" "${ENTS_DUMP}" || {
    echo "FAIL: entitlement missing from signed bundle: ${key}"
    exit 1
  }
done
echo "  Gate 2 passed: all 6 entitlements present"

echo "[6/7] Packaging DMG..."
# Uses sindresorhus/create-dmg (npm) — https://github.com/sindresorhus/create-dmg
# Zero-config: the tool auto-generates a polished drag-to-Applications DMG with the
# app icon composed into the background. No manual window size / icon coords / PNG assets.
# D-02: --no-code-sign — the .app is already ad-hoc signed; DMG container signing has
#   no Gatekeeper benefit for ad-hoc and would require a Developer ID cert we do not have.
# D-05: fail fast if create-dmg is missing (local dev: `npm install -g create-dmg`; CI installs it in its own step).
command -v create-dmg >/dev/null || { echo "FAIL: create-dmg not installed. Run: npm install -g create-dmg"; exit 1; }
# Clear xattrs on the source app so DMG contents are clean.
xattr -cr "${APP_PATH}"
# sindresorhus/create-dmg writes into the current working dir as "<App Name> <version>.dmg"
# (space, not dash). We build it into a temp dir and rename to ${DMG_OUT} to preserve the
# Phase 6 filename contract (D-06): LamboServer-<VERSION>.dmg + .sha256 sidecar.
rm -f "${DMG_OUT}"
STAGING=$(mktemp -d)
(cd "${STAGING}" && create-dmg --overwrite --no-code-sign "${OLDPWD}/${APP_PATH}")
# Locate the produced DMG (sindresorhus/create-dmg names it "<App Name> <version>.dmg")
PRODUCED_DMG=$(ls -1 "${STAGING}"/*.dmg 2>/dev/null | head -1)
test -n "${PRODUCED_DMG}" || { echo "FAIL: create-dmg did not produce a .dmg in ${STAGING}"; exit 1; }
mv "${PRODUCED_DMG}" "${DMG_OUT}"
# Gate: confirm final artifact exists at the expected filename (Phase 6 CI upload contract).
test -f "${DMG_OUT}" || { echo "FAIL: DMG rename to ${DMG_OUT} failed"; exit 1; }
echo "  DMG created: ${DMG_OUT}"

echo "[7/7] Generating SHA-256 checksum..."
shasum -a 256 "${DMG_OUT}" > "${DMG_OUT}.sha256"
echo "  Checksum file: ${DMG_OUT}.sha256"
echo "  $(cat "${DMG_OUT}.sha256")"

echo "Done. Output: ${DMG_OUT}"
echo ""
echo "To open after downloading:"
echo "  xattr -cr /Applications/${APP_NAME}.app"
echo "  Or: Right-click the app > Open > Open (bypasses Gatekeeper)"
