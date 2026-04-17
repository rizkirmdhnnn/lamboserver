# Phase 1: Remove pgAdmin - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-16
**Phase:** 01-remove-pgadmin
**Areas discussed:** Wails binding cleanup, Test verification

---

## Wails Binding Cleanup

| Option | Description | Selected |
|--------|-------------|----------|
| Regenerate with Wails | Run `wails generate module` after removing Go methods — Wails rebuilds App.d.ts and App.js automatically, guaranteeing no stale bindings | ✓ |
| Delete manually | Surgically remove pgAdmin-specific lines from App.d.ts and App.js by hand — faster but risks missing a reference | |
| You decide | Claude picks the safest approach during execution | |

**User's choice:** Regenerate with Wails (Recommended)
**Notes:** User chose the safest approach — let Wails tooling handle binding regeneration rather than manual editing.

---

## Test Verification

| Option | Description | Selected |
|--------|-------------|----------|
| Build + full test suite | Run `go build ./...` AND `go test ./...` to catch both compile errors and transitive test breakage | ✓ |
| Build only | Just `go build ./...` — faster, but won't catch tests that import or reference pgAdmin indirectly | |
| Build + targeted tests | Run `go build ./...` plus tests only in packages that were modified (app, system, services) | |

**User's choice:** Build + full test suite (Recommended)
**Notes:** User wants thorough verification — full test suite to ensure no transitive breakage from pgAdmin removal.

---

## Claude's Discretion

- Deletion order across files (optimize for clean intermediate build states)

## Deferred Ideas

None — discussion stayed within phase scope.
