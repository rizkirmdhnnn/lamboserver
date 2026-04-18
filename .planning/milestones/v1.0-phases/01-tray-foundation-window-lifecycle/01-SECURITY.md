---
phase: 01-tray-foundation-window-lifecycle
asvs_level: 1
audited_date: "2026-04-17"
auditor: gsd-secure-phase
block_on: high
threats_total: 6
threats_closed: 6
threats_open: 0
result: SECURED
---

# Security Audit — Phase 01: Tray Foundation + Window Lifecycle

## Summary

All 6 threats in the phase threat register are CLOSED. Both mitigate dispositions have confirmed
implementation evidence. Both accept dispositions are documented below. No unregistered threat
flags were raised by the executor.

## Threat Verification

| Threat ID | Category | Disposition | Status | Evidence |
|-----------|----------|-------------|--------|----------|
| T-01-01 | Tampering | accept | CLOSED | Accepted — see Accepted Risks log |
| T-01-02 | Denial of Service | mitigate | CLOSED | internal/tray/controller_darwin.go lines 39–48 (onShowWindow) and 51–60 (onQuit): instance != nil, instance.app != nil, ctx != nil guards present before any wailsRuntime call |
| T-01-03 | Elevation of Privilege | accept | CLOSED | Accepted — see Accepted Risks log |
| T-02-01 | Denial of Service | mitigate | CLOSED | app.go lines 170–172: `if a.Tray != nil { a.Tray.Destroy() }` present as first operation in shutdown() |
| T-02-02 | Information Disclosure | accept | CLOSED | Accepted — see Accepted Risks log |
| T-02-03 | Denial of Service | mitigate | CLOSED | internal/tray/controller_darwin.go line 59: onQuit calls wailsRuntime.Quit(ctx) only — no direct service stop calls. Service stop methods are idempotent (existing behaviour, no new path). Double-shutdown path routes exclusively through Wails OnShutdown -> app.shutdown(). |

## Mitigation Evidence Detail

### T-01-02 — CGO callback nil-guards (controller_darwin.go:39–60)

```go
//export onShowWindow
func onShowWindow() {
    if instance == nil || instance.app == nil {
        return
    }
    ctx := instance.app.Context()
    if ctx == nil {
        return
    }
    wailsRuntime.Show(ctx)
}

//export onQuit
func onQuit() {
    if instance == nil || instance.app == nil {
        return
    }
    ctx := instance.app.Context()
    if ctx == nil {
        return
    }
    wailsRuntime.Quit(ctx)
}
```

Three-layer guard: instance nil-check, app nil-check, context nil-check — all present.

### T-02-01 — Tray nil-guard in shutdown() (app.go:170–172)

```go
if a.Tray != nil {
    a.Tray.Destroy()
}
```

Placed as the first operation in shutdown(), before a.MySQL.Stop() — confirmed at app.go line 170.

### T-02-03 — Single shutdown path (controller_darwin.go:51–60, app.go:168–183)

onQuit calls only wailsRuntime.Quit(ctx). It does not call app.shutdown() directly.
Wails routes that call through its own teardown which fires OnShutdown -> app.shutdown() once.
No second invocation of shutdown() is possible from the tray path.

## Accepted Risks Log

### T-01-01 — Tampering: tray_darwin.m menu actions

**Rationale:** NSStatusItem is OS-managed. Only the owning process (LamboServer) can create or
modify its own status item. macOS process isolation prevents any external process from injecting
menu items into another application's NSStatusItem. No mitigation is needed or possible beyond
the OS trust boundary already in place.

**Owner:** macOS operating system (process isolation)
**Review trigger:** macOS API change removing NSStatusItem process isolation

### T-01-03 — Elevation of Privilege: tray Quit -> shutdown()

**Rationale:** The "Quit" menu item calls wailsRuntime.Quit(ctx), which triggers the existing
app.shutdown() path. shutdown() stops services using the already-authorized privileged helper
(lambo-helper with sudoers entry). No new privilege escalation path is introduced — this is the
same path used by every other quit action in the application.

**Owner:** Existing system/helper.go authorization model
**Review trigger:** Changes to the shutdown() escalation path or helper authorization model

### T-02-02 — Information Disclosure: Services running after window close

**Rationale:** By design (WNDW-04, D-13). LamboServer services are independent launchd daemons.
Continuing to run while the window is hidden is the primary value proposition of system tray
integration. Users expect this behaviour. The tray icon in the menu bar provides a visible
indicator that the app is active.

**Owner:** Product design decision
**Review trigger:** User-facing requirement change requiring services to stop on window close

## Unregistered Threat Flags

None. The executor's SUMMARY.md ## Threat Flags sections (both 01-01 and 01-02) explicitly
state no new network endpoints, auth paths, or trust boundaries were introduced.
