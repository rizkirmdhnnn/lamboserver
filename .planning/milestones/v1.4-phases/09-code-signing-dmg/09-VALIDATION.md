---
phase: 9
slug: code-signing-dmg
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-17
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Shell commands (codesign, hdiutil) — no test framework needed |
| **Config file** | none — validation is via CLI tool output |
| **Quick run command** | `codesign --verify --deep build/bin/LamboServer.app` |
| **Full suite command** | `codesign --verify --deep build/bin/LamboServer.app && test -f LamboServer-1.0.0.dmg` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `codesign --verify --deep build/bin/LamboServer.app`
- **After every plan wave:** Run full suite command
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | SIGN-01 | — | Ad-hoc signature applied | cli | `codesign -v build/bin/LamboServer.app` | N/A | ⬜ pending |
| 09-01-02 | 01 | 1 | SIGN-02 | — | Entitlements embedded | cli | `codesign -d --entitlements - build/bin/LamboServer.app` | N/A | ⬜ pending |
| 09-01-03 | 01 | 1 | SIGN-03 | — | Deep verify passes | cli | `codesign --verify --deep build/bin/LamboServer.app` | N/A | ⬜ pending |
| 09-02-01 | 02 | 1 | DMG-01 | — | DMG contains app + Applications | cli | `hdiutil attach LamboServer-1.0.0.dmg && ls /Volumes/LamboServer/` | N/A | ⬜ pending |
| 09-02-02 | 02 | 1 | DMG-02 | — | Build script runs full pipeline | cli | `bash scripts/build-dmg.sh` | N/A | ⬜ pending |
| 09-02-03 | 02 | 1 | DOC-01 | — | Gatekeeper instructions exist | grep | `grep -q "Open Anyway" RELEASE-NOTES.md` | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements — validation uses built-in macOS CLI tools.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| DMG drag-to-Applications works | DMG-01 | Requires Finder GUI interaction | Open DMG, drag app to Applications, verify it copies |
| Gatekeeper warning + workaround | DOC-01 | Requires fresh quarantine state | Download DMG from external source, open, follow workaround steps |

---

## Validation Sign-Off

- [ ] All tasks have automated verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
