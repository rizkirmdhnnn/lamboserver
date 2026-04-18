// Package config provides thread-safe, persistent application configuration
// storage for LamboServer.
//
// The Store type manages AppConfig (active PHP/Node versions, site mappings,
// Nginx ports, and debug mode) using sync.RWMutex for concurrent access.
// Configuration is persisted to ~/.lamboserver/config.json and loaded on
// startup; reads return from an in-memory cache without file I/O.
//
// This package does not validate version strings or site configurations —
// callers are responsible for supplying valid values.
package config
