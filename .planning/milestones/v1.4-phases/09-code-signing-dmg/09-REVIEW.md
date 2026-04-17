---
phase: 09-code-signing-dmg
reviewed: 2026-04-17T10:46:11Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - scripts/build-dmg.sh
  - .gitignore
  - RELEASE_NOTES.md
findings:
  critical: 0
  warning: 2
  info: 2
  total: 4
status: issues_found
---

# Phase 09: Code Review Report

**Reviewed:** 2026-04-17T10:46:11Z
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Three files were reviewed: the DMG build script (`scripts/build-dmg.sh`), the gitignore, and the release notes. The shell script is the only source file with reviewable logic. It is functionally correct when run from the project root but has two usability/correctness weaknesses: relative paths break silently if the script is invoked from a different working directory, and `set -u` is absent so undefined variables expand silently to empty strings. No security vulnerabilities or critical bugs were found. The gitignore and release notes are clean.

## Warnings

### WR-01: Relative paths break when script is not run from project root

**File:** `scripts/build-dmg.sh:6-8`
**Issue:** `APP_PATH`, `DMG_OUT`, and `ENTITLEMENTS` are all relative paths anchored to the working directory (`build/bin/`, `build/darwin/`). The script contains no guard to enforce that it is run from the project root. Running it from the `scripts/` directory (e.g., `cd scripts && ./build-dmg.sh`) causes all three paths to resolve incorrectly — `wails build` still writes to the correct location but `codesign` and `hdiutil` receive nonexistent paths, producing confusing errors.

**Fix:** Add a directory guard at the top of the script so it always operates relative to the repository root regardless of where it is invoked:

```bash
#!/bin/bash
set -e

# Resolve script location and cd to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.."
```

### WR-02: Missing `set -u` — undefined variables expand silently to empty string

**File:** `scripts/build-dmg.sh:2`
**Issue:** Only `set -e` is set. Without `set -u`, any typo in a variable name (e.g., `${ENTITLEMENS}` instead of `${ENTITLEMENTS}`) silently expands to an empty string. In the codesign invocation at line 14 this would produce `--entitlements ""`, which codesign accepts without error but ignores, silently producing a signing result without the intended entitlements. The resulting `.app` may behave differently from expected on macOS (e.g., missing hardened-runtime or sandbox entitlements).

**Fix:** Add `set -u` alongside `set -e`:

```bash
set -euo pipefail
```

`pipefail` is also recommended so that a failing command in a pipeline (e.g., if `hdiutil` were piped) does not silently succeed.

## Info

### IN-01: No pre-flight check that entitlements file exists

**File:** `scripts/build-dmg.sh:8,14`
**Issue:** `ENTITLEMENTS="build/darwin/entitlements.plist"` is used directly in the `codesign` call without verifying the file exists first. The entitlements file does currently exist in the repo, so this is not a bug today. However, if the file is accidentally deleted or the path changes, the error comes from `codesign` with a low-signal message rather than a clear pre-flight failure.

**Fix:** Add an existence check after the variables are defined:

```bash
if [[ ! -f "${ENTITLEMENTS}" ]]; then
  echo "ERROR: Entitlements file not found: ${ENTITLEMENTS}" >&2
  exit 1
fi
```

### IN-02: DMG output lands in working directory with no explicit path

**File:** `scripts/build-dmg.sh:7`
**Issue:** `DMG_OUT="${APP_NAME}-${VERSION}.dmg"` has no directory component. The DMG is written to wherever the script is invoked from. With the WR-01 fix applied the script will always run from the project root, so the DMG will land there — but this is implicit. A `dist/` or `release/` output directory would make intent clearer and prevent the DMG from being committed (`.gitignore` already excludes `*.dmg` at root, so this is not a security issue, just a clarity point).

**Fix:** Define an explicit output directory:

```bash
OUTPUT_DIR="dist"
mkdir -p "${OUTPUT_DIR}"
DMG_OUT="${OUTPUT_DIR}/${APP_NAME}-${VERSION}.dmg"
```

Update `.gitignore` to cover the subdirectory if needed (`dist/*.dmg` or the existing `*.dmg` glob already covers it).

---

_Reviewed: 2026-04-17T10:46:11Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
