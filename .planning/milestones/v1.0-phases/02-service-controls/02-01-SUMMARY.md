---
phase: 02-service-controls
plan: 01
status: complete
started: 2026-04-18
completed: 2026-04-18
---

# Summary: Expand AppController Interface and CGO Bridge

## What Was Built

Expanded the Go-side CGO bridge for service controls across three files:

1. **AppController interface** (`controller.go`): Added 4 methods — `GetAllStatuses()`, `StartService()`, `StopService()`, `RestartService()` — to the consumer-defined interface, matching the existing `App` struct methods.

2. **CGO export callbacks** (`controller_darwin.go`): Added two new `//export` functions:
   - `onServiceAction`: Routes service name + action string from ObjC, dispatches in a fire-and-forget goroutine (D-06)
   - `onRefreshStatuses`: Queries all statuses via `GetAllStatuses()`, iterates in dependency order (D-02), and calls `C.UpdateServiceStatus` for each service

3. **C header** (`tray_darwin.h`): Declared `UpdateServiceStatus(int index, const char *status)` and `RefreshServiceStatuses(void)` for the ObjC layer to implement/call.

## Key Decisions

- Fire-and-forget goroutine in `onServiceAction` per D-06 — no toast, no spinner
- Fixed service iteration order in `onRefreshStatuses` per D-02: dnsmasq, nginx, php, mysql, postgresql
- C strings freed immediately after use in loop (not deferred) to avoid leaks

## Self-Check: PASSED

- `go build ./internal/tray/` compiles successfully
- 4 `//export` functions confirmed (onShowWindow, onQuit, onServiceAction, onRefreshStatuses)
- All 4 new interface methods present in AppController

## Key Files

key-files.created:
- internal/tray/controller.go (modified — expanded interface)
- internal/tray/controller_darwin.go (modified — 2 new CGO exports)
- internal/tray/tray_darwin.h (modified — 2 new C declarations)
