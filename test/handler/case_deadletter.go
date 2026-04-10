package handler

import (
	"context"
	"strings"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func DeadLetterTestCase() TestCase {
	name := "Should add checkpoint and untrack on terminal dead letter event"

	var (
		chpts = []chptsup.Tuple{{N: -1, V: -1}}

		tracker = runtime.NewPositionTracker(chpts)
		chptman = chptsup.NewManager(test.NopLogger, chpts, 1)

		event1 = evnt.WorkflowCreatedEvent{
			NodeID:    "node-1",
			Key:       wfl.ID("wfl-1"),
			Type:      wfl.Type("type-a"),
			Input:     "input-1",
			Timestamp: 0,
		}
		event2 = &evnt.DeadLetterEvent{
			NodeID:    "node-1",
			Key:       wfl.ID("wfl-1"),
			Type:      wfl.Type("type-a"),
			Timestamp: 1,
		}

		receivedPartition1 int
		receivedPartition2 int
		receivedPartition3 int
		receivedPartition4 int

		receivedEvent1 evnt.Event
		receivedEvent2 evnt.Event
		receivedChpt   int64
	)

	return TestCase{
		Name: name,
		Setup: TestSetup{
			Store: mock.NewEventStoreMock().RegisterPartitionCount(
				func() int {
					return 1
				},
			).RegisterSaveEvent(
				func(ctx context.Context, partition int, event evnt.Event) (offset int64, err error) {
					receivedPartition1 = partition
					receivedEvent1 = event
					return 10, nil
				},
			).RegisterSaveEvent(
				func(ctx context.Context, partition int, event evnt.Event) (offset int64, err error) {
					receivedPartition2 = partition
					receivedEvent2 = event
					return 11, nil
				},
			).RegisterSaveCheckpoint(
				func(ctx context.Context, partition int, chpt int64) error {
					receivedPartition4 = partition
					receivedChpt = chpt
					return nil
				},
			),
			Tracker: tracker,
			Chptman: chptman,
		},
		Events: []evnt.Event{
			event1,
			event2,
		},
		Want: Want{
			Errors: []error{nil, nil},
			Check: func(t *testing.T, config TestSetup) {
				deadLetter := event2
				deadLetter.WorkflowOffset = 10
				asserterror.Equal(t, receivedPartition1, 0)
				asserterror.EqualDeep(t, receivedEvent1, event1)
				asserterror.Equal(t, receivedPartition2, 0)
				asserterror.EqualDeep(t, receivedEvent2, deadLetter)
				asserterror.Equal(t, receivedPartition3, 0)
				asserterror.Equal(t, receivedPartition4, 0)
				asserterror.Equal(t, receivedChpt, 10)

				asserterror.Equal(
					t,
					strings.HasSuffix(
						chptman.String(), "partitions=1 chptSize=1 [p0=\"chpt:{0 10}\"]",
					),
					true,
				)
				asserterror.Equal(
					t,
					strings.HasSuffix(
						tracker.String(), "partitions=1 [p0:NextN=1] tracked=none",
					),
					true,
				)
			},
		},
	}
}
