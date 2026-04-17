# Phase 8: Production Build - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 08-production-build
**Areas discussed:** PATH fix strategy, Build configuration, Testing approach

---

## PATH Fix Strategy

### How to fix PATH resolution

| Option | Description | Selected |
|--------|-------------|----------|
| Probe known paths | Check /opt/homebrew/bin and /usr/local/bin in BinaryLocator | |
| Set cmd.Env per exec | Inject PATH on every exec.Command call | |
| Not needed | All binaries are in ~/.lamboserver/, no Homebrew dependency | ✓ |

**User's choice:** Not needed — user clarified that all binaries are downloaded directly (MySQL from cdn.mysql.com, PHP from herdphp.com, etc.), not installed via Homebrew.
**Notes:** Codebase confirmed: zero Homebrew references, all BinaryLocators use LocalPath in ~/.lamboserver/

---

## Build Configuration

### Build target

| Option | Description | Selected |
|--------|-------------|----------|
| Universal (darwin/universal) | Works on both Intel and Apple Silicon | ✓ |
| Current arch only | Smaller, faster build | |

**User's choice:** Universal binary
**Notes:** Standard for macOS distribution

---

## Testing Approach

### Validation level

| Option | Description | Selected |
|--------|-------------|----------|
| Build + launch test | Run wails build, open .app, verify launch | ✓ |
| Build only | Just ensure wails build succeeds | |
| You decide | Claude determines | |

**User's choice:** Build + launch test
**Notes:** Manual verification that .app launches with correct icon and metadata

---

## Claude's Discretion

- Makefile target vs direct wails command
- LSMinimumSystemVersion adjustment

## Deferred Ideas

None
