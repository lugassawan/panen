package shared

import "errors"

// Sentinel errors returned by repository implementations.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Sentinel errors for the self-update flow.
var (
	ErrUpdateInProgress    = errors.New("update already in progress")
	ErrUpdateCancelled     = errors.New("update cancelled")
	ErrChecksumMismatch    = errors.New("checksum verification failed")
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)
