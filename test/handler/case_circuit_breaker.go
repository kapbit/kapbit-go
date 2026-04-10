package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/persist"
	"github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/circbrk-go"
)

// fastCBOptions returns WithCircuitBreaker options that open the CB after a
// single failure and keep it open for 24 hours, making tests deterministic.
func fastCBOptions() persist.SetOption {
	return persist.WithCircuitBreaker(
		circbrk.WithWindowSize(1),
		circbrk.WithFailureRate(1.0),
		circbrk.WithOpenDuration(24*time.Hour),
		circbrk.WithSuccessThreshold(1),
	)
}

// CBOpenSystemEventTestCase verifies that after a store failure trips the CB,
// the next system-event Handle call returns CircuitBreakerOpenError.
// Also covers: WithCircuitBreaker, reportCircuitBreakerStateChange (Open state).
func CBOpenSystemEventTestCase() TestCase {
	name := "Should return CircuitBreakerOpenError for system events when CB is open"

	var (
		storeErr = errors.New("store failure")
		chpts    = []chptsup.Tuple{{N: -1, V: -1}}
		event    = evnt.ActiveWriterEvent{NodeID: "node-1", PartitionIndex: 0}
	)

	return TestCase{
		Name: name,
		Setup: TestSetup{
			Store: mock.NewEventStoreMock().RegisterPartitionCount(
				func() int { return 1 },
			).RegisterSaveEvent(
				// First call: store fails → CB trips (Closed → Open).
				func(_ context.Context, _ int, _ evnt.Event) (int64, error) {
					return 0, storeErr
				},
			),
			Tracker: runtime.NewPositionTracker(chpts),
			Chptman: chptsup.NewManager(test.NopLogger, chpts, 1),
		},
		HandlerOptions: []persist.SetOption{fastCBOptions()},
		Events: []evnt.Event{
			event, // trips the CB
			event, // blocked by the open CB
		},
		Want: Want{
			Errors: []error{
				storeErr,
				codes.NewCircuitBreakerOpenError("storage"),
			},
		},
	}
}

// CBOpenWorkflowEventTestCase verifies that a non-created workflow event is
// blocked with CircuitBreakerOpenError when the partition's CB is open.
func CBOpenWorkflowEventTestCase() TestCase {
	name := "Should return CircuitBreakerOpenError for workflow events when CB is open"

	var (
		storeErr = errors.New("store failure")
		chpts    = []chptsup.Tuple{{N: -1, V: -1}}
		created  = evnt.WorkflowCreatedEvent{
			NodeID: "node-1", Key: wfl.ID("wfl-1"), Type: wfl.Type("type-a"), Input: "input",
		}
		outcome = evnt.StepOutcomeEvent{
			NodeID:      "node-1",
			Key:         wfl.ID("wfl-1"),
			Type:        wfl.Type("type-a"),
			OutcomeSeq:  wfl.OutcomeSeq(0),
			OutcomeType: wfl.OutcomeTypeExecution,
			Outcome:     mock.NewOutcomeMock(),
		}
	)

	return TestCase{
		Name: name,
		Setup: TestSetup{
			Store: mock.NewEventStoreMock().RegisterPartitionCount(
				func() int { return 1 },
			).RegisterSaveEvent(
				// WorkflowCreated succeeds → position reserved on partition 0.
				func(_ context.Context, _ int, _ evnt.Event) (int64, error) {
					return 10, nil
				},
			).RegisterSaveEvent(
				// StepOutcome fails → CB trips.
				func(_ context.Context, _ int, _ evnt.Event) (int64, error) {
					return 0, storeErr
				},
			),
			Tracker: runtime.NewPositionTracker(chpts),
			Chptman: chptsup.NewManager(test.NopLogger, chpts, 1),
		},
		HandlerOptions: []persist.SetOption{fastCBOptions()},
		Events: []evnt.Event{
			created, // succeeds → position tracked
			outcome, // fails → CB opens
			outcome, // blocked by the open CB
		},
		Want: Want{
			Errors: []error{
				nil,
				storeErr,
				codes.NewCircuitBreakerOpenError("storage"),
			},
			Check: func(t *testing.T, setup TestSetup) {
				// wfl-1 position still tracked (outcome was not terminal).
				_, pst := setup.Tracker.Position(wfl.ID("wfl-1"))
				asserterror.Equal(t, pst, true)
			},
		},
	}
}

// CBOpenWorkflowCreatedTestCase verifies that WorkflowCreated returns
// CantResolvePartitionError when all partition CBs are open.
func CBOpenWorkflowCreatedTestCase() TestCase {
	name := "Should return CantResolvePartitionError for WorkflowCreated when CB is open"

	var (
		storeErr = errors.New("store failure")
		chpts    = []chptsup.Tuple{{N: -1, V: -1}}
	)

	return TestCase{
		Name: name,
		Setup: TestSetup{
			Store: mock.NewEventStoreMock().RegisterPartitionCount(
				func() int { return 1 },
			).RegisterSaveEvent(
				// System event fails → the only partition's CB opens.
				func(_ context.Context, _ int, _ evnt.Event) (int64, error) {
					return 0, storeErr
				},
			),
			Tracker: runtime.NewPositionTracker(chpts),
			Chptman: chptsup.NewManager(test.NopLogger, chpts, 1),
		},
		HandlerOptions: []persist.SetOption{fastCBOptions()},
		Events: []evnt.Event{
			// System event: trips the only partition's CB.
			evnt.ActiveWriterEvent{NodeID: "node-1", PartitionIndex: 0},
			// WorkflowCreated: resolver finds no open partition.
			evnt.WorkflowCreatedEvent{
				NodeID: "node-1", Key: wfl.ID("wfl-2"), Type: wfl.Type("type-a"), Input: "input",
			},
		},
		Want: Want{
			Errors: []error{
				storeErr,
				persist.NewCantResolvePartitionError(
					codes.NewCircuitBreakerOpenError(persist.EventStoreServiceName),
				),
			},
		},
	}
}
