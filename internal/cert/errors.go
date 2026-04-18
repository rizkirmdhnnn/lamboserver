package cert

import "errors"

// ErrCANotFound is returned when the CA certificate or key is missing.
var ErrCANotFound = errors.New("certificate authority not found")

// ErrCertGenerationFailed is returned when certificate generation fails.
var ErrCertGenerationFailed = errors.New("certificate generation failed")

// ErrCorruptedCA is returned when the CA certificate or key file is not a valid PEM file.
var ErrCorruptedCA = errors.New("CA file is corrupted or not a valid PEM file")
