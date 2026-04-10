package persist

import "fmt"

type CantResolvePartitionError struct {
	cause error
}

func NewCantResolvePartitionError(cause error) error {
	return &CantResolvePartitionError{cause: cause}
}

func (e *CantResolvePartitionError) Error() string {
	return fmt.Sprintf("can't resolve partition: %v", e.cause)
}

func (e *CantResolvePartitionError) Unwrap() error {
	return e.cause
}
