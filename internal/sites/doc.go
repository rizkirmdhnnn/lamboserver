// Package site manages local development site configuration for LamboServer.
//
// The Manager type creates and destroys sites by generating Nginx virtual host
// configuration files, linking SSL certificates issued by the local CA, and
// persisting site records to the config store. It uses FileSystem and
// CertManager interfaces for testability.
//
// This package does not manage DNS resolver entries directly; see the dns
// package for resolver configuration.
package sites
