// Filename /internal/data/errors.go
package data

import "errors"

// Define custom error variables for common error scenarios.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrInvalidID      = errors.New("invalid ID")
	ErrNoRecords      = errors.New("no matching records found")
)
