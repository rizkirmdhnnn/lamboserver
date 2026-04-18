---
phase: 02-service-controls
verified: 2026-04-18T00:00:00Z
status: human_needed
score: 6/6
overrides_applied: 0
human_verification:
  - test: "Verify service status indicators render correctly in the live tray menu"
    expected: "5 services appear (DNSMasq, Nginx, PHP, MySQL, PostgreSQL) with green filled dot for running, hollow grey dot for stopped, and dash + 'Not Installed' for not-installed services"
    why_human: "Visual rendering of NSAttributedString colored dots in a native macOS menu cannot be verified by static analysis or headless build"
  - test: "Verify Start/Stop/Restart submenu actions trigger service operations"
    expected: "Clicking Start on a stopped service starts it; clicking Stop on a running service stops it; clicking Restart restarts it. Menu closes naturally after click."
    why_human: "Fire-and-forget goroutine behavior and macOS menu dismissal cannot be verified without running the application"
  - test: "Verify context-aware submenu enable/disable state"
    expected: "For a running service: Start is greyed out, Stop and Restart are enabled. For a stopped service: Start is enabled, Stop and Restart are greyed out."
    why_human: "NSMenuItem enabled state requires visual inspection in a running app"
  - test: "Verify status refreshes on menu reopen after an action"
    expected: "After clicking Stop on Nginx, close the menu and reopen it. Nginx dot should now be hollow grey. After clicking Start, reopen and it should be green filled."
    why_human: "Real-time state propagation through the menuWillOpen: -> onRefreshStatuses -> GetAllStatuses -> launchd chain requires live service state"
  - test: "Verify not-installed services have no submenu arrow"
    expected: "A service that is not installed appears with a dash + 'Not Installed' label, is greyed out, and shows no submenu arrow on hover"
    why_human: "NSMenuItem submenu removal (setSubmenu:nil) and disabled visual state require running app inspection"
---

# Phase 2: Service Controls Verification Report

**Phase Goal:** Users can see the status of each service and trigger Start/Stop/Restart directly from the tray menu
**Verified:** 2026-04-18
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Tray menu shows a running/stopped status indicator for each of Nginx, MySQL, PHP, PostgreSQL, and dnsmasq | VERIFIED | `updateServiceItem:withStatus:` in `tray_darwin.m` renders `\u25CF` (green, running) and `\u25CB` (grey, stopped) using `NSAttributedString`. All 5 services listed in `serviceKeys[]` / `serviceNames[]`. |
| 2 | Each service entry has a submenu containing Start, Stop, and Restart actions | VERIFIED | `CreateTray` loop (lines 223-257) builds `NSMenu *sub` per service with Start/Stop/Restart `NSMenuItem` instances connected to `serviceStart:`, `serviceStop:`, `serviceRestart:` selectors. |
| 3 | After triggering a Start/Stop/Restart action, the tray menu status reflects the updated state when next opened | VERIFIED | `menuWillOpen:` calls `onRefreshStatuses()` synchronously every time the menu opens. `onRefreshStatuses` calls `instance.app.GetAllStatuses()` and pushes results through `C.UpdateServiceStatus` to update all 5 menu items before the menu renders. |

**Score:** 3/3 truths verified (code evidence)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/tray/controller.go` | AppController interface with 5 methods | VERIFIED | Contains `Context()`, `GetAllStatuses() map[string]string`, `StartService(name string) error`, `StopService(name string) error`, `RestartService(name string) error` |
| `internal/tray/controller_darwin.go` | CGO bridge callbacks for service actions and status refresh | VERIFIED | `onServiceAction` dispatches start/stop/restart in a fire-and-forget goroutine; `onRefreshStatuses` queries statuses and calls `C.UpdateServiceStatus` for each service in dependency order |
| `internal/tray/tray_darwin.h` | C function declarations for service status bridge | VERIFIED | Declares `void UpdateServiceStatus(int index, const char *status)` and `void RefreshServiceStatuses(void)` |
| `internal/tray/tray_darwin.m` | Complete ObjC service menu UI with NSMenuDelegate, colored dots, submenus | VERIFIED | 352 lines. Contains `NSMenuDelegate` conformance, `menuWillOpen:`, `updateServiceItem:withStatus:`, service action selectors, `UpdateServiceStatus` C function, all color and symbol constants |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `controller_darwin.go` | `controller.go` (AppController interface) | `instance.app.(GetAllStatuses\|StartService\|StopService\|RestartService)` | WIRED | Lines 73-77, 87 in `controller_darwin.go` call all 4 interface methods through the singleton `instance.app` |
| `controller_darwin.go` | `tray_darwin.h` (C functions) | `C.UpdateServiceStatus` call in `onRefreshStatuses` | WIRED | Line 96 `C.UpdateServiceStatus(C.int(i), cStatus)` confirmed |
| `tray_darwin.m (menuWillOpen:)` | `controller_darwin.go (onRefreshStatuses)` | `onRefreshStatuses()` extern call | WIRED | Line 124 in `tray_darwin.m` calls `onRefreshStatuses()` declared as `extern void onRefreshStatuses(void)` (line 8) |
| `tray_darwin.m (serviceStart:/serviceStop:/serviceRestart:)` | `controller_darwin.go (onServiceAction)` | `onServiceAction()` Go export called from ObjC selectors | WIRED | Lines 42-54 in `tray_darwin.m` call `onServiceAction([serviceKeys[idx] UTF8String], "start"/"stop"/"restart")` |
| `app.go` | `internal/tray/controller.go` | `tray.New(a, ...)` passing `*App` as `AppController` | WIRED | Line 155 in `app.go`: `a.Tray = tray.New(a, tray.Icon, "1.0.0")` — `*App` satisfies `AppController` (confirmed: `GetAllStatuses`, `StartService`, `StopService`, `RestartService` all implemented on `*App`) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|-------------------|--------|
| `tray_darwin.m (updateServiceItem:withStatus:)` | `status` NSString | `onRefreshStatuses` → `instance.app.GetAllStatuses()` → `a.Manager.All()` (launchd/pgrep queries) | Yes — `svc.Status()` queries launchd for running/stopped state | FLOWING |
| `controller_darwin.go (onServiceAction)` | `name`, `action` | Hardcoded `serviceKeys[]` array in ObjC + `[sender tag]` index | Yes — service names are registry keys that resolve to real `services.Manager.Get(name)` calls | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| `internal/tray` package compiles | `go build ./internal/tray/` | BUILD_OK | PASS |
| All project tests pass (no regressions) | `go test ./...` | All 14 test packages pass | PASS |
| 4 service-specific CGO export functions present | `grep -c "//export" controller_darwin.go` | 7 total exports (includes Phase 3 additions) | PASS |
| AppController interface has all 4 service methods | `grep "GetAllStatuses\|StartService\|StopService\|RestartService" controller.go` | All 4 found | PASS |
| `menuWillOpen:` calls `onRefreshStatuses()` | grep in tray_darwin.m | Line 124 confirmed | PASS |
| NSMenuDelegate assigned to menu | grep `menu.delegate` | Line 292 confirmed | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SRVC-01 | 02-01, 02-02 | Tray menu shows per-service status indicators (running/stopped) for Nginx, MySQL, PHP, PostgreSQL, dnsmasq | SATISFIED | `updateServiceItem:withStatus:` renders colored dot per service; `onRefreshStatuses` feeds real status from `GetAllStatuses()` |
| SRVC-02 | 02-01, 02-02 | Each service has a submenu with Start, Stop, and Restart actions | SATISFIED | `CreateTray` loop builds 3-item submenu per service; `serviceStart:/serviceStop:/serviceRestart:` selectors call `onServiceAction` which routes to `StartService/StopService/RestartService` |
| SRVC-03 | 02-01, 02-02 | Tray menu refreshes service status after each action completes | SATISFIED | `menuWillOpen:` runs `onRefreshStatuses()` synchronously every time the menu opens, pulling fresh status from launchd/pgrep via `GetAllStatuses()` |

### Anti-Patterns Found

No blockers or warnings found.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | No TODO/FIXME/placeholder comments found | — | — |
| — | — | No empty return stubs found | — | — |
| — | — | No hardcoded empty data in rendering paths | — | — |

### Human Verification Required

The code is fully implemented and wired. All automated checks pass. The following items require human testing in a running application because they involve visual rendering, native macOS menu behavior, and live service state:

#### 1. Status Dot Rendering

**Test:** Run `wails dev`, click the tray icon, observe all 5 service entries.
**Expected:** Running services show a green filled dot (●); stopped services show a grey hollow dot (○); not-installed services show a dash (─) with "ServiceName — Not Installed" label greyed out.
**Why human:** `NSAttributedString` colored dot rendering in a native macOS menu requires visual inspection.

#### 2. Submenu Action Execution

**Test:** Hover over a running service to open its submenu. Click "Stop". Close menu. Reopen the main window or check service state.
**Expected:** Service stops. Menu closes naturally after click (no UI freeze or error).
**Why human:** Fire-and-forget goroutine behavior and side-effects on a live launchd service cannot be tested headlessly.

#### 3. Context-Aware Enable/Disable State

**Test:** Hover over a running service submenu and a stopped service submenu.
**Expected:** Running service: Start greyed out, Stop and Restart active. Stopped service: Start active, Stop and Restart greyed out.
**Why human:** `NSMenuItem setEnabled:` visual state requires native menu rendering.

#### 4. Status Refresh After Action

**Test:** Click "Stop Nginx". Close menu. Reopen menu.
**Expected:** Nginx entry now shows hollow grey dot. Click "Start Nginx", reopen, now shows green filled dot.
**Why human:** Requires live launchd state propagation through the `menuWillOpen:` → `GetAllStatuses()` → launchd query chain.

#### 5. Not-Installed Services Have No Submenu

**Test:** Identify a service not installed on the test machine. Hover over it in the tray menu.
**Expected:** No submenu arrow. Item is greyed out. No hover-popup appears.
**Why human:** `setSubmenu:nil` visual behavior requires native macOS menu inspection.

### Gaps Summary

No gaps. All code artifacts exist, are substantive, and are fully wired. The data flow from live launchd state through Go to the ObjC menu layer is implemented correctly. The three ROADMAP success criteria are satisfied at the code level. Remaining items are human verification only (visual and behavioral testing in a running app).

---

_Verified: 2026-04-18_
_Verifier: Claude (gsd-verifier)_
