package codes

import (
	"fmt"
	"reflect"
)

// EncodeError is returned when a value cannot be serialized into its storage
// format.
type EncodeError struct {
	tp     reflect.Type
	reason string
	err    error
}

// NewEncodeError creates a new EncodeError for the specified type and reason.
func NewEncodeError(tp reflect.Type, reason string, cause error) error {
	return &EncodeError{tp: tp, reason: reason, err: cause}
}

// Error returns a formatted error message.
func (e *EncodeError) Error() string {
	return fmt.Sprintf("encode error for %v: %s (%v)", e.tp, e.reason, e.err)
}

// Unwrap returns the underlying error.
func (e *EncodeError) Unwrap() error { return e.err }

// DecodeError is returned when a value cannot be deserialized from its storage
// format.
type DecodeError struct {
	tp     string
	reason string
	cause  error
}

// NewDecodeError creates a new DecodeError for the specified target type
// and reason.
func NewDecodeError(tp string, reason string, cause error) error {
	return &DecodeError{tp: tp, reason: reason, cause: cause}
}

// Error returns a formatted error message.
func (e *DecodeError) Error() string {
	return fmt.Sprintf("decode error [%s]: %s (%v)", e.tp, e.reason, e.cause)
}

// Unwrap returns the underlying error.
func (e *DecodeError) Unwrap() error { return e.cause }
