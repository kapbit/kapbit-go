package store

import (
	"context"
	"errors"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/test/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func LoadRecentRepositoryErrorTestCase() LoadRecentTestCase {
	name := "Should return an error when Repository.LoadRecent fails"

	repoErr := errors.New("repository error")

	return LoadRecentTestCase{
		Name: name,
		Setup: LoadRecentSetup{
			NodeID: "node-1",
			Codec:  mock.NewCodecMock(),
			Repository: mock.NewRepositoryMock().RegisterLoadRecent(
				func(_ context.Context, _ int, _ evs.StartCondition,
					_ evs.CheckpointCallback, _ evs.RecentEventRecordCallback,
				) error {
					return repoErr
				},
			),
		},
		Params: LoadRecentParams{
			Partition: 0,
			Count:     1,
			ChptFn:    nopChptFn,
			EventFn:   nopEventFn,
		},
		Want: LoadRecentWant{
			Error: repoErr,
		},
	}
}

func LoadRecentDecodeErrorTestCase() LoadRecentTestCase {
	name := "Should return an error when Codec.Decode fails"

	var (
		decodeErr = errors.New("decode error")
		record    = evs.EventRecord{Key: "some-key"}
	)

	return LoadRecentTestCase{
		Name: name,
		Setup: LoadRecentSetup{
			NodeID: "node-1",
			Codec: mock.NewCodecMock().RegisterDecode(
				func(evs.EventRecord) (evnt.Event, error) {
					return nil, decodeErr
				},
			),
			Repository: mock.NewRepositoryMock().RegisterLoadRecent(
				func(_ context.Context, _ int, _ evs.StartCondition,
					_ evs.CheckpointCallback, eventFn evs.RecentEventRecordCallback,
				) error {
					return eventFn(0, false, record)
				},
			),
		},
		Params: LoadRecentParams{
			Partition: 0,
			Count:     1,
			ChptFn:    nopChptFn,
			EventFn:   nopEventFn,
		},
		Want: LoadRecentWant{
			Error: decodeErr,
		},
	}
}

func LoadRecentSuccessTestCase() LoadRecentTestCase {
	name := "Should decode records and call EventCallback with the decoded events"

	var (
		record    = evs.EventRecord{Key: "wfl-1"}
		wantEvent = evnt.WorkflowCreatedEvent{Key: "wfl-1", Type: "test-type"}

		gotOffset int64
		gotEvent  evnt.Event
	)

	return LoadRecentTestCase{
		Name: name,
		Setup: LoadRecentSetup{
			NodeID: "node-1",
			Codec: mock.NewCodecMock().RegisterDecode(
				func(evs.EventRecord) (evnt.Event, error) {
					return wantEvent, nil
				},
			),
			Repository: mock.NewRepositoryMock().RegisterLoadRecent(
				func(_ context.Context, _ int, _ evs.StartCondition,
					_ evs.CheckpointCallback, eventFn evs.RecentEventRecordCallback,
				) error {
					return eventFn(7, false, record)
				},
			),
		},
		Params: LoadRecentParams{
			Partition: 0,
			Count:     1,
			ChptFn:    nopChptFn,
			EventFn: func(offset int64, _ bool, event evnt.Event) error {
				gotOffset = offset
				gotEvent = event
				return nil
			},
		},
		Want: LoadRecentWant{
			Error: nil,
			Check: func(t *testing.T, _ evs.Store) {
				asserterror.Equal(t, gotOffset, int64(7))
				asserterror.EqualDeep(t, gotEvent, wantEvent)
			},
		},
	}
}
