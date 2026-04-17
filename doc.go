// Package main is the entry point for LamboServer, a macOS desktop application
// that manages local development infrastructure (PHP, Node.js, Nginx, DNS, SSL)
// through a Wails-based GUI.
//
// The App struct serves as the composition root, wiring all service managers
// with their dependencies via constructor injection. Public methods on App are
// automatically exposed to the React frontend through Wails bindings.
//
// Domain-specific methods are split across focused files: app_php.go,
// app_node.go, app_services.go, app_dashboard.go, app_debug.go, and
// app_cert.go. See the architecture documentation for rationale.
package main
