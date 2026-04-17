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
