// Package nginx manages the Nginx HTTP server for local development in
// LamboServer.
//
// The Manager type handles Nginx binary detection (Homebrew or system),
// LaunchDaemon lifecycle management (install, start, stop, restart, reload),
// and per-site virtual host configuration file generation from templates.
// It uses FileSystem, LaunchdService, and HelperRunner interfaces for
// testability.
//
// This package does not handle upstream proxy configuration beyond routing
// HTTP requests to per-site PHP-FPM Unix sockets.
package nginx
