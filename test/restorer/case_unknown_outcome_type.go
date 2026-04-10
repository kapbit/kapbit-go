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

func UnknownOutcomeType() TestCase {
	name := "Should return an error on unknown outcome type"

	store := mock.NewEventStoreMock()
	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) (err error) {
			err = eventFn(1, false, evnt.WorkflowCreatedEvent{
				Key: "wfl-1",
			})
			if err != nil {
				return err
			}
			return eventFn(1, false, evnt.StepOutcomeEvent{
				Key: "wfl-1",
				// Unknown outcome type
				OutcomeType: wfl.OutcomeType(0),
			})
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
					spec := wfl.Spec{
						ID: wfl.ID("wfl-1"),
					}
					return wflimpl.NewWorkflow("node_id", spec, progress, nil, slog.Default()), nil
				},
			),
			Opts: []rtm.SetOption{
				rtm.WithLogger(test.NopLogger),
			},
		},
		Want: Want{
			Error: rtm.NewRestoreError(
				rtm.NewPartitionError(0,
					rtm.NewUnknownOutcomeTypeError(0, 1, wfl.OutcomeType(0), "wfl-1"),
				),
			),
		},
	}
}
