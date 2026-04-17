# LamboServer

## What This Is

A macOS desktop application for managing local web development infrastructure. LamboServer provides a unified GUI to install, configure, and control Nginx, PHP, MySQL, PostgreSQL, Node.js, and dnsmasq — with web admin panels (phpMyAdmin for MySQL, pgweb for PostgreSQL) for database management. Built with Go (Wails) and React.

## Core Value

Developers can start, stop, and manage their entire local development stack from a single desktop app without touching the terminal.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- ✓ Nginx lifecycle management (start/stop/restart, virtual host config generation) — pre-v1.0
- ✓ PHP version management (install, switch, PHP-FPM coordination) — pre-v1.0
- ✓ MySQL service management (install, start/stop, socket-based connections) — pre-v1.0
- ✓ PostgreSQL service management (install, start/stop, socket-based connections) — pre-v1.0
- ✓ Node.js version management (install, switch versions) — pre-v1.0
- ✓ dnsmasq for local DNS resolution — pre-v1.0
- ✓ Site management (create/delete sites, Nginx vhosts, SSL certificates) — pre-v1.0
- ✓ Local CA and SSL certificate management (self-signed, Keychain trust) — pre-v1.0
- ✓ phpMyAdmin web admin panel for MySQL — pre-v1.0
- ✓ Configuration persistence (~/.lamboserver/config.json) — pre-v1.0
- ✓ Debug logging system — pre-v1.0
- ✓ macOS desktop notifications (AppleScript) — pre-v1.0
- ✓ System tray placeholder — pre-v1.0
- ✓ pgAdmin fully removed from codebase — v1.0
- ✓ CloudBeaver fully removed from codebase — v1.3
- ✓ pgweb as PostgreSQL web admin (install, start/stop, open in browser) — v1.3
- ✓ Proper .app bundle with Info.plist, bundle identifier (dev.lamboserver.app), and metadata — v1.4
- ✓ App icon in macOS .icns format with all required sizes (geometric Lambo wireframe) — v1.4
- ✓ Wails production build (universal binary, arm64+x86_64) — v1.4
- ✓ Ad-hoc code signing for distribution without Apple Developer account — v1.4
- ✓ DMG installer with drag-to-Applications layout — v1.4

### Active

<!-- Current scope. Building toward these. -->

(None — v1.4 milestone complete, next milestone not yet defined)

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Cloud deployment — this is a local desktop tool
- Windows/Linux support — macOS only for now
- External API integrations — self-contained local tool
- User authentication — single-user desktop app
- pgAdmin restoration — removed in v1.0, replaced by CloudBeaver then pgweb
- CloudBeaver restoration — broken installer, dead download URL, Java dependency; replaced by pgweb (Go binary)
- Adminer integration — changed direction to pgweb (Go binary, no PHP dependency for PostgreSQL admin)

## Milestones

- ✅ v1.0 — Remove pgAdmin, Replace with CloudBeaver (shipped 2026-04-16)
- ✅ v1.2 — Fix MySQL and PostgreSQL Installation Bugs (shipped 2026-04-16)
- ✅ v1.3 — Replace CloudBeaver with pgweb (shipped 2026-04-17)
- ✅ v1.4 — macOS Installer (shipped 2026-04-17)

## Current State

**Shipped:** v1.4 Phase 9 — Code Signing & DMG (2026-04-17)

Ad-hoc signed universal binary .app packaged as DMG installer. `scripts/build-dmg.sh` automates the full pipeline (build -> sign -> DMG). Release notes include Gatekeeper workaround for macOS Sequoia.

**Tech stack:** Go 1.21+, Wails v2.12.0, React, TypeScript
**Services:** Nginx, PHP, MySQL, PostgreSQL, Node.js, dnsmasq, phpMyAdmin, pgweb

## Context

- LamboServer is a brownfield Go/Wails desktop application
- pgweb is the PostgreSQL admin interface (standalone Go binary, replaces CloudBeaver)
- phpMyAdmin is the MySQL admin interface
- Frontend uses React with React Router DOM for page routing
- The pgweb package uses DaemonWebAdminService interface with exec.Command-based process management

## Constraints

- **Platform**: macOS only (Apple Silicon + Intel)
- **Framework**: Wails v2.12.0 (Go backend, React frontend)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| CloudBeaver over pgAdmin | CloudBeaver supports multiple database types | ✓ Good — pgAdmin removed in v1.0; later CloudBeaver itself replaced |
| pgweb over CloudBeaver | Standalone Go binary, no Java dependency, working installer, localhost-only | ✓ Good — v1.3 shipped |
| pgweb over Adminer | Go binary needs no PHP runtime; purpose-built for PostgreSQL | ✓ Good — simpler dependency chain |
| Unified service architecture | ServiceAdapter pattern enables generic service control | ✓ Good |
| DaemonWebAdminService for pgweb | exec.Command + process management; independent from ServiceAdapter | ✓ Good — clean separation |
| D-08 auto-stop coupling | pgweb stops when PostgreSQL stops (both shutdown and StopService paths) | ✓ Good — prevents orphaned processes |
| Ad-hoc signing, no notarization | No Apple Developer account ($99/year) | ✓ Good — Gatekeeper workaround documented |
| No Hardened Runtime | Blocks exec.Command calls to Homebrew services | ✓ Good — essential for service management |
| hdiutil over create-dmg | No external dependencies for DMG creation | ✓ Good — zero-dependency build pipeline |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check -- still the right priority?
3. Audit Out of Scope -- reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-17 after v1.4 milestone*
