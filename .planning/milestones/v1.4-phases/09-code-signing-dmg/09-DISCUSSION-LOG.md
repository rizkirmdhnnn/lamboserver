# Phase 9: Code Signing & DMG - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md -- this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 09-code-signing-dmg
**Areas discussed:** Build script design, DMG creation approach, DMG contents, Gatekeeper docs

---

## Build Script Design

| Option | Description | Selected |
|--------|-------------|----------|
| Shell script | A scripts/build-dmg.sh that runs: wails build -> codesign -> hdiutil. Matches existing scripts/generate-icns.sh pattern. | ✓ |
| Makefile | make build, make sign, make dmg, make release -- individual targets plus a combined one. | |
| You decide | Claude picks the best approach based on codebase patterns. | |

**User's choice:** Shell script
**Notes:** Matches existing pattern in scripts/generate-icns.sh

| Option | Description | Selected |
|--------|-------------|----------|
| Include icon generation | Script runs generate-icns.sh first, then build -> sign -> DMG. | |
| Skip icon step | Icon is already committed as iconfile.icns. Script starts at wails build. | ✓ |
| You decide | Claude picks based on whether iconfile.icns is tracked in git. | |

**User's choice:** Skip icon step
**Notes:** Icon already built and committed

---

## DMG Creation Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Plain hdiutil | Use macOS built-in hdiutil. No external dependencies. Simple and reliable. | ✓ |
| create-dmg (npm/brew) | Third-party tool with window sizing, icon positioning, background image support. | |
| You decide | Claude picks the simplest approach that meets DMG-01 requirements. | |

**User's choice:** Plain hdiutil
**Notes:** No external dependency needed

| Option | Description | Selected |
|--------|-------------|----------|
| LamboServer-1.0.0.dmg | Includes version number. Clear for users downloading releases. | ✓ |
| LamboServer.dmg | No version in filename. Simpler. | |
| You decide | Claude picks a naming convention. | |

**User's choice:** LamboServer-1.0.0.dmg
**Notes:** Version in filename for release clarity

---

## DMG Contents

| Option | Description | Selected |
|--------|-------------|----------|
| App + Applications only | Clean, standard macOS DMG. Just LamboServer.app and Applications alias. | ✓ |
| Include a README | Add README.txt with Gatekeeper workaround instructions inside DMG. | |
| Include LICENSE + README | Add both LICENSE and README.txt inside DMG. | |

**User's choice:** App + Applications only
**Notes:** Clean standard layout

---

## Gatekeeper Docs

| Option | Description | Selected |
|--------|-------------|----------|
| GitHub release notes | Workaround instructions in GitHub release description. Users see them on download page. | ✓ |
| README.md section | Add Installation section to project README. | |
| Both | Instructions in README.md AND GitHub release notes. | |

**User's choice:** GitHub release notes
**Notes:** Instructions visible on download page

| Option | Description | Selected |
|--------|-------------|----------|
| Just build the DMG | Script outputs DMG only. Release notes written manually. | ✓ |
| Generate a template | Script creates RELEASE-NOTES.md template with Gatekeeper instructions pre-filled. | |

**User's choice:** Just build the DMG
**Notes:** Release notes are a separate concern

---

## Claude's Discretion

- Exact hdiutil flags and volume name
- codesign flags beyond -s - and --entitlements
- Whether to include codesign --verify as part of build script

## Deferred Ideas

None
