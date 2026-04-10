package kapbit

import (
	"context"
	"errors"
	"testing"

	kapbit "github.com/kapbit/kapbit-go"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func ExecutionErrorTestCase() TestCase {
	name := "Should return KapbitError when ExecWorkflow returns a generic error and keep gate open"

	var (
		workflw = mock.NewWorkflowMock()
		reslt   = wfl.Result("")
		err     = errors.New("execution error")
		gate    = &support.IngressGate{}

		factory = mock.NewFactoryMock().RegisterNew(
			func(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error) {
				return workflw, nil
			},
		)
		service = mock.NewEngineMock().RegisterRegisterWorkflow(
			func(ctx context.Context, workflow wfl.Workflow) (ref *rtm.WorkflowRef, err error) {
				return rtm.NewWorkflowRef(workflow, 0), nil
			},
		).RegisterExecWorkflow(
			func(ctx context.Context, ref *rtm.WorkflowRef) (result wfl.Result, error error) {
				return reslt, err
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNow(
			func() int64 { return 0 },
		)

		mcks = []*mok.Mock{
			workflw.Mock, factory.Mock, service.Mock, timeProvider.Mock,
		}
	)
	return TestCase{
		Name: name,
		Setup: Setup{
			Tools: kapbit.Tools{
				Factory: factory,
				Engine: service,
				Worker: mock.NewWorkerMock().RegisterStart(
					func(ctx context.Context) {},
				),
				Gate: gate,
			},
			Options: kapbit.Options{
				TimeProvider: timeProvider,
				Logger:       test.NopLogger,
			},
			mock: mcks,
		},
		Params: wfl.Params{
			ID:    wfl.ID("wfl-1"),
			Type:  wfl.Type("type-a"),
			Input: wfl.Input("input-1"),
		},
		Want: Want{
			Result: reslt,
			Error:  kapbit.NewKapbitError(err),
			Check: func(t *testing.T, k *kapbit.Kapbit) {
				if gate.Closed() {
					t.Error("gate should remain open")
				}
			},
		},
	}
}
