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
	asserterror "github.com/ymz-ncnk/assert/error"
)

func ActiveWriterTestCase() TestCase {
	name := "Should save ActiveWriterEvent"

	var (
		chpts = []chptsup.Tuple{{N: -1, V: -1}}

		tracker = runtime.NewPositionTracker(chpts)
		chptman = chptsup.NewManager(test.NopLogger, chpts, 1)

		event1 = evnt.ActiveWriterEvent{
			NodeID:         "node-1",
			PartitionIndex: 0,
			Timestamp:      0,
		}

		receivedPartition1 int
		receivedEvent1     evnt.Event
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
			),
			Tracker: tracker,
			Chptman: chptman,
			// Resolver: support.MustShardResolver(1),
		},
		Events: []evnt.Event{
			event1,
		},
		Want: Want{
			Errors: []error{nil},
			Check: func(t *testing.T, config TestSetup) {
				asserterror.Equal(t, receivedPartition1, 0)
				asserterror.EqualDeep(t, receivedEvent1, event1)

				asserterror.Equal(
					t,
					strings.HasSuffix(
						chptman.String(), "partitions=1 chptSize=1 [p0=\"chpt:{-1 -1}\"]",
					),
					true,
				)
				asserterror.Equal(
					t,
					strings.HasSuffix(
						tracker.String(), "partitions=1 [p0:NextN=0] tracked=none",
					),
					true,
				)
			},
		},
	}
}
