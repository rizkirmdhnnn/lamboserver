# Roadmap: LamboServer

## Milestones

- ✅ **v1.0 System Tray Integration** — Phases 1-4 (shipped 2026-04-18)
- 🚧 **v1.1 Build Pipeline Overhaul** — Phases 5-7 (in progress)

## Phases

<details>
<summary>✅ v1.0 System Tray Integration (Phases 1-4) — SHIPPED 2026-04-18</summary>

- [x] **Phase 1: Tray Foundation & Window Lifecycle** — completed 2026-04-17
- [x] **Phase 2: Service Controls** — completed 2026-04-18
- [x] **Phase 3: Quick Access Links** — completed 2026-04-18
- [x] **Phase 4: Build Pipeline** — completed 2026-04-18

</details>

### 🚧 v1.1 Build Pipeline Overhaul (In Progress)

**Milestone Goal:** Fix the "app is damaged" launch error and overhaul the build-to-distribution pipeline for reliable, professional DMG releases.

- [x] **Phase 5: Signing Correctness** - Fix codesign invocation, verify entitlements, audit embedded binaries (completed 2026-04-18)
- [x] **Phase 6: CI Pipeline Hardening** - Pin tools, enable caching, sync version, add SHA-256 artifact and release docs (completed 2026-04-18)
- [ ] **Phase 7: Professional DMG Appearance** - Custom background, drag-to-Applications layout, Retina support

## Phase Details

### Phase 5: Signing Correctness
**Goal**: The app bundle is correctly signed so it launches on macOS Sequoia without a "damaged" error
**Depends on**: Phase 4 (v1.0 Build Pipeline)
**Requirements**: SIGN-01, SIGN-02, SIGN-03, SIGN-04
**Success Criteria** (what must be TRUE):
  1. App launches on macOS Sequoia without triggering the "app is damaged" error after quarantine clearance
  2. `codesign --verify --deep --strict` reports a valid signature on the built `.app` bundle
  3. Entitlements are present and verified by CI — `codesign -d --entitlements` gate passes before DMG creation
  4. Embedded nginx and dnsmasq binaries are ad-hoc codesigned for ARM64/universal at extraction time
**Plans:** 2/2 plans complete
Plans:
- [x] 05-01-PLAN.md — Fix build script signing and add runtime binary codesigning
- [x] 05-02-PLAN.md — Simplify CI pipeline to call build-dmg.sh

### Phase 6: CI Pipeline Hardening
**Goal**: The GitHub Actions release pipeline is deterministic, cached, and produces a verifiable artifact with user-facing installation guidance
**Depends on**: Phase 5
**Requirements**: CI-01, CI-02, CI-03, CI-04, CI-05, CI-06, CI-07
**Success Criteria** (what must be TRUE):
  1. CI runs on a pinned `macos-15` runner with Wails CLI locked to `v2.12.0`
  2. Go modules and npm packages are restored from cache — cold build is the exception, not the rule
  3. The DMG release artifact is accompanied by a `SHA-256` checksum file for download verification
  4. GitHub Release notes include macOS Sequoia first-launch instructions (Privacy & Security flow)
  5. The app version shown in release assets matches the git tag that triggered the build
**Plans:** 3/3 plans complete
Plans:
- [x] 06-01-ci-runner-caching-PLAN.md — Pin runner + Wails CLI, enable Go/npm caching, export VERSION to build-dmg.sh
- [x] 06-02-build-script-version-checksum-PLAN.md — Extend build-dmg.sh with version-sync and SHA-256 checksum; consolidate cleanup trap
- [x] 06-03-release-notes-composition-PLAN.md — Create RELEASE_NOTES_TEMPLATE.md; add CI Compose step; wire softprops body_path + two-file upload

### Phase 7: Professional DMG Appearance
**Goal**: The DMG installer presents a polished drag-to-Applications experience on Retina displays using `sindresorhus/create-dmg`'s auto-composed visual
**Depends on**: Phase 6
**Requirements**: DMG-01, DMG-02, DMG-03
**Success Criteria** (what must be TRUE):
  1. Opening the DMG shows a polished drag-to-Applications visual auto-composed by `sindresorhus/create-dmg` from the app icon
  2. The Applications alias and app icon are laid out by the tool's defaults (no manual coordinates required)
  3. Output file preserves the Phase 6 filename contract (`LamboServer-<VERSION>.dmg` + `.sha256` sidecar)
**Plans:** 3 plans (revised post-implementation — see 07-CONTEXT.md Revision Note)
Plans:
- [x] 07-01-PLAN.md — (retired) Background PNG generator — superseded by sindresorhus/create-dmg auto-composition
- [x] 07-02-PLAN.md — Swap build-dmg.sh packaging step to sindresorhus/create-dmg + filename rename
- [x] 07-03-PLAN.md — Add `npm install -g create-dmg` CI step and README one-liner
**UI hint**: yes

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Tray Foundation & Window Lifecycle | v1.0 | 2/2 | Complete | 2026-04-17 |
| 2. Service Controls | v1.0 | 2/2 | Complete | 2026-04-18 |
| 3. Quick Access Links | v1.0 | 2/2 | Complete | 2026-04-18 |
| 4. Build Pipeline | v1.0 | 1/1 | Complete | 2026-04-18 |
| 5. Signing Correctness | v1.1 | 2/2 | Complete   | 2026-04-18 |
| 6. CI Pipeline Hardening | v1.1 | 3/3 | Complete    | 2026-04-18 |
| 7. Professional DMG Appearance | v1.1 | 3/3 | Complete    | 2026-04-18 |
