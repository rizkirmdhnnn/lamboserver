# Milestones

## v1.4 macOS Installer (Shipped: 2026-04-17)

**Phases completed:** 3 phases, 6 plans, 3 tasks

**Key accomplishments:**

- One-liner:
- One-liner:
- One-liner:
- Universal binary LamboServer.app built via `wails build -platform darwin/universal -clean`: arm64+x86_64 fat binary, dev.lamboserver.app bundle ID, 1.0.0 version, LSMinimumSystemVersion 12.0.0, custom icon present.
- One-liner:
- Full pipeline executed: wails build -> ad-hoc codesign with entitlements -> hdiutil UDZO DMG; RELEASE_NOTES.md with macOS Sequoia Gatekeeper workaround ready for GitHub release.

---

## v1.3 Replace CloudBeaver with pgweb (Shipped: 2026-04-17)

**Phases completed:** 2 phases, 3 plans, 5 tasks

**Key accomplishments:**

- Deleted CloudBeaver package, removed all references from Go backend, frontend, config, and docs — zero "cloudbeaver" matches in source files
- Added pgweb daemon lifecycle (install/start/stop/status via exec.Command) with D-08 auto-stop coupling and 5 Wails IPC bindings
- Added React Web Admin card for pgweb with three states (not-installed/stopped/running) on the PostgreSQL database page
- 15 unit tests covering pgweb package (install, start, stop, isRunning, checkPort, status)

---

## v1.2 Fix MySQL and PostgreSQL Installation Bugs (Shipped: 2026-04-16)

**Phases completed:** 2 phases, 4 plans

**Key accomplishments:**

- MySQL InitDataDir auto-clean on failure — retry works without manual data dir cleanup
- PostgreSQL InitDataDir auto-clean on failure — same pattern
- Toast error notification component (reusable across both database pages)
- Step-by-step install progress in button text (Downloading → Initializing → ✓ Installed)
- StatusNotInstalled enum — service adapters distinguish "not installed" from "stopped"
- Fixed status detection bugs in DatabasePage, PgDatabasePage, and CloudBeaver

---

## v1.0 Remove pgAdmin, Replace with CloudBeaver (Shipped: 2026-04-16)

**Git tag:** v1.1
**Phases completed:** 2 phases, 3 plans

**Key accomplishments:**

- Deleted all pgAdmin Go code (4 files, ~440 lines) and stripped every reference from app.go and paths.go
- Removed stale OpenPgAdmin declarations from Wails TypeScript and JavaScript bindings
- Verified CloudBeaver as sole PostgreSQL admin interface with correct connection presets (localhost:5432)
- Updated docs/DEVELOPMENT.md with complete services/ directory tree

---
