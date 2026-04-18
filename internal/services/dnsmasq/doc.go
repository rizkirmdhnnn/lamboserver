// Package dns manages dnsmasq for local DNS resolution in LamboServer.
//
// The Manager type handles dnsmasq installation detection, writes resolver
// configuration files under /etc/resolver/ for the local TLD, and controls
// the dnsmasq LaunchDaemon lifecycle (install, start, stop, restart). It uses
// FileSystem, LaunchdService, and AdminRunner interfaces for testability.
//
// This package does not manage external DNS records or upstream resolvers.
package dnsmasq
