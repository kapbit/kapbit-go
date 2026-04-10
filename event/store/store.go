package store

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
)

// Store handles the high-level management of events and checkpoints.
// It coordinates between the lower-level repository and the codec used
// for event serialization.
type Store struct {
	nodeID     string
	codec      Codec
	repository Repository
}

// New creates a new Store instance.
func New(nodeID string, codec Codec, repository Repository) Store {
	return Store{nodeID: nodeID, codec: codec, repository: repository}
}

// SaveEvent serializes an event and persists it to the specified partition.
// Returns the storage offset of the newly saved event.
func (r Store) SaveEvent(ctx context.Context, partition int,
	event evnt.Event,
) (offset int64, err error) {
	record, err := r.codec.Encode(event)
	if err != nil {
		return
	}
	return r.repository.SaveEvent(ctx, partition, record)
}

// SaveCheckpoint persists the consumer's progress (offset) for a given partition.
func (r Store) SaveCheckpoint(ctx context.Context, partition int,
	chpt int64,
) (err error) {
	return r.repository.SaveCheckpoint(ctx, partition, chpt)
}

// LoadRecent reads historical events from a partition, searching backwards
// until a specific number of new workflows (WorkflowCreatedEvent) are found.
// It then replays events forward, calling chptFn for checkpoints and
// eventFn for each decoded event.
func (r Store) LoadRecent(ctx context.Context, partition int, count int,
	chptFn CheckpointCallback, eventFn EventCallback,
) error {
	// Identify the start point by searching for 'count' new workflows.
	startCondition := func(record EventRecord) (bool, error) {
		if record.WorkflowCreatedEvent() {
			count--
		}
		return count <= 0, nil
	}
	// Decode each record before passing it to the caller's callback.
	callback := func(offset int64, beforeChpt bool, record EventRecord) (
		err error,
	) {
		event, err := r.codec.Decode(record)
		if err != nil {
			return
		}
		return eventFn(offset, beforeChpt, event)
	}
	return r.repository.LoadRecent(ctx, partition, startCondition,
		CheckpointCallback(chptFn),
		callback)
}

// PartitionCount returns the total number of partitions available in the system.
func (r Store) PartitionCount() int {
	return r.repository.PartitionCount()
}
