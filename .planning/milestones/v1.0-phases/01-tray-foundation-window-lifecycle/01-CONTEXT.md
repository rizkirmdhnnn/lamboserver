# Phase 1: Tray Foundation & Window Lifecycle - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Create a custom CGO system tray package (`internal/tray/`) that renders a monochrome icon in the macOS menu bar. Replace the current close-window confirmation dialog with hide-to-tray behavior. Provide "Show Window" and "Quit" menu items. Services keep running while window is hidden.

</domain>

<decisions>
## Implementation Decisions

### Tray Library
- **D-01:** Custom CGO package at `internal/tray/` using NSStatusBar/NSStatusItem directly. No third-party systray libraries (getlantern/systray, fyne-io/systray, energye/systray all cause ObjC linker conflicts with Wails v2).
- **D-02:** Use a unique ObjC delegate class name (`LamboSystrayDelegate`) to avoid symbol collisions with Wails.
- **D-03:** All Cocoa UI mutations dispatched via `dispatch_async(dispatch_get_main_queue(), ...)` to coexist with Wails' NSApplication event loop.

### Tray Icon
- **D-04:** Monochrome silhouette of the existing LamboServer app logo, rendered as 22x22px macOS template PNG (auto-adapts to dark/light mode). Source icon at `build/appicon.png` and `build/darwin/LamboServer.iconset/`.

### Menu Layout
- **D-05:** Menu shows "LamboServer v{version}" as a disabled text header at the top.
- **D-06:** Menu structure (Phase 1 minimal — later phases extend between header and Quit):
  ```
  LamboServer v1.0.0  (disabled header)
  ─────────────
  Show Window
  ─────────────
  Quit
  ```
- **D-07:** Claude has discretion on exact separator placement and menu item ordering within the Phase 1 structure.

### Window Lifecycle
- **D-08:** Set `HideWindowOnClose: true` in `main.go` Wails options. Remove the existing `OnBeforeClose` / `beforeClose` confirmation dialog entirely.
- **D-09:** Use `runtime.Hide` / `runtime.Show` (application-level) for show/hide. Avoid `runtime.WindowHide` / `runtime.WindowShow` due to documented freeze bug (Wails #2572).
- **D-10:** "Show Window" tray menu item calls `runtime.Show(a.ctx)`.

### Quit Behavior
- **D-11:** "Quit" from tray menu quits immediately with no confirmation dialog.
- **D-12:** "Quit" triggers the existing `shutdown()` flow (stops services in dependency order) via `runtime.Quit(a.ctx)`, not `os.Exit`.
- **D-13:** Services keep running while the app is minimized to tray. Only "Quit" stops them.

### Claude's Discretion
- Menu item ordering and separator placement within the Phase 1 structure
- CGO file organization within `internal/tray/` (tray.go, tray_darwin.go, tray.m, tray_unsupported.go, etc.)
- Whether to use `NSApplicationActivationPolicyAccessory` for Dock icon hiding at runtime (research suggested this but it's optional in Phase 1)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Wails Configuration
- `main.go` — Wails app entry point. Add `HideWindowOnClose: true` to options. Remove `OnBeforeClose: app.beforeClose`.
- `app.go:112-183` — Lifecycle hooks: `startup()`, `beforeClose()` (to be removed), `shutdown()` (reused by Quit).

### Existing Architecture
- `app.go:33-54` — `App` struct definition. Tray controller will be added as a new field.
- `app.go:59` — `NewApp()` composition root. Tray initialization goes in `startup()` after `a.ctx` is assigned.
- `.planning/codebase/ARCHITECTURE.md` — Full architecture analysis.
- `.planning/codebase/CONVENTIONS.md` — Naming conventions (consumer-defined interfaces, D-02 pattern).

### Research
- `.planning/research/STACK.md` — CGO tray library recommendation and Wails compatibility analysis.
- `.planning/research/ARCHITECTURE.md` — Tray controller architecture, GCD dispatch pattern, lifecycle integration.
- `.planning/research/PITFALLS.md` — 14 specific pitfalls with prevention strategies.
- `.planning/research/SUMMARY.md` — Synthesized research findings.

### Build Assets
- `build/appicon.png` — Source for deriving the monochrome tray icon.
- `build/darwin/LamboServer.iconset/` — Existing iconset with multiple resolutions.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/tray/tray.go` — Placeholder package already exists with doc comment. Ready to be fleshed out.
- `app.go:shutdown()` — Existing service shutdown in dependency order. "Quit" from tray will trigger this same flow via `runtime.Quit(a.ctx)`.
- `wailsRuntime` import alias — Already imported as `wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"` in `app.go`.

### Established Patterns
- **Consumer-defined interfaces (D-02):** Each package defines only the interface methods it needs. The tray package should define a narrow `AppHandler` interface with only the methods it calls on `App`.
- **Composition root in `NewApp()`:** All dependency wiring in one place. Tray controller wired here.
- **Lifecycle in `startup()`/`shutdown()`:** Tray init goes in `startup()` after `a.ctx` is set. Tray cleanup goes in `shutdown()`.
- **`a.Debug.Action(name, err)` pattern:** For logging tray operations.

### Integration Points
- `main.go:18-47` — Wails `options.App{}` struct: add `HideWindowOnClose: true`, remove `OnBeforeClose`.
- `app.go:33` — `App` struct: add `Tray *tray.Controller` field.
- `app.go:112` — `startup()`: initialize tray after `a.ctx = ctx`.
- `app.go:171` — `shutdown()`: destroy tray before stopping services.

</code_context>

<specifics>
## Specific Ideas

- Tray icon should be a silhouette of the existing LamboServer logo (not a generic server icon)
- Menu header shows "LamboServer v{version}" as disabled text (like Docker Desktop pattern)
- No confirmation on Quit — user chose Quit intentionally, so just stop services and exit
- The existing Indonesian-language quit dialog (`app.go:162`) will be removed entirely

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-tray-foundation-window-lifecycle*
*Context gathered: 2026-04-17*
