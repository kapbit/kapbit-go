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

type executeTestCases struct{}

var ExecuteTestCases = executeTestCases{}

func (c executeTestCases) ShouldSuccess() ExecuteTestCase {
	name := "Should execute the workflow and free the slot on success"

	var (
		result   = wfl.Result("test result")
		workflow = mock.NewWorkflowMock().RegisterRun(
			func(ctx context.Context) (wfl.Result, error) {
				return result, nil
			},
		).RegisterSpec(
			func() wfl.Spec { return wfl.Spec{ID: wfl.ID("wfl-1")} },
		)
		counter = support.NewCounterWith(1, 1)
		runtime = rtm.New(test.NopLogger, counter, nil, nil)
	)
	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: runtime,
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(workflow, 0),
		},
		Want: ExecuteWant{
			Result: result, mock: []*mok.Mock{workflow.Mock},
		},
	}
}

func (c executeTestCases) ProtocolViolationError() ExecuteTestCase {
	name := "Should reject workflow on ProtocolViolationError"

	var (
		err = wfl.NewProtocolViolationError(wfl.ID("wfl-1"), wfl.Type("type-a"),
			errors.New("test error"))
		wf, rt, emit, captured, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
	)
	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt, Tools: &evnt.Kit{
				Emitter: emit,
				Time:    mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(wf, 0),
		},
		Want: ExecuteWant{
			Error: err,
			Check: c.checkRejection(captured, rt, err),
			mock:  mcks,
		},
	}
}

func (c executeTestCases) EncodeError() ExecuteTestCase {
	name := "Should reject workflow on EncodeError"

	var (
		err = codes.NewEncodeError(reflect.TypeFor[wfl.Input](), "test reason",
			errors.New("test error"))
		wf, rt, emit, captured, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
	)

	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt,
			Tools: &evnt.Kit{
				Emitter: emit,
				Time:    mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(wf, 0),
		},
		Want: ExecuteWant{
			Error: err,
			Check: c.checkRejection(captured, rt, err),
			mock:  mcks,
		},
	}
}

func (c executeTestCases) NoActiveWriterError() ExecuteTestCase {
	name := "Should close the ctx on PersistenceError (NoActiveWriter)"

	var (
		err = codes.NewPersistenceError(errors.New("test error"),
			codes.PersistenceKindFenced)
		wf, rt, _, _, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
	)

	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt,
			Tools: &evnt.Kit{
				Time: mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(wf, 0),
		},
		Want: ExecuteWant{
			Error: err,
			Check: func(t *testing.T, ctx context.Context) {
				asserterror.Equal(t, ctx.Err(), context.Canceled)
			},
			mock: mcks,
		},
	}
}

func (c executeTestCases) RejectionError() ExecuteTestCase {
	name := "Should reject workflow on PersistenceError (Rejection)"

	var (
		err = codes.NewPersistenceError(errors.New("test error"),
			codes.PersistenceKindRejection)
		wf, rt, emit, captured, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
	)

	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt,
			Tools: &evnt.Kit{
				Emitter: emit,
				Time:    mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(wf, 0),
		},
		Want: ExecuteWant{
			Error: err,
			Check: c.checkRejection(captured, rt, err),
			mock:  mcks,
		},
	}
}

func (c executeTestCases) UnexpectedKind() ExecuteTestCase {
	name := "Should panic on unexpected PersistenceError kind (UnknownState)"

	var (
		err = codes.NewPersistenceError(errors.New("test error"),
			codes.PersistenceKindUnknownState)
		wf, rt, _, _, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
	)

	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt,
			Tools: &evnt.Kit{
				Time: mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{
			Ref: rtm.NewWorkflowRef(wf, 0),
		},
		Want: ExecuteWant{
			Error:       err,
			ShouldPanic: true,
			mock:        mcks,
		},
	}
}

func (c executeTestCases) UnexpectedError() ExecuteTestCase {
	name := "Should retry workflow on unexpected error"

	var (
		err                   = errors.New("unexpected error")
		wf, rt, emit, _, mcks = c.testComponents(
			func(context.Context) (wfl.Result, error) {
				return nil, err
			})
		ref = rtm.NewWorkflowRef(wf, 0)
	)
	wf.RegisterProgress(func() *wfl.Progress {
		return wfl.NewProgress("wfl-1", "type-a")
	})

	return ExecuteTestCase{
		Name: name,
		Setup: ExecuteSetup{
			Runtime: rt,
			Tools: &evnt.Kit{
				Emitter: emit,
				Time:    mock.NewTimeProviderMock().RegisterNow(func() int64 { return 10 }),
			},
		},
		Params: ExecuteParams{Ref: ref},
		Want: ExecuteWant{
			Error: executor.NewWillRetryLaterError(err),
			mock:  mcks,
			Check: func(t *testing.T, ctx context.Context) {
				select {
				case retryRef := <-rt.RetryQueue():
					asserterror.EqualDeep(t, ref, retryRef)
				default:
					t.Errorf("expected retry queue to be non-empty")
				}
			},
		},
	}
}

func (c executeTestCases) testComponents(
	runFn func(context.Context) (wfl.Result, error),
) (mock.WorkflowMock, *rtm.Runtime, mock.EmitterMock, *evnt.Event, []*mok.Mock) {
	var (
		emittedEvent evnt.Event
		runtime      = c.runtime()
		workflow     = c.workflow(runFn)

		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				emittedEvent = event
				return nil
			},
		)
		mcks = []*mok.Mock{workflow.Mock, emitter.Mock}
	)
	return workflow, runtime, emitter, &emittedEvent, mcks
}

func (c executeTestCases) runtime() *rtm.Runtime {
	counter := support.NewCounterWith(1, 1)
	return rtm.New(test.NopLogger, counter, nil, make(chan *rtm.WorkflowRef, 1))
}

func (c executeTestCases) workflow(
	runFn func(context.Context) (wfl.Result, error),
) mock.WorkflowMock {
	return mock.NewWorkflowMock().
		RegisterRun(runFn).
		RegisterNSpec(10, func() wfl.Spec {
			return wfl.Spec{
				ID:  wfl.ID("wfl-1"),
				Def: wfl.MustDefinition(wfl.Type("type-a"), nil),
			}
		}).
		RegisterNodeID(func() string { return "node-1" })
}

func (c executeTestCases) checkRejection(emitted *evnt.Event,
	runtime *rtm.Runtime,
	wantErr error,
) func(*testing.T, context.Context) {
	return func(t *testing.T, _ context.Context) {
		expected := &evnt.WorkflowRejectedEvent{
			NodeID:    "node-1",
			Key:       wfl.ID("wfl-1"),
			Type:      wfl.Type("type-a"),
			Reason:    wantErr.Error(),
			Timestamp: 10,
		}
		asserterror.EqualDeep(t, *emitted, expected)
		asserterror.Equal(t, runtime.Running().Count(), 0)
	}
}
