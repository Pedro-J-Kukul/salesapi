// File: internal/validator/validator.go
package validator

import (
	"regexp"
	"slices"
)

// ----------------------------------------------------------------------
//
//	Common Regular Expressions
//
// ----------------------------------------------------------------------

// EmailRegex is a regular expression for validating email addresses.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Password Comlpexity Regex
var (
	PasswordNumberRX  = regexp.MustCompile("[0-9]")
	PasswordUpperRX   = regexp.MustCompile("[A-Z]")
	PasswordLowerRX   = regexp.MustCompile("[a-z]")
	PasswordSpecialRX = regexp.MustCompile("[!@#~$%^&*()+|_]")
	PasswordMinLength = 8
	PasswordMaxLength = 72
)

// ----------------------------------------------------------------------
//
//	 Validation Utilities
//
// ----------------------------------------------------------------------

// Validator is a struct that holds validation errors.
type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if there are no validation errors.
func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message for a specific key.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message if the condition is false.
func (v *Validator) Check(condition bool, key, message string) {
	if !condition {
		v.AddError(key, message)
	}
}

// In checks if a value is in a list of permitted values.
func (v *Validator) Permitted(value string, permittedValues ...string) bool {
	return slices.Contains(permittedValues, value)
}

// Matches checks if a string matches a given regular expression.
func (v *Validator) Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
