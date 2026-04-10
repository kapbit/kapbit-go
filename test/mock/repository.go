package mock

import (
	"context"

	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/ymz-ncnk/mok"
)

// RepositoryMock is a mock implementation of the eventstore.Repository interface.
type RepositoryMock struct {
	*mok.Mock
}

// NewRepositoryMock returns a new instance of RepositoryMock.
func NewRepositoryMock() RepositoryMock {
	return RepositoryMock{mok.New("Repository")}
}

// --- Function type definitions for mocking ---

type (
	RepositorySaveEventFn      func(ctx context.Context, partition int, event evs.EventRecord) (int64, error)
	RepositorySaveCheckpointFn func(ctx context.Context, partition int, chpt int64) error
	RepositoryLoadRecentFn     func(ctx context.Context, partition int,
		startCondition evs.StartCondition,
		chptFn evs.CheckpointCallback,
		eventFn evs.RecentEventRecordCallback,
	) error
	RepositoryPartitionCountFn func() int
	RepositoryCloseFn          func(ctx context.Context) error
)

// --- Register methods ---

func (m RepositoryMock) RegisterSaveEvent(fn RepositorySaveEventFn) RepositoryMock {
	m.Register("SaveEvent", fn)
	return m
}

func (m RepositoryMock) RegisterNSaveEvent(n int, fn RepositorySaveEventFn) RepositoryMock {
	m.RegisterN("SaveEvent", n, fn)
	return m
}

func (m RepositoryMock) RegisterSaveCheckpoint(fn RepositorySaveCheckpointFn) RepositoryMock {
	m.Register("SaveCheckpoint", fn)
	return m
}

func (m RepositoryMock) RegisterLoadRecent(fn RepositoryLoadRecentFn) RepositoryMock {
	m.Register("LoadRecent", fn)
	return m
}

func (m RepositoryMock) RegisterPartitionCount(fn RepositoryPartitionCountFn) RepositoryMock {
	m.Register("PartitionCount", fn)
	return m
}

func (m RepositoryMock) RegisterNPartitionCount(n int, fn RepositoryPartitionCountFn) RepositoryMock {
	m.RegisterN("PartitionCount", n, fn)
	return m
}

func (m RepositoryMock) RegisterClose(fn RepositoryCloseFn) RepositoryMock {
	m.Register("Close", fn)
	return m
}

// --- Interface method implementations ---

func (m RepositoryMock) SaveEvent(ctx context.Context, partition int,
	event evs.EventRecord,
) (offset int64, err error) {
	results, err := m.Call("SaveEvent", ctx, partition, mok.SafeVal[evs.EventRecord](event))
	if err != nil {
		panic(err)
	}
	offset, _ = results[0].(int64)
	err, _ = results[1].(error)
	return
}

func (m RepositoryMock) SaveCheckpoint(ctx context.Context, partition int, chpt int64) error {
	results, err := m.Call("SaveCheckpoint", ctx, partition, chpt)
	if err != nil {
		panic(err)
	}
	e, _ := results[0].(error)
	return e
}

func (m RepositoryMock) LoadRecent(ctx context.Context, partition int,
	startCondition evs.StartCondition,
	chptFn evs.CheckpointCallback,
	eventFn evs.RecentEventRecordCallback,
) error {
	results, err := m.Call("LoadRecent", ctx, partition,
		mok.SafeVal[any](startCondition),
		mok.SafeVal[any](chptFn),
		mok.SafeVal[any](eventFn))
	if err != nil {
		panic(err)
	}
	e, _ := results[0].(error)
	return e
}

func (m RepositoryMock) WatchAfter(ctx context.Context, partition int,
	action func() error,
	callback evs.EventRecordCallback,
) (int64, error) {
	results, err := m.Call("WatchAfter", ctx, partition,
		mok.SafeVal[any](action),
		mok.SafeVal[any](callback))
	if err != nil {
		panic(err)
	}
	offset, _ := results[0].(int64)
	e, _ := results[1].(error)
	return offset, e
}

func (m RepositoryMock) PartitionCount() int {
	results, err := m.Call("PartitionCount")
	if err != nil {
		panic(err)
	}
	n, _ := results[0].(int)
	return n
}

func (m RepositoryMock) Close(ctx context.Context) error {
	results, err := m.Call("Close", ctx)
	if err != nil {
		panic(err)
	}
	e, _ := results[0].(error)
	return e
}
