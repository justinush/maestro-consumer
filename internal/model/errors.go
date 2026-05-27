package model

import "errors"

var (
	// ErrNotFound is returned when a requested app-owned record does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalid is returned when app-level request validation fails (before Maestro inputSchema).
	ErrInvalid = errors.New("invalid input")
)
