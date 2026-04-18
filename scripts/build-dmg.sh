#!/bin/bash
set -e

APP_NAME="LamboServer"
VERSION=$(grep -o '"productVersion": *"[^"]*"' wails.json | grep -o '[0-9][0-9.]*')
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "[1/6] Building universal .app..."
CGO_ENABLED=1 wails build -platform darwin/universal -clean

echo "[2/6] Verifying CGO tray linkage..."
otool -L "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "Cocoa.framework" || { echo "FAIL: Cocoa.framework not linked"; exit 1; }
echo "  Cocoa.framework linked"
lipo -info "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "x86_64" || { echo "FAIL: missing x86_64 arch"; exit 1; }
lipo -info "${APP_PATH}/Contents/MacOS/${APP_NAME}" | grep -q "arm64" || { echo "FAIL: missing arm64 arch"; exit 1; }
echo "  Universal binary (arm64 + x86_64)"

echo "[3/6] Clearing quarantine and provenance attributes..."
xattr -cr "${APP_PATH}"
echo "  Extended attributes cleared"

echo "[4/6] Ad-hoc signing with entitlements..."
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

echo "[5/6] Packaging DMG..."
STAGING=$(mktemp -d)
trap "rm -f ${ENTS_DUMP}; rm -rf ${STAGING}" EXIT
cp -r "${APP_PATH}" "${STAGING}/"
# Clear xattrs on the staging copy so DMG contents are clean
xattr -cr "${STAGING}/${APP_NAME}.app"
ln -s /Applications "${STAGING}/Applications"
hdiutil create \
  -volname "${APP_NAME}" \
  -srcfolder "${STAGING}" \
  -ov \
  -format UDZO \
  "${DMG_OUT}"

echo "[6/6] Done. Output: ${DMG_OUT}"
echo ""
echo "To open after downloading:"
echo "  xattr -cr /Applications/${APP_NAME}.app"
echo "  Or: Right-click the app > Open > Open (bypasses Gatekeeper)"
