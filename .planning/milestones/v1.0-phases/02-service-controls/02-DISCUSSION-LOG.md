# Phase 2: Service Controls - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 02-service-controls
**Areas discussed:** Menu layout, Status indicators, Action feedback, Not-installed services

---

## Menu Layout

| Option | Description | Selected |
|--------|-------------|----------|
| Submenus per service | Each service is a top-level item that expands to a submenu with Start/Stop/Restart. Compact, similar to Docker Desktop. | ✓ |
| Flat actions per service | Each service gets its own section with Start/Stop/Restart as separate items in the main menu. Longer menu. | |
| You decide | Claude picks the best layout | |

**User's choice:** Submenus per service
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Dependency order | dnsmasq → Nginx → PHP → MySQL → PostgreSQL. Matches startup order in app.go. | ✓ |
| Alphabetical | dnsmasq, MySQL, Nginx, PHP, PostgreSQL | |
| You decide | Claude picks | |

**User's choice:** Dependency order
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Actions only | Start, Stop, Restart — clean and minimal | ✓ |
| Actions + status text | Status as disabled text item at top, then actions | |
| Actions + version | Version as disabled text, then actions | |

**User's choice:** Actions only (Start, Stop, Restart)
**Notes:** None

---

## Status Indicators

| Option | Description | Selected |
|--------|-------------|----------|
| Colored dot symbols | ● green for running, ○ hollow for stopped. NSAttributedString colorization. | ✓ |
| Text suffix | Append 'Running' or 'Stopped' as plain text | |
| Checkmark/dash | ✓ for running, — for stopped | |

**User's choice:** Colored dot symbols
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, context-aware | Running: Start disabled. Stopped: Stop/Restart disabled. | ✓ |
| All always enabled | All three actions always clickable | |
| You decide | Claude picks | |

**User's choice:** Yes, context-aware actions
**Notes:** None

---

## Action Feedback

| Option | Description | Selected |
|--------|-------------|----------|
| Fire-and-forget, refresh on reopen | Action runs in background goroutine. Menu closes naturally. Status updates on next open. | ✓ |
| Toast notification on completion | Same fire-and-forget plus macOS toast notification using go-toast library | |
| Keep menu open with spinner | Menu stays open during action, shows loading state, updates in-place | |

**User's choice:** Fire-and-forget, refresh on reopen
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| On menu open | Query all statuses each time tray icon clicked. Uses NSMenuDelegate menuWillOpen: callback. | ✓ |
| On menu open + after action | Same plus refresh after action completes (pre-cache) | |
| You decide | Claude picks | |

**User's choice:** On menu open via menuWillOpen:
**Notes:** None

---

## Not-Installed Services

| Option | Description | Selected |
|--------|-------------|----------|
| Show disabled with label | Show with dash (─) and "Not Installed" label, greyed out, no submenu | ✓ |
| Hide completely | Only show installed services | |
| Show with Install action | Show with Install action in submenu | |

**User's choice:** Show disabled with "Not Installed" label
**Notes:** None

---

## Claude's Discretion

- NSAttributedString formatting details (font size, color values)
- Go ↔ ObjC callback bridge structure for service actions and status queries
- Single RefreshMenu function vs individual item updates
- CGO callback naming conventions
- Thread safety for concurrent service action goroutines

## Deferred Ideas

None — discussion stayed within phase scope
