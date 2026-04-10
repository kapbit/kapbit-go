package runtime

import (
	"fmt"

	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type RestoreError struct {
	cause error
}

func NewRestoreError(cause error) error {
	return &RestoreError{cause}
}

func (e *RestoreError) Error() string {
	return "restore error: " + e.cause.Error()
}

func (e *RestoreError) Unwrap() error {
	return e.cause
}

func NewPartitionError(partition int, cause error) error {
	return fmt.Errorf("partition %d: %w", partition, cause)
}

func NewUnegisteredWorfklowTypeError(offset int64, tp string) error {
	return fmt.Errorf("workflow type %q is not registered (offset=%d); ensure it "+
		"is present in the workflow definitions map",
		tp, offset)
}

func NewRetryQueueFullError(size int) error {
	return fmt.Errorf("retry queue is full (size=%d); increase its capacity",
		size)
}

func NewInvalidCheckpointError(chpt int64) error {
	return fmt.Errorf("invalid checkpoint %d", chpt)
}

// Restorer errors

func NewUnknownOutcomeTypeError(shard int, offset int64, tp wfl.OutcomeType,
	wid wfl.ID,
) error {
	return fmt.Errorf("unknown outcome type %d (shard=%d, offset=%d)", tp, shard,
		offset)
}

func NewUnknownEventTypeError(shard int, offset int64, event evnt.Event) error {
	return fmt.Errorf("unknown event type %T (shard=%d, offset=%d)", event,
		shard, offset)
}

func NewFencedError(partition int, offset int64, nodeID string) error {
	return fmt.Errorf("fenced: another node %q became active writer (partition=%d, offset=%d)",
		nodeID, partition, offset)
}
