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

func SkipEventTestCase() TestCase {
	name := "Should skip events for unknown workflows during restoration"

	var (
		wid   = wfl.ID("unknown-wfl")
		store = mock.NewEventStoreMock()
	)

	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) (err error) {
			return eventFn(0, false, evnt.StepOutcomeEvent{
				Key:         wid,
				OutcomeType: wfl.OutcomeTypeExecution,
				OutcomeSeq:  1,
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
			IdempotencyWindowStr: "support.FIFOSet cap=1000 len=0 []",
			ChptmanStr:           "checkpoint.Manager partitions=1 chptSize=100 [p0=\"chpt:{-1 -1}\"]",
			TrackerStr:           "runtime.PositionTracker partitions=1 [p0:NextN=0] tracked=none",
		},
	}
}
