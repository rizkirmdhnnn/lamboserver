# Phase 3: Quick Access Links - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-18
**Phase:** 03-quick-access-links
**Areas discussed:** Menu placement, Site display & ordering, Web admin visibility, Empty states

---

## Menu Placement

| Option | Description | Selected |
|--------|-------------|----------|
| Flat sections | Sites listed directly in main menu between services and Show Window, separated by dividers. Web admin tools in their own section. | |
| Sites submenu | Sites go under a single 'Sites' parent with submenu arrow. Web admin tools stay flat. | |
| All in submenus | Both sites and web admin tools under a 'Quick Access' parent submenu. Most compact main menu. | ✓ |

**User's choice:** All in submenus — single "Quick Access" parent submenu for both sites and web admin tools.
**Notes:** User preferred the most compact main menu option despite the extra click depth.

---

## Site Display & Ordering

### Site Label Format

| Option | Description | Selected |
|--------|-------------|----------|
| Domain only | Just show 'my-app.test'. Clean and compact. | |
| Domain + path | Show 'my-app.test — ~/Code/my-app'. More info but takes more space. | |
| You decide | Claude picks the best approach. | ✓ |

**User's choice:** Claude's discretion on label format.

### Sort Order

| Option | Description | Selected |
|--------|-------------|----------|
| Alphabetical | Sites sorted A-Z by domain. Predictable positioning. | |
| Most recently created first | Newest sites at the top. | ✓ |
| You decide | Claude picks. | |

**User's choice:** Most recently created first.

### Overflow Handling

| Option | Description | Selected |
|--------|-------------|----------|
| Truncate + 'More...' item | Show first 15, add disabled '(+N more — see main window)' item. | ✓ |
| Show all, no cap | Ignore 15 cap, show every site. | |
| You decide | Claude picks. | |

**User's choice:** Truncate at 15 with disabled overflow hint.

---

## Web Admin Visibility

| Option | Description | Selected |
|--------|-------------|----------|
| Hidden entirely | If not installed, the item doesn't appear at all. Matches QKAC-02/03. | ✓ |
| Disabled/greyed out | Always visible but greyed out with '(Not Installed)' label. | |

**User's choice:** Hidden entirely when not installed.
**Notes:** Aligns with requirements QKAC-02 and QKAC-03 which specify "shown only when installed".

---

## Empty States

### Empty Quick Access Submenu

| Option | Description | Selected |
|--------|-------------|----------|
| Disabled hint item | Show 'No sites configured' inside Quick Access submenu. Submenu always present. | ✓ |
| Hide entire submenu | Remove Quick Access submenu entirely when nothing to show. | |
| You decide | Claude picks. | |

**User's choice:** Disabled hint item — submenu always present so users know the feature exists.

### Section Separator

| Option | Description | Selected |
|--------|-------------|----------|
| No separator needed | Flat list inside submenu. | |
| Separator between them | Divider line between sites and web admin tools. | |
| You decide | Claude picks. | ✓ |

**User's choice:** Claude's discretion on separator placement.

---

## Claude's Discretion

- Site label format (domain only vs domain + path)
- Separator between sites and web admin tools inside Quick Access submenu
- CGO callback naming and data passing approach

## Deferred Ideas

None — discussion stayed within phase scope.
