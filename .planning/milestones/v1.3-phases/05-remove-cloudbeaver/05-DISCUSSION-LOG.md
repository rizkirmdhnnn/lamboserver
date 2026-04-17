# Phase 5: Remove CloudBeaver - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-16
**Phase:** 05-remove-cloudbeaver
**Areas discussed:** PostgreSQL page after removal, Config & port cleanup, pgAdmin dead code

---

## PostgreSQL Page After Removal

| Option | Description | Selected |
|--------|-------------|----------|
| Remove the panel entirely | No web admin section at all — just PostgreSQL service controls. Phase 6 adds pgweb panel fresh. | ✓ |
| Show a placeholder | Keep the panel area but show 'No web admin installed' with a hint that pgweb is coming. | |
| You decide | Claude picks the best approach based on codebase patterns and Phase 6 needs. | |

**User's choice:** Remove the panel entirely
**Notes:** Cleanest removal — Phase 6 will add pgweb UI from scratch.

---

## Config & Port Cleanup

| Option | Description | Selected |
|--------|-------------|----------|
| Remove entirely | Delete DefaultCloudBeaverPort and all CloudBeaver config. Clean slate for Phase 6. | ✓ |
| Keep as comments | Comment out the port constant so Phase 6 can reference it. | |
| You decide | Claude picks based on whether Phase 6 would benefit from knowing the old port. | |

**User's choice:** Remove entirely
**Notes:** pgweb will define its own port in Phase 6 — no need to preserve CloudBeaver's port.

---

## pgAdmin Dead Code

| Option | Description | Selected |
|--------|-------------|----------|
| Remove it too | pgAdmin was replaced in v1.0 but directory still lingers. Remove alongside CloudBeaver. | ✓ |
| Leave it alone | Keep scope strictly to CloudBeaver. pgAdmin cleanup separate if needed. | |
| You decide | Claude checks if actually dead code and removes if so. | |

**User's choice:** Remove it too
**Notes:** Clean sweep — both pgAdmin and CloudBeaver packages removed in this phase.

---

## Claude's Discretion

- Ordering of removal steps (backend-first vs frontend-first)
- Whether to consolidate into one plan or split into backend/frontend plans
- Handling of test files referencing CloudBeaver or pgAdmin

## Deferred Ideas

None — discussion stayed within phase scope.
