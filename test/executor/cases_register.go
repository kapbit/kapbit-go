package executor_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	executor "github.com/kapbit/kapbit-go/executor"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type registerTestCases struct{}

var RegisterTestCases = registerTestCases{}

func (c registerTestCases) ShouldSuccess() RegisterTestCase {
	name := "Should register the Workflow"

	workflow, runtime, tools, mcks := c.testComponents(2)
	var emittedEvent evnt.Event
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			emittedEvent = event
			return nil
		})
	ref := rtm.NewSavedWorkflowRef(workflow, 0)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Ref: ref,
			Check: func(t *testing.T, ctx context.Context) {
				createdEvent := evnt.WorkflowCreatedEvent{
					NodeID:    "node_id",
					Key:       wfl.ID("wfl-1"),
					Type:      wfl.Type("type-a"),
					Input:     nil,
					Timestamp: 10,
				}
				asserterror.EqualDeep(t, emittedEvent, createdEvent)
				asserterror.Equal(t, runtime.Running().Count(), 1)
				asserterror.Equal(t, runtime.IdempotencyWindow().Contains("wfl-1"), true)
				asserterror.Equal(t, ref.Saved(), true)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) NoAvailaleSlots() RegisterTestCase {
	name := "Should return an error if there is no available slots"

	workflow, runtime, _, mcks := c.testComponents(1)
	runtime.Running().Increment()
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error: executor.ErrNoSlotsAvailable,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 1)
				asserterror.Equal(t, runtime.IdempotencyWindow().Contains("wfl-1"), false)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) IdempotencyViolation() RegisterTestCase {
	name := "Should return an error on idempotency violation"

	workflow, runtime, _, mcks := c.testComponents(1)
	runtime.IdempotencyWindow().Add("wfl-1")
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error: executor.ErrIdempotencyViolation,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 0)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) EncodeError() RegisterTestCase {
	name := "Should do nothing on EncodeError"

	workflow, runtime, tools, mcks := c.testComponents(6)
	err := codes.NewEncodeError(reflect.TypeFor[wfl.Input](), "test reason",
		errors.New("test error"))
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 0)
				asserterror.Equal(t, runtime.IdempotencyWindow().Contains("wfl-1"), false)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) NoActiveWriter() RegisterTestCase {
	name := "Should close the ctx on PersistenceError (NoActiveWriter)"

	workflow, runtime, tools, mcks := c.testComponents(6)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindFenced,
	)
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, ctx.Err(), context.Canceled)
				asserterror.Equal(t, runtime.Running().Count(), 0)
				asserterror.Equal(t, runtime.IdempotencyWindow().Contains("wfl-1"), false)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) RejectionError() RegisterTestCase {
	name := "Should do nothing on PersistenceError (Rejection)"

	workflow, runtime, tools, mcks := c.testComponents(6)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindRejection,
	)
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 0)
				asserterror.Equal(t, runtime.IdempotencyWindow().Contains("wfl-1"), false)
				select {
				case <-runtime.RetryQueue():
					t.Error("expected ref not to be in retry queue")
				default:
				}
			},
			mock: mcks,
		},
	}
}

func (c registerTestCases) UnexpectedKind() RegisterTestCase {
	name := "Should panic on unexpected PersistenceError kind (Other)"

	workflow, runtime, tools, mcks := c.testComponents(8)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindOther,
	)
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error:       err,
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c registerTestCases) UnexpectedError() RegisterTestCase {
	name := "Should panic on unexpected error"

	workflow, runtime, tools, mcks := c.testComponents(8)
	err := errors.New("test error")
	tools.Emitter.(mock.EmitterMock).RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return RegisterTestCase{
		Name: name,
		Setup: Setup{
			Runtime: runtime,
			Tools:   tools,
		},
		Params: Params{Workflow: workflow},
		Want: RegisterWant{
			Error:       executor.NewWillRetryLaterError(err),
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c registerTestCases) testComponents(specCalls int) (
	mock.WorkflowMock, *rtm.Runtime, *evnt.Kit, []*mok.Mock,
) {
	var (
		runtime      = c.runtime()
		workflow     = c.workflow(specCalls)
		timeProvider = c.timeProvider()
		emitter      = mock.NewEmitterMock()
	)
	tools := &evnt.Kit{
		Emitter: emitter,
		Time:    timeProvider,
	}
	mcks := []*mok.Mock{workflow.Mock, emitter.Mock, timeProvider.Mock}
	return workflow, runtime, tools, mcks
}

func (c registerTestCases) runtime() *rtm.Runtime {
	counter := support.NewCounter(1)
	window := support.NewFIFOSet[wfl.ID](1)
	return rtm.New(test.NopLogger, counter, window, make(chan *rtm.WorkflowRef, 1))
}

func (c registerTestCases) workflow(specCalls int) mock.WorkflowMock {
	return mock.NewWorkflowMock().RegisterNodeID(
		func() string { return "node_id" },
	).RegisterNSpec(specCalls,
		func() wfl.Spec {
			return wfl.Spec{
				ID:  wfl.ID("wfl-1"),
				Def: wfl.MustDefinition(wfl.Type("type-a"), nil),
			}
		},
	)
}

func (c registerTestCases) timeProvider() mock.TimeProviderMock {
	return mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 })
}
