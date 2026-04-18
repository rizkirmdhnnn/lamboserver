// Package php manages PHP version installation, detection, activation, and
// PHP-FPM daemon lifecycle for LamboServer.
//
// The Manager type detects PHP installations from four sources (Homebrew,
// System, Herd, and manually installed), downloads new PHP releases from
// herdphp.com, and activates a version by updating a symlink on the PATH.
// The FpmManager type controls per-version PHP-FPM processes through macOS
// launchd (start, stop, restart, status). Both types use FileSystem,
// CommandRunner, and LaunchdService interfaces for testability.
//
// This package does not configure php.ini settings or manage PHP extensions.
package php
