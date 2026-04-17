# Technology Stack

**Analysis Date:** 2026-04-16

## Languages

**Primary:**
- Go 1.25.0 - Backend application, service managers, system integration
- TypeScript 4.6.4 - Frontend type definitions
- JavaScript/JSX (React) - Frontend UI components

**Secondary:**
- Shell/Bash - System scripts, helper integration
- AppleScript - macOS notification system integration

## Runtime

**Environment:**
- macOS (Apple Silicon or Intel required)
- Wails Runtime v2.12.0 - Desktop application framework (Go/React bridge)

**Package Manager:**
- npm (Node.js) - Frontend dependencies
- Go module system - Backend dependencies (`go.mod` at `go.mod`)

## Frameworks

**Core:**
- Wails v2.12.0 - Desktop application framework providing IPC bridge between Go backend and React frontend (`github.com/wailsapp/wails/v2`)
- Echo v4.13.3 - HTTP server library (dependency of Wails, used for internal routing)

**Frontend:**
- React 18.2.0 - UI framework
- React Router DOM 7.14.1 - Client-side routing
- Vite 3.0.7 - Build tool and dev server
- Lucide React 1.8.0 - Icon library

**Testing:**
- Testify v1.11.1 - Go testing assertions and mocking utilities

**Build/Dev:**
- Vite TypeScript/React Plugin - Frontend build configuration

## Key Dependencies

**Critical:**
- github.com/go-sql-driver/mysql v1.9.1 - MySQL database driver for database/sql operations
- github.com/jackc/pgx/v5 v5.9.1 - PostgreSQL database driver (pgx) for database/sql
- github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c - Opens URLs in default browser

**Infrastructure:**
- github.com/wailsapp/go-webview2 v1.0.22 - WebView2 renderer for Windows (indirect dependency)
- github.com/wailsapp/mimetype v1.4.1 - MIME type detection
- github.com/labstack/gommon v0.4.2 - Common utilities for Echo server
- gopkg.in/yaml.v3 v3.0.1 - YAML configuration parsing
- github.com/google/uuid v1.6.0 - UUID generation for certificate and service identifiers

**System Integration:**
- git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 - Windows toast notifications (indirect)
- github.com/godbus/dbus/v5 v5.1.0 - D-Bus system integration for Linux (indirect)
- github.com/go-ole/go-ole v1.3.0 - Windows COM automation (indirect)

**Utilities:**
- github.com/samber/lo v1.49.1 - Go utility functions (map, filter, reduce)
- github.com/bep/debounce v1.2.1 - Debounce function for UI events

## Configuration

**Environment:**
- Configuration stored in JSON at `~/.lamboserver/config.json`
- Config structure defined in `internal/config/store.go` with fields:
  - `ActivePhpVersion`: Selected PHP version
  - `ActiveNodeVersion`: Selected Node.js version
  - `Sites`: Array of site configurations
  - `NginxPort`: HTTP port (default 80)
  - `NginxSSLPort`: HTTPS port (default 443)
  - `DebugMode`: Verbose logging toggle
  - `MySQLEnabled`: MySQL service toggle
  - `PostgreSQLEnabled`: PostgreSQL service toggle
- Application data stored in `~/.lamboserver/` directory hierarchy

**Build:**
- `wails.json` - Wails application configuration (`wails.json`)
- `frontend/vite.config.ts` - Vite frontend bundler configuration
- `frontend/tsconfig.json` - TypeScript compiler options
- `frontend/package.json` - Frontend dependencies and build scripts

## Platform Requirements

**Development:**
- macOS (Apple Silicon or Intel)
- Go 1.25.0 or higher
- Node.js (any recent version) for frontend development
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

**Production:**
- Deployment target: macOS desktop application
- No external dependencies required - all runtimes (PHP, Node.js, MySQL, PostgreSQL, Nginx, dnsmasq) are embedded or managed locally
- Application runs with privilege escalation where needed for privileged operations (port 80, DNS)

---

*Stack analysis: 2026-04-16*
