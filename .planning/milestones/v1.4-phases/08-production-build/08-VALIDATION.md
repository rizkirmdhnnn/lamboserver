---
phase: 8
slug: production-build
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-17
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + manual verification |
| **Config file** | none — uses existing go test infrastructure |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -v ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -v ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | BUILD-01 | — | N/A | manual | `wails build -platform darwin/universal -clean` | ✅ | ⬜ pending |
| 08-01-02 | 01 | 1 | BUILD-01 | — | N/A | manual | `lipo -info build/bin/LamboServer.app/Contents/MacOS/LamboServer` | ✅ | ⬜ pending |
| 08-01-03 | 01 | 1 | BUILD-02 | — | N/A | manual | `open build/bin/LamboServer.app` | ✅ | ⬜ pending |
| 08-01-04 | 01 | 1 | BUILD-03 | — | N/A | manual | Service start/stop from GUI | ✅ | ⬜ pending |
| 08-01-05 | 01 | 1 | BUILD-01 | — | N/A | automated | `defaults read build/bin/LamboServer.app/Contents/Info.plist CFBundleIdentifier` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Universal binary builds | BUILD-01 | Requires actual `wails build` execution | Run `wails build -platform darwin/universal -clean`, verify `lipo -info` shows both arm64 and x86_64 |
| App launches from Finder | BUILD-02 | Requires GUI interaction | Double-click .app in Finder, verify services start without PATH errors |
| Service management works | BUILD-03 | Requires GUI interaction | Install/start/stop/restart services from app GUI |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
