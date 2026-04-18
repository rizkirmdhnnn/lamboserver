#!/bin/bash
set -e

APP_NAME="LamboServer"
VERSION=$(grep -o '"productVersion": *"[^"]*"' wails.json | grep -o '[0-9][0-9.]*')
APP_PATH="build/bin/${APP_NAME}.app"
DMG_OUT="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "[1/5] Building universal .app..."
CGO_ENABLED=1 wails build -platform darwin/universal -clean

echo "[2/5] Verifying CGO tray linkage..."
BINARY="${APP_PATH}/Contents/MacOS/${APP_NAME}"
otool -L "$BINARY" | grep -q "Cocoa.framework" || { echo "FAIL: Cocoa.framework not linked"; exit 1; }
echo "  Cocoa.framework linked"
lipo -info "$BINARY" | grep -q "x86_64" || { echo "FAIL: missing x86_64 arch"; exit 1; }
lipo -info "$BINARY" | grep -q "arm64" || { echo "FAIL: missing arm64 arch"; exit 1; }
echo "  Universal binary (arm64 + x86_64)"

echo "[3/5] Ad-hoc signing with entitlements..."
codesign --force --deep -s - --entitlements "${ENTITLEMENTS}" "${APP_PATH}"
codesign --verify --deep "${APP_PATH}"
echo "  Signature verified"

echo "[4/5] Packaging DMG..."
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

echo "[5/5] Done. Output: ${DMG_OUT}"
