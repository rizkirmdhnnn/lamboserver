# Phase 7: App Bundle & Icon - Pattern Map

**Mapped:** 2026-04-17
**Files analyzed:** 4 modified files + 1 new script + 1 replaced asset
**Analogs found:** 3 / 4 config files (1 analog from within-project; no analog for shell script or PNG asset)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `wails.json` | config | transform (template source) | `build/windows/info.json` | role-match (same template variable names, different format) |
| `build/darwin/Info.plist` | config | transform (Go template rendered at build time) | `build/darwin/Info.dev.plist` | exact (identical structure, same placeholder on line 11) |
| `build/darwin/Info.dev.plist` | config | transform (Go template rendered at `wails dev`) | `build/darwin/Info.plist` | exact (twins — identical structure) |
| `build/appicon.png` | asset | file-I/O | None | no analog (unique binary asset) |
| `build/darwin/LamboServer.iconset/` | asset | file-I/O | None | no analog (generated directory, new to project) |
| `scripts/generate-icns.sh` | utility | batch (sips resize → iconutil compile) | None | no analog (no shell scripts exist in project) |

---

## Pattern Assignments

### `wails.json` (config, transform)

**Analog:** `build/windows/info.json` (lines 1-18) — confirms the template variable names Wails uses across platforms

**Current state** (`wails.json` lines 1-13):
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "LamboServer",
  "outputfilename": "LamboServer",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Achmad Rizki Ramadhan",
    "email": "achmadrizkiramadhan0101@gmail.com"
  }
}
```

**Change required — add `info` block after `author`:**
```json
  "info": {
    "companyName": "LamboServer",
    "productName": "LamboServer",
    "productVersion": "1.0.0",
    "copyright": "\u00a9 2026 LamboServer",
    "comments": "Local web development environment manager for macOS"
  }
```

Field-to-template-variable mapping confirmed from `build/windows/info.json`:
- `productName` → `{{.Info.ProductName}}`
- `productVersion` → `{{.Info.ProductVersion}}`
- `copyright` → `{{.Info.Copyright}}`
- `comments` → `{{.Info.Comments}}`
- `companyName` → `{{.Info.CompanyName}}`

Note: Use the Unicode escape `\u00a9` for the copyright symbol to avoid UTF-8 encoding issues (see RESEARCH.md Pitfall 5).

---

### `build/darwin/Info.plist` (config, transform)

**Analog:** `build/darwin/Info.dev.plist` — identical structure (the two files are template twins)

**Current state of the target key** (line 10-11):
```xml
<key>CFBundleIdentifier</key>
<string>com.wails.{{.Name}}</string>
```

**Change required — replace line 11 only:**
```xml
<key>CFBundleIdentifier</key>
<string>dev.lamboserver.app</string>
```

**Context: surrounding keys that must NOT change** (lines 7-17):
```xml
<key>CFBundleName</key>
<string>{{.Info.ProductName}}</string>
<key>CFBundleExecutable</key>
<string>{{.OutputFilename}}</string>
<key>CFBundleIdentifier</key>
<string>dev.lamboserver.app</string>  <!-- CHANGED -->
<key>CFBundleVersion</key>
<string>{{.Info.ProductVersion}}</string>
<key>CFBundleGetInfoString</key>
<string>{{.Info.Comments}}</string>
<key>CFBundleShortVersionString</key>
<string>{{.Info.ProductVersion}}</string>
```

Critical constraint: `CFBundleIdentifier` has no Wails template variable — it must be a literal string. Do NOT use `{{.Info.BundleIdentifier}}` (that variable does not exist in Wails v2).

---

### `build/darwin/Info.dev.plist` (config, transform)

**Analog:** `build/darwin/Info.plist` — identical structure (the two files are template twins)

**Current state of the target key** (line 10-11 — identical to Info.plist):
```xml
<key>CFBundleIdentifier</key>
<string>com.wails.{{.Name}}</string>
```

**Change required — same as Info.plist, replace line 11 only:**
```xml
<key>CFBundleIdentifier</key>
<string>dev.lamboserver.app</string>
```

Note: `Info.dev.plist` has one additional block not present in `Info.plist` (lines 62-66):
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsLocalNetworking</key>
    <true/>
</dict>
```
This block must be preserved unchanged.

---

### `build/appicon.png` (asset, file-I/O)

**Analog:** None — this is a binary image asset replacement.

**Specification:**
- Dimensions: 1024x1024 pixels
- Format: PNG with RGBA channels
- Design: Lamborghini-inspired bull or shield silhouette, monochrome dark (black/dark gray on white background)
- Bold, simple shapes — fine internal lines will not survive sips downscaling to 16x16

**Generation approach:** Python 3 + Pillow (both confirmed installed at Python 3.13.0). Draw geometric primitives for the shield/bull silhouette programmatically.

**Validation:** `shasum build/appicon.png` — record before replacement, confirm hash differs after.

---

### `build/darwin/LamboServer.iconset/` (asset directory, batch)

**Analog:** None — this directory does not yet exist in the project.

**Required files (all derived from `build/appicon.png` via sips):**
```
icon_16x16.png      (16x16)
icon_16x16@2x.png   (32x32)
icon_32x32.png      (32x32)
icon_32x32@2x.png   (64x64)
icon_128x128.png    (128x128)
icon_128x128@2x.png (256x256)
icon_256x256.png    (256x256)
icon_256x256@2x.png (512x512)
icon_512x512.png    (512x512)
icon_512x512@2x.png (1024x1024)
```

Note: `icon_16x16@2x.png` and `icon_32x32.png` are BOTH 32x32 — this is intentional per Apple's iconset spec.

---

### `scripts/generate-icns.sh` (utility, batch)

**Analog:** None — no shell scripts exist in the project.

**Complete pattern** (from RESEARCH.md Code Examples, verified against machine availability):
```bash
#!/bin/bash
set -e

SRC="build/appicon.png"
ICONSET="build/darwin/LamboServer.iconset"
ICNS_OUT="build/darwin/iconfile.icns"

mkdir -p "$ICONSET"

sips -z 16   16   "$SRC" --out "$ICONSET/icon_16x16.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_16x16@2x.png"
sips -z 32   32   "$SRC" --out "$ICONSET/icon_32x32.png"
sips -z 64   64   "$SRC" --out "$ICONSET/icon_32x32@2x.png"
sips -z 128  128  "$SRC" --out "$ICONSET/icon_128x128.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_128x128@2x.png"
sips -z 256  256  "$SRC" --out "$ICONSET/icon_256x256.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_256x256@2x.png"
sips -z 512  512  "$SRC" --out "$ICONSET/icon_512x512.png"
sips -z 1024 1024 "$SRC" --out "$ICONSET/icon_512x512@2x.png"

iconutil -c icns "$ICONSET" -o "$ICNS_OUT"

echo "Generated: $ICNS_OUT"
```

Tool paths confirmed on this machine: `/usr/bin/sips`, `/usr/bin/iconutil`.

---

## Shared Patterns

### Go Template Variables in Wails Config Files
**Source:** `build/windows/info.json` (lines 4-16) and `build/darwin/Info.plist` (lines 7-25)
**Apply to:** `wails.json` (source) and both plist files (consumers)

The same variable names appear in `build/windows/info.json`, confirming they are the canonical Wails v2 field names:
```json
"ProductVersion": "{{.Info.ProductVersion}}",
"CompanyName": "{{.Info.CompanyName}}",
"FileDescription": "{{.Info.ProductName}}",
"LegalCopyright": "{{.Info.Copyright}}",
"ProductName": "{{.Info.ProductName}}",
"Comments": "{{.Info.Comments}}"
```
These map 1:1 to the `info` block keys in `wails.json`.

### CFBundleIdentifier Literal String Rule
**Source:** Research verified — Wails v2 does NOT provide a `{{.Info.BundleIdentifier}}` template variable.
**Apply to:** Both `build/darwin/Info.plist` and `build/darwin/Info.dev.plist`

The placeholder `com.wails.{{.Name}}` on line 11 of both plist files resolves to `com.wails.LamboServer` at build time. It must be replaced with the literal string `dev.lamboserver.app` in both files.

### Plist Template Structure (Do Not Modify)
**Source:** `build/darwin/Info.plist` lines 1-63, `build/darwin/Info.dev.plist` lines 1-68
**Apply to:** Both plist files — only line 11 changes; all other lines stay exactly as-is

The files share identical structure lines 1-62. `Info.dev.plist` has an additional `NSAppTransportSecurity` block at lines 62-66 that must not be removed.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `build/appicon.png` (replacement) | asset | file-I/O | Binary image asset — no code analog; design spec comes from CONTEXT.md D-04/D-05 |
| `build/darwin/LamboServer.iconset/` | asset | file-I/O | Generated directory — no iconset directories exist in project yet |
| `scripts/generate-icns.sh` | utility | batch | No shell scripts exist anywhere in the project |

---

## Metadata

**Analog search scope:** `/Users/rizkirmdhn/Documents/Code/lamboserver` (project root, excluding node_modules)
**Files scanned:** `wails.json`, `build/darwin/Info.plist`, `build/darwin/Info.dev.plist`, `build/windows/info.json`, `build/README.md`
**Shell scripts found:** 0 (Glob `**/*.sh` returned no results)
**Pattern extraction date:** 2026-04-17

**Key observations:**
1. `Info.plist` and `Info.dev.plist` are structurally identical templates — changes to one must mirror to the other.
2. `build/windows/info.json` is the only existing analog that confirms `wails.json` `info` block field names.
3. The `©` symbol in the copyright string should use the JSON unicode escape `\u00a9` to guarantee safe UTF-8 handling.
4. No shell scripts exist in the project — `scripts/generate-icns.sh` will be the first; it should follow the pattern in RESEARCH.md Code Examples verbatim.
5. Wails reads `build/appicon.png` as the icon source AND will use a pre-placed `build/darwin/iconfile.icns` if present — replace both to guarantee correct icon in all Wails behaviors.
