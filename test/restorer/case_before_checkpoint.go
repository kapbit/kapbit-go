package restorer

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

func BeforeCheckpointTestCase() TestCase {
	name := "Should handle events before checkpoint during restoration"

	var (
		wid   = wfl.ID("wfl-before")
		store = mock.NewEventStoreMock()
	)

	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) (err error) {
			// Event marked as beforeChpt=true
			return eventFn(0, true, evnt.WorkflowCreatedEvent{
				Key:  wid,
				Type: "test-workflow",
			})
		},
	).RegisterNPartitionCount(2, func() int { return 1 })

	return TestCase{
		Name: name,
		Config: TestConfig{
			NodeID:  "node-1",
			Store:   store,
			Factory: mock.NewFactoryMock(),
			Opts: []rtm.SetOption{
				rtm.WithLogger(test.NopLogger),
			},
		},
		Want: Want{
			CounterValue:         0,
			IdempotencyWindowStr: "support.FIFOSet cap=1000 len=1 [wfl-before]",
			ChptmanStr:           "checkpoint.Manager partitions=1 chptSize=100 [p0=\"chpt:{-1 -1}\"]",
			TrackerStr:           "runtime.PositionTracker partitions=1 [p0:NextN=0] tracked=none",
		},
	}
}
