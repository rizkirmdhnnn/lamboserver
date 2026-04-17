# Phase 6: Add pgweb - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-16
**Phase:** 06-add-pgweb
**Areas discussed:** UI placement & controls, Lifecycle coupling, Installation experience, Browser open behavior

---

## UI Placement & Controls

| Option | Description | Selected |
|--------|-------------|----------|
| Separate card below service | Dedicated 'Web Admin' card below PostgreSQL service card | ✓ |
| Integrated in service card | Add pgweb buttons directly into PostgreSQL service card row | |
| You decide | Claude picks based on existing patterns | |

**User's choice:** Separate card below service
**Notes:** Matches how phpMyAdmin would appear on the MySQL page. Keeps pgweb visually distinct.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Compact install prompt | Small card with 'Not installed' text and [Install] button | ✓ |
| Full empty state | Larger card with icon, description, and [Install] button | |
| You decide | Claude picks | |

**User's choice:** Compact install prompt
**Notes:** Minimal, no lengthy explanation needed.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Only when PostgreSQL is running | Web Admin card appears only when PG is running | ✓ |
| Always visible | Card always visible when PostgreSQL is installed | |
| You decide | Claude decides | |

**User's choice:** Only when PostgreSQL is running
**Notes:** No point showing pgweb controls if the database isn't up.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Status + Start button | Stopped status dot with [Start] button | ✓ |
| Start + Open combined | [Start] and disabled [Open] button | |
| You decide | Claude picks | |

**User's choice:** Status + Start button
**Notes:** Similar to how PostgreSQL itself shows when stopped.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Status + Stop + Open | Running status with port, [Stop] and [Open] buttons | ✓ |
| Compact with just Open | Running dot + [Open in Browser] only | |
| You decide | Claude picks | |

**User's choice:** Status + Stop + Open
**Notes:** Full control visibility when running.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Between service and databases | Service → Web Admin → Create DB → DB list | ✓ |
| After database cards | Service → Create DB → DB list → Web Admin | |
| You decide | Claude places it where it flows best | |

**User's choice:** Between service and databases
**Notes:** Logical flow: control the server, then its tools, then its data.

---

## Lifecycle Coupling

| Option | Description | Selected |
|--------|-------------|----------|
| No, independent | User manually starts/stops pgweb | ✓ |
| Yes, auto-start with PostgreSQL | pgweb starts when PostgreSQL starts | |
| You decide | Claude picks | |

**User's choice:** No, independent
**Notes:** Simpler, no surprise processes. User only runs pgweb when needed.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, auto-stop pgweb | Stopping PostgreSQL also stops pgweb | ✓ |
| No, leave pgweb running | pgweb stays running even if PostgreSQL stops | |
| You decide | Claude picks | |

**User's choice:** Yes, auto-stop pgweb
**Notes:** pgweb can't function without PostgreSQL — prevents orphan processes.

---

| Option | Description | Selected |
|--------|-------------|----------|
| No, always start fresh | pgweb always starts stopped | ✓ |
| Yes, remember state | Auto-start if was running at last close | |
| You decide | Claude picks | |

**User's choice:** No, always start fresh
**Notes:** Simple, predictable, no surprise processes at launch.

---

## Installation Experience

| Option | Description | Selected |
|--------|-------------|----------|
| Separate install button | Own [Install] button in Web Admin card | ✓ |
| Auto-install with PostgreSQL | Bundled with PostgreSQL installation | |
| You decide | Claude picks | |

**User's choice:** Separate install button
**Notes:** Independent installation. Downloads Go binary from GitHub releases.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Step labels like PostgreSQL | "Downloading..." → "✓ Installed" | ✓ |
| Simple spinner | Just disable button with loading state | |
| You decide | Claude picks | |

**User's choice:** Step labels like PostgreSQL
**Notes:** Matches existing PostgreSQL install pattern in PgDatabasePage.tsx.

---

## Browser Open Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-connect | Pre-configured to connect via socket, trust auth | ✓ |
| Show connection screen | pgweb default connection form | |
| You decide | Claude picks | |

**User's choice:** Auto-connect
**Notes:** User sees databases immediately — no connection form needed.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Localhost only | 127.0.0.1:8081 only | ✓ |
| All interfaces | 0.0.0.0:8081 — accessible from LAN | |
| You decide | Claude picks safest default | |

**User's choice:** Localhost only
**Notes:** Safer for local dev tool. Not accessible from other machines.

---

## Claude's Discretion

- Service architecture (how pgweb fits Service/WebAdminService interfaces)
- Process management (PID tracking, os/exec handle)
- Error handling (download failures, port conflicts)
- pgweb version selection and platform detection

## Deferred Ideas

None — discussion stayed within phase scope.
