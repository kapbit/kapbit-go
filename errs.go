package kapbit

import (
	"errors"
	"fmt"

	"github.com/kapbit/kapbit-go/codes"
	"github.com/kapbit/kapbit-go/executor"
)

const ErrorPrefix = "kapbit: "

var (
	ErrClosed          = errors.New("closed")
	ErrFenced          = errors.New("fenced")
	ErrEntryGateClosed = errors.New("entry gate closed")
)

// NewKapbitError wraps an existing error with the "kapbit: " prefix.
func NewKapbitError(err error) error {
	return fmt.Errorf(ErrorPrefix+"%w", err)
}

// IsFencedError checks if the error is a PersistenceError with kind
// PersistenceKindNoActiveWriter, indicating that another Kapbit instance is
// the active writer.
func IsFencedError(err error) bool {
	var perr *codes.PersistenceError
	return errors.As(err, &perr) && perr.Kind() == codes.PersistenceKindFenced
}

// IsWillRetryLaterError checks if the error is a IsWillRetryLaterError,
// indicating that the workflow failed but has been scheduled for retry.
func IsWillRetryLaterError(err error) bool {
	var retryErr *executor.WillRetryLaterError
	return errors.As(err, &retryErr)
}

func IsRejectedError(err error) bool {
	var perr *codes.PersistenceError
	return errors.As(err, &perr) && perr.Kind() == codes.PersistenceKindRejection
}

func IsCircuitBreakerOpenError(err error) bool {
	var cberr *codes.CircuitBreakerOpenError
	return errors.As(err, &cberr)
}

type NoActiveWriterEventError struct {
	partition int
}

func NewNoActiveWriterEventError(partition int) error {
	return NoActiveWriterEventError{partition: partition}
}

func (e NoActiveWriterEventError) Partition() int {
	return e.partition
}

func (e NoActiveWriterEventError) Error() string {
	return fmt.Sprintf("no active writer event found for partition %d", e.partition)
}
