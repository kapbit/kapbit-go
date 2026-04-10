package handler

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	"github.com/kapbit/kapbit-go/test/mock"
)

func ScenarioTestCase() TestCase {
	name := "Concurrent scenario with multi-partition and interleaved events"

	var (
		chpts          = []chptsup.Tuple{{N: -1, V: -1}, {N: -1, V: -1}}
		partitions     = 2
		tracker        = runtime.NewPositionTracker(chpts)
		chptman        = chptsup.NewManager(test.NopLogger, chpts, 1)
		offsetSupplier = NewOffsetSupplier(partitions)
	)

	workflows := GenerateRandomEvents(5)
	scenario := NewScenario(partitions, workflows)
	tc := TestCase{
		Name: name,
		Setup: TestSetup{
			Store: mock.NewEventStoreMock().RegisterPartitionCount(
				func() int {
					return partitions
				},
			).RegisterNSaveEvent(scenario.EventsCount(),
				func(ctx context.Context, partition int, event evnt.Event) (offset int64, err error) {
					scenario.AppendEvent(partition, event)
					return offsetSupplier.Next(partition), nil
				},
			).RegisterNSaveCheckpoint(1000,
				func(ctx context.Context, partition int, chpt int64) error {
					scenario.AppendChpt(partition, chpt)
					return nil
				},
			),
			Tracker: tracker,
			Chptman: chptman,
		},
		Scenario: scenario,
		Want: Want{
			Check: func(t *testing.T, config TestSetup) {
				scenario.Verify(t)
			},
		},
	}
	return tc
}
