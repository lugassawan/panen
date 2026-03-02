package shared

import "errors"

// Sentinel errors returned by repository implementations.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
