// Package system provides platform abstractions for macOS system operations
// used throughout LamboServer.
//
// It supplies: Paths, the single source of truth for all file and directory
// locations under ~/.lamboserver/; LaunchdManager, which manages macOS service
// lifecycle via launchctl (load/unload/start/stop); Integration, which injects
// the LamboServer bin directory into shell rc files (.bashrc, .zshrc);
// BinaryLocator, which discovers binaries by checking a known local path
// before falling back to system PATH; and the real implementations of the
// FileSystem (RealFS), CommandRunner (RealCmdRunner), and AdminRunner
// (RealAdminRunner) interfaces used by service managers.
//
// Darwin-only operations such as osascript privilege elevation and Keychain
// access are implemented in darwin.go. This package does not contain
// service-specific logic.
package system
