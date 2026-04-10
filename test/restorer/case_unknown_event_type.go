package restorer

import (
	"context"

	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
)

func UnknownEventTypeTestCase() TestCase {
	name := "Should return an error on unknown event type"

	store := mock.NewEventStoreMock()
	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) error {
			if err := chptFn(2); err != nil {
				return err
			}
			return eventFn(1, false, "string type")
		},
	).RegisterNPartitionCount(2,
		func() int { return 1 },
	)

	return TestCase{
		Name: name,
		Config: TestConfig{
			Store: store,
			Opts: []rtm.SetOption{
				rtm.WithLogger(test.NopLogger),
			},
		},
		Want: Want{
			Error: rtm.NewRestoreError(
				rtm.NewPartitionError(0,
					rtm.NewUnknownEventTypeError(0, 1, "string type")),
			),
		},
	}
}
