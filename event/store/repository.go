package store

import (
	"context"
)

type (
	StartCondition            func(record EventRecord) (bool, error)
	RecentEventRecordCallback func(offset int64, beforeChpt bool, record EventRecord) error
	LoadEventCallback         func(offset int64, event EventRecord) (stop bool, err error)
	EventRecordCallback       func(ctx context.Context, event EventRecord) (bool, error)
)

type Repository interface {
	SaveEvent(ctx context.Context, partition int, event EventRecord) (
		offset int64, err error)
	SaveCheckpoint(ctx context.Context, partition int, chpt int64) (err error)

	// ReadInterval(ctx context.Context, partition int, currentOffset int64,
	// 	wait time.Duration,
	// 	callback LoadEventCallback,
	// ) (err error)

	// ClaimLeadership(ctx context.Context, event EventRecord) (err error)

	// LoadFromOffset scans the log forward starting from a specific offset.
	// usage:
	// 1. Runtime loads the last "Global Shard Checkpoint" (e.g., 5000).
	// 2. Runtime calls LoadFromOffset(..., 5000, ...).
	// 3. Storage replays everything from 5000 to the end.

	// LoadRecent loads workflows from every shard.
	//
	// If worklowCount > than the number of workflows after the shard's
	// checkpoint, calculates the offset from which to start.
	// For all worklows before the checkpoint EventRecord contains only
	// WorkflowID all other fields stay unset.
	// If workflowCount <= than the number of workflow after the shard's
	// checkpoint, starts from the checkpoint.
	LoadRecent(ctx context.Context, partition int, startCondition StartCondition,
		chptFn CheckpointCallback,
		eventFn RecentEventRecordCallback,
	) error

	WatchAfter(ctx context.Context, partition int, action func() error,
		callback EventRecordCallback) (int64, error)

	PartitionCount() int
	Close(ctx context.Context) error
}
