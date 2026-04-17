---
phase: 7
slug: app-bundle-icon
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-17
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -v ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -v ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | BUNDLE-01 | — | N/A | manual | `cat wails.json \| grep productVersion` | ✅ | ⬜ pending |
| 07-01-02 | 01 | 1 | BUNDLE-02 | — | N/A | manual | `grep CFBundleIdentifier build/darwin/Info.plist` | ✅ | ⬜ pending |
| 07-01-03 | 01 | 1 | BUNDLE-03 | — | N/A | manual | `grep -c '{{.Info.' build/darwin/Info.plist` | ✅ | ⬜ pending |
| 07-02-01 | 02 | 1 | ICON-01 | — | N/A | manual | `file build/appicon.png` | ✅ | ⬜ pending |
| 07-02-02 | 02 | 1 | ICON-02 | — | N/A | manual | `iconutil --convert icns` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements.

*This phase is primarily build configuration and asset generation — no new test files needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Custom icon appears in Dock/Finder | ICON-01 | Visual verification | Build .app, open in Finder, check icon in Dock |
| Info.plist renders correct metadata | BUNDLE-03 | Build output inspection | Run `wails build`, inspect `build/bin/LamboServer.app/Contents/Info.plist` |
| Icon shows at all macOS sizes | ICON-02 | Visual verification | Check icon in Finder list, grid, and Get Info views |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
