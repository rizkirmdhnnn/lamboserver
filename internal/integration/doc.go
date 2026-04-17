// Package integration contains integration tests that exercise multi-component
// workflows in a controlled temporary environment.
//
// Tests in this package verify end-to-end sequences such as the startup flow,
// PHP and Node version switching, site creation, and service lifecycle
// transitions. They use real implementations of all managers but operate
// against temporary directories to avoid touching the host system.
//
// This package does not test individual manager units — see per-package
// *_test.go files for unit tests.
package integration
