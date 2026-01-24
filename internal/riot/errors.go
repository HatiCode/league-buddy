package riot

import "errors"

var (
	// ErrNotFound is returned when the requested resource doesn't exist.
	ErrNotFound = errors.New("not found")

	// ErrRateLimited is returned when the API rate limit is exceeded.
	ErrRateLimited = errors.New("rate limited")

	// ErrUnauthorized is returned when the API key is invalid or expired.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidRegion is returned when an unsupported region is specified.
	ErrInvalidRegion = errors.New("invalid region")
)
