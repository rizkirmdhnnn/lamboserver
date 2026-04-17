---
phase: 5
slug: remove-cloudbeaver
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-17
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Shell commands (absence verification — pure removal phase) |
| **Config file** | none |
| **Quick run command** | `grep -ri cloudbeaver app.go internal/ frontend/src/ docs/` (expect zero results) |
| **Full suite command** | `go build ./... && grep -ri cloudbeaver app.go internal/ frontend/src/ docs/` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run absence grep
- **After every plan wave:** Run `go build ./...` + absence grep
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | CLEAN-01 | T-05-01 | Removal completeness | absence | `test ! -d internal/services/cloudbeaver && ! grep -qi cloudbeaver app.go && go build ./...` | N/A | ✅ green |
| 05-01-02 | 01 | 1 | CLEAN-01 | T-05-01 | Removal completeness | absence | `! grep -qi cloudbeaver frontend/src/pages/PgDatabasePage.tsx && ! grep -qi cloudbeaver docs/DEVELOPMENT.md` | N/A | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. This is a pure removal phase — verification is absence-based (grep + build), not behavioral. No test files needed.

---

## Manual-Only Verifications

All phase behaviors have automated verification. Removal completeness is verified by shell commands that confirm zero references exist.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** complete

---

## Validation Audit 2026-04-17

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

**Note:** Phase 5 is a pure removal phase. Verification is absence-based — confirming deleted code does not exist via grep/build commands. No unit tests are applicable since there is no new behavioral code to test.
