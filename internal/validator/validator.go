package validator

import (
	"regexp"
	"slices"
)

// Validator struct to hold validation errors.
type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// IsEmpty checks if there are no validation errors.
func (v *Validator) IsEmpty() bool {
	return len(v.Errors) == 0
}

// AddErrors adds a new error message for a given key if it doesn't already exist.
func (v *Validator) AddErrors(key string, message string) {
	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message for a key if the condition is false.
func (v *Validator) Check(ok bool, key string, message string) {
	if !ok {
		v.AddErrors(key, message)
	}
}

// Matches checks if the value matches the given regular expression.
func (v *Validator) Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// IsOkay checks if the value is within the permitted values.
func (v *Validator) IsOkay(value string, permittedValues ...string) bool {
	return slices.Contains(permittedValues, value)
}
