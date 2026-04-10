package restorer

import (
	"context"
	"errors"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

func FactoryErrorTestCase() TestCase {
	name := "Should return an error on workflow factory failure"

	var (
		err   = errors.New("factory error")
		store = mock.NewEventStoreMock()
	)

	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) error {
			return eventFn(1, false, evnt.WorkflowCreatedEvent{})
		},
	).RegisterNPartitionCount(2,
		func() int { return 1 },
	)

	return TestCase{
		Name: name,
		Config: TestConfig{
			Store: store,
			Factory: mock.NewFactoryMock().RegisterNew(
				func(nodeID string, params wfl.Params, progress *wfl.Progress) (
					wfl.Workflow, error,
				) {
					return nil, err
				},
			),
			Opts: []rtm.SetOption{
				rtm.WithLogger(test.NopLogger),
			},
		},
		Want: Want{
			Error: rtm.NewRestoreError(rtm.NewPartitionError(0, err)),
		},
	}
}
