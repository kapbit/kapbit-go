package mock

import (
	"context"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/ymz-ncnk/mok"
)

// EventStoreMock is a mock implementation of the EventStore interface.
type EventStoreMock struct {
	*mok.Mock
}

// NewEventStoreMock returns a new instance of EventStoreMock.
func NewEventStoreMock() EventStoreMock {
	return EventStoreMock{mok.New("EventStore")}
}

// --- Function type definitions for mocking ---

type (
	EventStoreSaveEventFn      func(ctx context.Context, partition int, event evnt.Event) (int64, error)
	EventStoreSaveCheckpointFn func(ctx context.Context, partition int, chpt int64) error
	EventStoreLoadRecentFn     func(ctx context.Context, partition int, workflowsCount int,
		chptFn evs.CheckpointCallback, eventFn evs.EventCallback) error
	EventStoreReadIntervalFn func(ctx context.Context, partition int, currentOffset int64,
		wait time.Duration,
		callback evs.LoadEventCallback,
	) error
)
type EventStorePartitionCountFn func() int

// --- Register methods ---

func (m EventStoreMock) RegisterSaveEvent(fn EventStoreSaveEventFn) EventStoreMock {
	m.Register("SaveEvent", fn)
	return m
}

func (m EventStoreMock) RegisterNSaveEvent(n int, fn EventStoreSaveEventFn) EventStoreMock {
	m.RegisterN("SaveEvent", n, fn)
	return m
}

func (m EventStoreMock) RegisterSaveCheckpoint(fn EventStoreSaveCheckpointFn) EventStoreMock {
	m.Register("SaveCheckpoint", fn)
	return m
}

func (m EventStoreMock) RegisterNSaveCheckpoint(n int, fn EventStoreSaveCheckpointFn) EventStoreMock {
	m.RegisterN("SaveCheckpoint", n, fn)
	return m
}

func (m EventStoreMock) RegisterLoadRecent(fn EventStoreLoadRecentFn) EventStoreMock {
	m.Register("LoadRecent", fn)
	return m
}

func (m EventStoreMock) RegisterPartitionCount(fn EventStorePartitionCountFn) EventStoreMock {
	m.Register("PartitionCount", fn)
	return m
}

func (m EventStoreMock) RegisterNPartitionCount(n int, fn EventStorePartitionCountFn) EventStoreMock {
	m.RegisterN("PartitionCount", n, fn)
	return m
}

func (m EventStoreMock) RegisterReadInterval(fn EventStoreReadIntervalFn) EventStoreMock {
	m.Register("ReadInterval", fn)
	return m
}

// --- Interface method implementations ---

func (m EventStoreMock) SaveEvent(ctx context.Context, partition int, event evnt.Event) (offset int64, err error) {
	results, err := m.Call("SaveEvent", ctx, partition, mok.SafeVal[evnt.Event](event))
	if err != nil {
		panic(err)
	}
	offset, _ = results[0].(int64)
	err, _ = results[1].(error)
	return
}

func (m EventStoreMock) SaveCheckpoint(ctx context.Context, partition int, chpt int64) (err error) {
	results, err := m.Call("SaveCheckpoint", ctx, partition, chpt)
	if err != nil {
		panic(err)
	}
	err, _ = results[0].(error)
	return
}

func (m EventStoreMock) LoadRecent(ctx context.Context, partition int, workflowsCount int,
	chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
) (err error) {
	results, err := m.Call("LoadRecent", ctx, partition, workflowsCount,
		mok.SafeVal[any](chptFn), mok.SafeVal[any](eventFn))
	if err != nil {
		panic(err)
	}
	err, _ = results[0].(error)
	return
}

func (m EventStoreMock) PartitionCount() (n int) {
	results, err := m.Call("PartitionCount")
	if err != nil {
		panic(err)
	}
	n, _ = results[0].(int)
	return
}

func (m EventStoreMock) ReadInterval(ctx context.Context, partition int,
	currentOffset int64,
	wait time.Duration,
	callback evs.LoadEventCallback,
) (err error) {
	results, err := m.Call("ReadInterval", ctx, partition, currentOffset, wait,
		mok.SafeVal[any](callback))
	if err != nil {
		panic(err)
	}
	err, _ = results[0].(error)
	return
}
