#!/bin/bash
set -e

APP_NAME="LamboServer"
VERSION=$(grep -o '"productVersion": *"[^"]*"' wails.json | grep -o '[0-9][0-9.]*')
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"
BINARY="${APP_PATH}/Contents/MacOS/${APP_NAME}"

echo "[1/6] Building universal .app..."
CGO_ENABLED=1 wails build -platform darwin/universal -clean

echo "[2/6] Verifying CGO tray linkage..."
otool -L "$BINARY" | grep -q "Cocoa.framework" || { echo "FAIL: Cocoa.framework not linked"; exit 1; }
echo "  Cocoa.framework linked"
lipo -info "$BINARY" | grep -q "x86_64" || { echo "FAIL: missing x86_64 arch"; exit 1; }
lipo -info "$BINARY" | grep -q "arm64" || { echo "FAIL: missing arm64 arch"; exit 1; }
echo "  Universal binary (arm64 + x86_64)"

echo "[3/6] Clearing quarantine and provenance attributes..."
xattr -cr "${APP_PATH}"
echo "  Extended attributes cleared"

echo "[4/6] Ad-hoc signing with entitlements..."
# Sign the main binary directly (--deep is deprecated and unreliable)
codesign --force --sign - --entitlements "${ENTITLEMENTS}" --options runtime "${BINARY}"
# Sign the app bundle
codesign --force --sign - --entitlements "${ENTITLEMENTS}" "${APP_PATH}"
codesign --verify "${APP_PATH}"
echo "  Signature verified"

echo "[5/6] Packaging DMG..."
STAGING=$(mktemp -d)
trap "rm -rf ${STAGING}" EXIT
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
