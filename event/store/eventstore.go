package store

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
)

type (
	CheckpointCallback func(chpt int64) error
	EventCallback      func(offset int64, beforeChpt bool, event evnt.Event) error
)

type EventStore interface {
	SaveEvent(ctx context.Context, partition int, event evnt.Event) (offset int64,
		err error)
	SaveCheckpoint(ctx context.Context, partition int, chpt int64) (err error)
	LoadRecent(ctx context.Context, partition int, workflowsCount int,
		chptFn CheckpointCallback, eventFn EventCallback) error
	PartitionCount() int
}
