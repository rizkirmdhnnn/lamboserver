package nginx

import "errors"

// ErrNginxNotInstalled is returned when nginx binary is not found.
var ErrNginxNotInstalled = errors.New("nginx not installed")

// ErrServiceNotRunning is returned when attempting to stop/reload a non-running service.
var ErrServiceNotRunning = errors.New("nginx service not running")
