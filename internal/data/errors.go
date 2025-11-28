// File: internal/data/errors.go
package data

import "errors"

// Define custom error variables for common error scenarios.
var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrEditConflict     = errors.New("edit conflict")
	ErrInvalidID        = errors.New("invalid ID")
	ErrNoRecords        = errors.New("no matching records found")
	ErrDuplicateEmail   = errors.New("duplicate email")
	ErrInsufficientCash = errors.New("insufficient cash provided")
	ErrInvalidData      = errors.New("invalid data provided")
	ErrInvalidRole      = errors.New("invalid role specified")
	ErrAccountNotActive = errors.New("account is not active")
	ErrInvalidToken     = errors.New("invalid or expired token")
)
