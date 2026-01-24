package store

import "errors"

var (
	// ErrNotFound is returned when the requested entity doesn't exist.
	ErrNotFound = errors.New("not found")

	// ErrDuplicateKey is returned when a unique constraint is violated.
	ErrDuplicateKey = errors.New("duplicate key")
)
