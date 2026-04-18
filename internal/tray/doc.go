// Package tray manages the macOS system tray icon and menu for LamboServer.
//
// The Controller type creates an NSStatusItem in the macOS menu bar with a
// monochrome template icon and a dropdown menu containing Show Window and Quit
// actions. It uses CGO to bridge Go to Objective-C AppKit APIs (NSStatusBar,
// NSMenu) with all UI mutations dispatched to the main thread via GCD.
//
// On non-darwin platforms, all methods are no-ops.
package tray
