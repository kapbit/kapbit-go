package executor

import (
	"errors"
	"fmt"
)

var (
	ErrIdempotencyViolation = errors.New("idempotency violation")
	ErrNoSlotsAvailable     = errors.New("no slots available")
)

type WillRetryLaterError struct {
	cause error
}

func NewWillRetryLaterError(err error) error {
	return &WillRetryLaterError{cause: err}
}

func (e *WillRetryLaterError) Error() string {
	return fmt.Sprintf("workflow failed, scheduled for retry: %v", e.cause)
}

func (e *WillRetryLaterError) Unwrap() error {
	return e.cause
}
