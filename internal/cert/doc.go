// Package cert manages the local Certificate Authority and per-site SSL
// certificate generation for LamboServer.
//
// The Manager type creates an RSA CA key and self-signed root certificate,
// trusts the CA via the macOS Keychain, and signs per-site TLS certificates.
// It uses the FileSystem and AdminRunner interfaces so that file operations
// and privileged commands can be replaced in tests.
//
// This package does not handle certificate renewal or revocation.
package cert
