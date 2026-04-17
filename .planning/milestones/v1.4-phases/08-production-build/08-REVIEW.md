---
phase: 08-production-build
reviewed: 2026-04-17T00:00:00Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - build/darwin/entitlements.plist
  - build/darwin/Info.plist
  - build/darwin/Info.dev.plist
findings:
  critical: 0
  warning: 3
  info: 2
  total: 5
status: issues_found
---

# Phase 8: Code Review Report

**Reviewed:** 2026-04-17
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Three macOS build configuration files were reviewed: the new `entitlements.plist` for ad-hoc codesigning and the updated `Info.plist` / `Info.dev.plist` templates (both bumped from `10.13.0` to `12.0.0` minimum system version).

The `LSMinimumSystemVersion` change is correct and consistent across both plist templates. The entitlements file is well-formed and the chosen entitlement set is appropriate for a Wails app with ad-hoc signing and no sandbox. Three warnings and two info items were found: the most impactful is `NSAppleEventsUsageDescription` being placed in the entitlements file (where it is never read) rather than in `Info.plist` (where macOS looks for it), and a missing `NSAppTransportSecurity` block in the production `Info.plist` that exists in the dev template.

---

## Warnings

### WR-01: `NSAppleEventsUsageDescription` in wrong file — will never display

**File:** `build/darwin/entitlements.plist:17-18`
**Issue:** `NSAppleEventsUsageDescription` is placed inside the entitlements file. macOS reads `NS*UsageDescription` strings exclusively from `Info.plist` inside the app bundle. The entitlements file is consumed by `codesign` during signing and never consulted for UI permission dialogs. As a result, if macOS ever prompts the user to allow automation/AppleScript access, it will display a generic message instead of the human-readable string defined here. With ad-hoc signing this is currently masked, but it will surface if the project ever adopts Developer ID or App Store distribution.
**Fix:** Remove `NSAppleEventsUsageDescription` from `entitlements.plist` and add it to both `Info.plist` and `Info.dev.plist` inside the root `<dict>`:
```xml
<key>NSAppleEventsUsageDescription</key>
<string>LamboServer uses administrator dialogs to install system services.</string>
```

---

### WR-02: `NSHighResolutionCapable` is a string, not a Boolean

**File:** `build/darwin/Info.plist:23`, `build/darwin/Info.dev.plist:23`
**Issue:** The value for `NSHighResolutionCapable` is the string `"true"` rather than the Boolean type `<true/>`. Apple's documentation and the plist DTD define this as a `Boolean`. Using a string instead of a Boolean is technically malformed for this key. While macOS is permissive about this in practice today, it may behave unexpectedly on future releases or when validated by tools like `plutil` or App Store submission checks.
**Fix:** Replace the `<string>` tag with the plist Boolean literal in both files:
```xml
<!-- Before -->
<key>NSHighResolutionCapable</key>
<string>true</string>

<!-- After -->
<key>NSHighResolutionCapable</key>
<true/>
```

---

### WR-03: `NSAppTransportSecurity` present in dev plist but absent from production plist

**File:** `build/darwin/Info.dev.plist:62-66` (present), `build/darwin/Info.plist` (absent)
**Issue:** The dev template includes `NSAllowsLocalNetworking: true` under `NSAppTransportSecurity`, which permits WKWebView to load HTTP content over loopback. The production template omits this block entirely. Since LamboServer's core purpose is managing local services (pgweb, phpMyAdmin) reachable over `http://localhost:*`, the Wails WebView in the production build may be blocked from loading those local HTTP endpoints by App Transport Security. ATS is enforced in production builds even without sandboxing.
**Fix:** Add the same `NSAppTransportSecurity` block to `build/darwin/Info.plist` before the closing `</dict>`:
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsLocalNetworking</key>
    <true/>
</dict>
```

---

## Info

### IN-01: Missing XML declaration in `Info.plist` and `Info.dev.plist`

**File:** `build/darwin/Info.plist:1`, `build/darwin/Info.dev.plist:1`
**Issue:** Both plist templates are missing the XML declaration `<?xml version="1.0" encoding="UTF-8"?>` that opens `entitlements.plist` (and is present in well-formed Apple plist files per the DTD). `plutil` and Xcode tooling accept plist files without it, so this is not a build blocker, but it is a deviation from canonical Apple plist format.
**Fix:** Add as the first line of both files:
```xml
<?xml version="1.0" encoding="UTF-8"?>
```

---

### IN-02: Hardened Runtime entitlements have no effect under ad-hoc signing

**File:** `build/darwin/entitlements.plist:6-13`
**Issue:** `com.apple.security.cs.allow-jit`, `com.apple.security.cs.allow-unsigned-executable-memory`, and `com.apple.security.cs.disable-library-validation` are Hardened Runtime entitlements enforced only when an app is signed with a Developer ID or App Store certificate. Under ad-hoc signing (the current distribution method per project constraints), the OS ignores these flags entirely — the app runs without Hardened Runtime restrictions regardless. The entitlements are correct for future-proofing and will be needed if distribution ever moves to Developer ID, so no removal is necessary; this is informational only.
**Fix:** No action required. The comment in the file ("Required: WKWebView uses JIT on Apple Silicon") is accurate for the future Developer ID path. Consider adding a comment clarifying they are no-ops under ad-hoc: `<!-- No-op under ad-hoc; enforced only with Developer ID / App Store signing -->`.

---

_Reviewed: 2026-04-17_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
