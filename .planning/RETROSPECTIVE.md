# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.3 — Replace CloudBeaver with pgweb

**Shipped:** 2026-04-17
**Phases:** 2 | **Plans:** 3

### What Was Built
- Fully removed CloudBeaver from Go backend, frontend, config, and docs (Phase 5)
- pgweb daemon lifecycle package with exec.Command-based process management (Phase 6)
- React Web Admin card with install/start/stop/open controls and three-state rendering (Phase 6)
- D-08 auto-stop coupling: pgweb stops when PostgreSQL stops (both shutdown and StopService paths)

### What Worked
- Clean phase separation: removal (Phase 5) then addition (Phase 6) avoided merge conflicts
- DaemonWebAdminService pattern with FileSystem/CommandRunner interfaces enabled thorough unit testing
- Absence-based verification for removal phase was fast and definitive (grep + build)
- 15 unit tests with testify mocks provided solid coverage without requiring running processes

### What Was Inefficient
- Phase 6 SUMMARY.md frontmatter missing `requirements_completed` field — caused blank one-liners in auto-generated MILESTONES.md
- `config.DefaultPgwebPort` declared but never consumed — DRY violation across three locations (config, manager, app.go)
- `app.go OpenPgweb()` hardcodes URL instead of calling `a.Pgweb.URL()` — two sources of truth
- `project-structure.md` documentation drift (14 stale CloudBeaver references)

### Patterns Established
- exec.Command + cmd.Start() for non-blocking daemon process management
- syscall.Signal(0) probe for liveness checking
- Port preflight via net.Listen before starting services
- Type alias exports for cross-package constructor ergonomics in Go

### Key Lessons
1. Removal phases are best verified by absence checks (grep returns zero), not by writing tests for deleted code
2. When adding a new service binary, the install/start/stop lifecycle should be fully testable with interface mocks
3. SUMMARY.md frontmatter fields need to be complete — downstream tooling depends on them

### Cost Observations
- Model mix: ~70% sonnet (agents), ~30% opus (orchestration)
- Notable: Phase 5 removal completed in ~3 minutes; Phase 6 addition in ~22 minutes total

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 2 | 3 | Initial pgAdmin removal, CloudBeaver verification |
| v1.2 | 2 | 4 | Bug fix milestone — auto-clean, toast notifications, status detection |
| v1.3 | 2 | 3 | Service replacement — removal + addition pattern with Nyquist validation |

### Cumulative Quality

| Milestone | Tests Added | Key Quality Gate |
|-----------|------------|-----------------|
| v1.0 | 0 | Manual verification only |
| v1.2 | 0 | Manual verification + integration testing |
| v1.3 | 15 | Nyquist validation (automated + manual gates) |

### Top Lessons (Verified Across Milestones)

1. Phase separation (remove then add) prevents cross-contamination and simplifies verification
2. Desktop app testing requires both automated unit tests and manual UI verification — neither alone is sufficient
