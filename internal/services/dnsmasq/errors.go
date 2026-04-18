package dnsmasq

import "errors"

// ErrDnsmasqNotInstalled is returned when dnsmasq binary is not found.
var ErrDnsmasqNotInstalled = errors.New("dnsmasq not installed")

// ErrServiceNotRunning is returned when attempting to stop a non-running DNS service.
var ErrServiceNotRunning = errors.New("dns service not running")
