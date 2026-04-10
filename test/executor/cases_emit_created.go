package executor_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

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

type emitCreatedTestCases struct{}

var EmitCreatedTestCases = emitCreatedTestCases{}

func (c emitCreatedTestCases) ShouldSuccess() EmitCreatedTestCase {
	name := "Should successfully emit created event"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(1)
	var emittedEvent evnt.Event
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			emittedEvent = event
			return nil
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Check: func(t *testing.T, ctx context.Context) {
				createdEvent := evnt.WorkflowCreatedEvent{
					NodeID:    "node_id",
					Key:       wfl.ID("wfl-1"),
					Type:      wfl.Type(""),
					Input:     nil,
					Timestamp: 10,
				}
				asserterror.EqualDeep(t, createdEvent, emittedEvent)
				asserterror.Equal(t, runtime.Running().Count(), 1)
			},
			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) EncodeError() EmitCreatedTestCase {
	name := "Should return error and release slot on EncodeError"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(2)
	err := codes.NewEncodeError(reflect.TypeFor[wfl.Input](), "test reason",
		errors.New("test error"))
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 0)
			},

			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) NoActiveWriter() EmitCreatedTestCase {
	name := "Should close the ctx on PersistenceError, NoActiveWriter"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(2)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindFenced)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, ctx.Err(), context.Canceled)
				asserterror.Equal(t, runtime.Running().Count(), 0)
			},

			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) PersistenceRejection() EmitCreatedTestCase {
	name := "Should return error and release slot on PersistenceError, Rejection"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(2)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindRejection)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, runtime.Running().Count(), 0)
			},

			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) UnexpectedErrorKind() EmitCreatedTestCase {
	name := "Should panic on unexpected PersistenceError kind (Other)"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(1)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindOther)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Error:       err,
			ShouldPanic: true,

			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) UnexpectedError() EmitCreatedTestCase {
	name := "Should panic on unexpected error"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(1)
	err := errors.New("test error")
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitCreatedTestCase{
		Name: name,
		Setup: EmitCreatedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitCreatedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitCreatedWant{
			Error:       err,
			ShouldPanic: true,

			mock: mcks,
		},
	}
}

func (c emitCreatedTestCases) testComponents(specsCalls int) (
	workflow mock.WorkflowMock,
	runtime *rtm.Runtime,
	timeProvider mock.TimeProviderMock,
	emitter mock.EmitterMock,
	mcks []*mok.Mock,
) {
	runtime = c.runtime()
	workflow = c.workflow(specsCalls)
	timeProvider = c.timeProvider()
	emitter = mock.NewEmitterMock()
	mcks = []*mok.Mock{workflow.Mock, timeProvider.Mock, emitter.Mock}
	return
}

func (c emitCreatedTestCases) runtime() *rtm.Runtime {
	var (
		counter = support.NewCounterWith(1, 1)
		window  = support.NewFIFOSet[wfl.ID](1)
	)
	runtime := rtm.New(test.NopLogger, counter, window, nil)
	window.Add(wfl.ID("wfl-1"))
	return runtime
}

func (c emitCreatedTestCases) workflow(specCalls int) mock.WorkflowMock {
	return mock.NewWorkflowMock().RegisterNodeID(
		func() string { return "node_id" },
	).RegisterNSpec(specCalls,
		func() wfl.Spec { return wfl.Spec{ID: wfl.ID("wfl-1")} },
	)
}

func (c emitCreatedTestCases) timeProvider() mock.TimeProviderMock {
	return mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 })
}
