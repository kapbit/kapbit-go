package executor_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/test"

	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type emitRejectedTestCases struct{}

var EmitRejectedTestCases = emitRejectedTestCases{}

func (c emitRejectedTestCases) ShouldSuccess() EmitRejectedTestCase {
	name := "Should emit a rejected event and complete the slot on success"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(1)
	var emittedEvent evnt.Event
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			emittedEvent = event
			return nil
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref:    rtm.NewWorkflowRef(workflow, 0),
			Reason: "test reason",
		},
		Want: EmitRejectedWant{
			Check: func(t *testing.T, ctx context.Context) {
				deadLetterEvent := &evnt.WorkflowRejectedEvent{
					NodeID:    "node_id",
					Key:       wfl.ID("wfl-1"),
					Type:      wfl.Type("type-a"),
					Reason:    "test reason",
					Timestamp: 10,
				}
				asserterror.EqualDeep(t, deadLetterEvent, emittedEvent)
				asserterror.Equal(t, runtime.Running().Count(), 0)
			},
			mock: mcks,
		},
	}
}

func (c emitRejectedTestCases) EncodeError() EmitRejectedTestCase {
	name := "Should panic on EncodeError"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(5)
	err := codes.NewEncodeError(reflect.TypeFor[wfl.Input](), "test reason",
		errors.New("test error"))
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitRejectedWant{
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c emitRejectedTestCases) NoActiveWriter() EmitRejectedTestCase {
	name := "Should panic on EncodeError"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(2)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindFenced)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitRejectedWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, ctx.Err(), context.Canceled)
				asserterror.Equal(t, runtime.Running().Count(), 1)
			},
			mock: mcks,
		},
	}
}

func (c emitRejectedTestCases) Rejection() EmitRejectedTestCase {
	name := "Should panic on PersistenceError (Rejection)"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(5)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindRejection)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitRejectedWant{
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c emitRejectedTestCases) UnexpectedKind() EmitRejectedTestCase {
	name := "Should panic on unexpected PersistenceError kind (Other)"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(5)
	err := codes.NewPersistenceError(errors.New("test error"),
		codes.PersistenceKindOther)
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitRejectedWant{
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c emitRejectedTestCases) UnexpectedError() EmitRejectedTestCase {
	name := "Should panic on unexpected error"

	workflow, runtime, timeProvider, emitter, mcks := c.testComponents(5)
	err := errors.New("test error")
	emitter.RegisterEmit(
		func(ctx context.Context, event evnt.Event) error {
			return err
		},
	)
	return EmitRejectedTestCase{
		Name: name,
		Setup: EmitRejectedSetup{
			Runtime: runtime,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
		},
		Params: EmitRejectedParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: EmitRejectedWant{
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c emitRejectedTestCases) testComponents(specCalls int) (
	workflow mock.WorkflowMock,
	runtime *rtm.Runtime,
	timeProvider mock.TimeProviderMock,
	emitter mock.EmitterMock,
	mcks []*mok.Mock,
) {
	runtime = c.runtime()
	workflow = c.workflow(specCalls)
	timeProvider = c.timeProvider()
	emitter = mock.NewEmitterMock()
	mcks = []*mok.Mock{workflow.Mock, timeProvider.Mock, emitter.Mock}
	return
}

func (c emitRejectedTestCases) runtime() *rtm.Runtime {
	var (
		counter = support.NewCounterWith(1, 1)
		window  = support.NewFIFOSet[wfl.ID](1)
	)
	runtime := rtm.New(test.NopLogger, counter, window, nil)
	window.Add(wfl.ID("wfl-1"))
	return runtime
}

func (c emitRejectedTestCases) workflow(specCalls int) mock.WorkflowMock {
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

func (c emitRejectedTestCases) timeProvider() mock.TimeProviderMock {
	return mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 })
}
