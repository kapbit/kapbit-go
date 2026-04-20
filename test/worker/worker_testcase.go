package worker_test

import (
	"context"
	"testing"
	"time"

	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
	"github.com/ymz-ncnk/mok"
)

type TestCase struct {
	Name   string
	Config TestConfig
	Ctx    context.Context
	Want   Want
	mock   []*mok.Mock
}

type TestConfig struct {
	Engine mock.EngineMock
	Gate   *support.EntryGate
	Opts   []wrk.SetOption
}

type Want struct {
	Wait  time.Duration
	Check func(*testing.T, context.Context, *wrk.Worker)
}

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		opts := append(tc.Config.Opts, wrk.WithLogger(test.NopLogger))
		worker, err := wrk.New(tc.Config.Engine, tc.Config.Gate, opts...)
		assertfatal.EqualError(t, err, nil)

		go func() {
			ctx := tc.Ctx
			if ctx == nil {
				ctx = t.Context()
			}
			worker.Start(ctx)
		}()

		time.Sleep(tc.Want.Wait)
		if tc.Want.Check != nil {
			tc.Want.Check(t, t.Context(), worker)
		}

		infomap := mok.CheckCalls(tc.mock)
		asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
	})
}

func testComponents() (service mock.EngineMock, gate *support.EntryGate,
	retryCh chan *rtm.WorkflowRef, ref *rtm.WorkflowRef, mcks []*mok.Mock,
) {
	service = mock.NewEngineMock()
	gate = &support.EntryGate{}
	retryCh = make(chan *rtm.WorkflowRef, 1)

	workflow := newWorkflow()
	ref = rtm.NewWorkflowRef(workflow, 0)
	retryCh <- ref

	mcks = []*mok.Mock{service.Mock}
	return
}

func newWorkflow() mock.WorkflowMock {
	progress := wfl.NewProgress("test-wid", "test-type")
	return mock.NewWorkflowMock().RegisterNSpec(20,
		func() wfl.Spec {
			return wfl.Spec{ID: "test-wid", Def: wfl.MustDefinition("test-type", nil)}
		},
	).RegisterNProgress(20,
		func() *wfl.Progress {
			return progress
		},
	)
}
