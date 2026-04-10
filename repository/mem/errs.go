package mem

import (
	"fmt"
)

// var ErrTooLargePayload = errors.New("payload too large")

func NewTooLargePayloadError(size int) error {
	return fmt.Errorf("payload too large: %d", size)
}

func NewInvalidPartitionError(index, count int) error {
	return fmt.Errorf("invalid partition index %d, total partitions count is %d (valid range 0 to %d)",
		index, count, count-1)
}
