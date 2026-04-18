package php

import "errors"

// ErrVersionNotFound is returned when a requested PHP version is not installed.
var ErrVersionNotFound = errors.New("php version not found")

// ErrVersionNotAvailable is returned when a version is not in the download list.
var ErrVersionNotAvailable = errors.New("php version not available for download")

// ErrVersionActive is returned when trying to uninstall the currently active version.
var ErrVersionActive = errors.New("cannot uninstall active php version")

// ErrNoVersionsInstalled is returned when no PHP versions are installed.
var ErrNoVersionsInstalled = errors.New("no php versions installed")
