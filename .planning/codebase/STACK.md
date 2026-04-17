# Technology Stack

**Analysis Date:** 2026-04-17

## Languages

**Primary:**
- Go 1.25.0 - Backend logic, service management, system integration (`go.mod`)
- TypeScript ^4.6.4 - Frontend UI (`frontend/package.json`)

**Secondary:**
- CSS - Styling (`frontend/src/style.css`)
- Bash - Build and release scripts (`scripts/build-dmg.sh`)

## Runtime

**Environment:**
- macOS only (darwin) - Application uses macOS-specific APIs (launchd, Keychain, /etc/resolver)
- Wails v2 runtime - Go + embedded WebView (WKWebView on macOS)
- Architecture support: darwin/universal (arm64 + amd64) via `wails build -platform darwin/universal`

**Package Manager:**
- Go modules (`go.mod`, `go.sum` present)
- npm for frontend (`frontend/package.json`)
- Lockfile: `go.sum` present; `frontend/package-lock.json` should exist after install

## Frameworks

**Core:**
- Wails v2 (v2.12.0) - Desktop app framework bridging Go backend to WebView frontend (`go.mod`)
- React 18 (^18.2.0) - Frontend UI framework (`frontend/package.json`)

**Testing:**
- testify v1.11.1 - Go test assertions and mocking (`go.mod`)
- No frontend testing framework detected

**Build/Dev:**
- Vite ^3.0.7 - Frontend bundler and dev server (`frontend/package.json`)
- @vitejs/plugin-react ^2.0.1 - React HMR support (`frontend/vite.config.ts`)
- Wails CLI - Build tool (`wails build`, `wails dev`)

## Key Dependencies

**Critical (Go):**
- `github.com/wailsapp/wails/v2` v2.12.0 - Core desktop framework, IPC between Go and React
- `github.com/go-sql-driver/mysql` v1.9.1 - MySQL database driver for database CRUD operations (`internal/services/mysql/manager.go`)
- `github.com/jackc/pgx/v5` v5.9.1 - PostgreSQL database driver via `pgx/stdlib` (`internal/services/postgres/manager.go`)
- `github.com/pkg/browser` v0.0.0-20240102092130 - Opens URLs in default browser for web admin tools (`app.go`)

**Critical (Frontend):**
- `react-router-dom` ^7.14.1 - Client-side routing (declared but App.tsx uses manual state-based navigation)
- `lucide-react` ^1.8.0 - Icon library for sidebar navigation (`frontend/src/App.tsx`)

**Infrastructure (Go indirect):**
- `github.com/labstack/echo/v4` v4.13.3 - HTTP server used internally by Wails
- `github.com/gorilla/websocket` v1.5.3 - WebSocket for Wails dev mode communication
- `github.com/google/uuid` v1.6.0 - UUID generation
- `git.sr.ht/~jackmordaunt/go-toast/v2` v2.0.3 - macOS toast notifications
- `golang.org/x/crypto` v0.33.0 - Cryptographic primitives (SSL cert generation)

## Configuration

**Application Config:**
- `~/.lamboserver/config.json` - JSON-based config persisted by `internal/config/store.go`
- Thread-safe in-memory cache with `sync.RWMutex`, atomic file writes
- Config schema defined in `internal/config/store.go` as `AppConfig` struct

**Wails Config:**
- `wails.json` - Wails project configuration (app name, build commands, author info)

**TypeScript Config:**
- `frontend/tsconfig.json` - Strict mode enabled, ESNext target, react-jsx transform
- `frontend/vite.config.ts` - Minimal Vite config with React plugin

**Build Config:**
- `build/darwin/entitlements.plist` - macOS entitlements (JIT, unsigned memory, network, Apple Events)
- `build/darwin/Info.plist` / `Info.dev.plist` - macOS app bundle metadata

**No `.env` files detected** - Application uses no environment variables. All config is in `~/.lamboserver/config.json`.

## Build Pipeline

**Development:**
```bash
wails dev                    # Starts Go backend + Vite dev server with HMR
```

**Production Build:**
```bash
wails build -platform darwin/universal -clean  # Builds universal macOS .app
```

**DMG Packaging:**
```bash
./scripts/build-dmg.sh      # Builds app, ad-hoc signs, creates DMG
```

**Frontend Only:**
```bash
cd frontend && npm run dev   # Vite dev server
cd frontend && npm run build # tsc + vite build (outputs to frontend/dist/)
```

**CI/CD:**
- GitHub Actions workflow: `.github/workflows/release.yml`
- Trigger: Push tag matching `v*`
- Runner: `macos-latest`
- Steps: Setup Go 1.25 + Node 20 -> Install Wails -> Build universal .app -> Ad-hoc codesign -> Create DMG -> GitHub Release via `softprops/action-gh-release@v2`
- No Apple Developer signing (ad-hoc only, per project constraint)

## Embedded Assets

**Frontend:**
- `frontend/dist/` embedded via Go `//go:embed all:frontend/dist` directive in `main.go`
- Served by Wails asset server at runtime

**System Binaries:**
- `internal/system/embedded/darwin-arm64/nginx` - Pre-compiled Nginx binary
- `internal/system/embedded/darwin-arm64/dnsmasq` - Pre-compiled dnsmasq binary

## Platform Requirements

**Development:**
- macOS (required for Wails WebView, launchd, and system integration)
- Go 1.25+
- Node.js 20+ (for frontend build)
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

**Production:**
- macOS only (uses launchd, Keychain, /etc/resolver, WKWebView)
- No Apple Developer Account required (ad-hoc signing)
- Distributed as `.dmg` file

---

*Stack analysis: 2026-04-17*
