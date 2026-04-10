package restorer

import (
	"context"
	"log/slog"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	wflimpl "github.com/kapbit/kapbit-go/workflow/impl"
)

func RejectedEventTestCase() TestCase {
	name := "Should correctly handle WorkflowRejectedEvent during restoration"

	var (
		wid   = wfl.ID("wfl-rejected")
		store = mock.NewEventStoreMock()
	)

	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) (err error) {
			err = eventFn(0, false, evnt.WorkflowCreatedEvent{
				Key:  wid,
				Type: "test-workflow",
			})
			if err != nil {
				return err
			}
			return eventFn(1, false, &evnt.WorkflowRejectedEvent{
				Key: wid,
			})
		},
	).RegisterNPartitionCount(2, func() int { return 1 })

	return TestCase{
		Name: name,
		Config: TestConfig{
			NodeID: "node-1",
			Store:  store,
			Factory: mock.NewFactoryMock().RegisterNew(
				func(nodeID string, params wfl.Params, progress *wfl.Progress) (
					wfl.Workflow, error,
				) {
					spec := wfl.Spec{ID: wid}
					return wflimpl.NewWorkflow(nodeID, spec, progress, nil, slog.Default()), nil
				},
			),
			Opts: []rtm.SetOption{
				rtm.WithLogger(test.NopLogger),
			},
		},
		Want: Want{
			CounterValue:         0,
			IdempotencyWindowStr: "support.FIFOSet cap=1000 len=1 [wfl-rejected]",
			ChptmanStr:           "checkpoint.Manager partitions=1 chptSize=100 [p0=\"chpt:{-1 -1} [{0 0},{0 0}]\"]",
			TrackerStr:           "runtime.PositionTracker partitions=1 [p0:NextN=1] tracked=none",
		},
	}
}
