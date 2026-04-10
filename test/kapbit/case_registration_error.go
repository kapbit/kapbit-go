package kapbit

import (
	"context"
	"errors"

	kapbit "github.com/kapbit/kapbit-go"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func RegistrationErrorTestCase() TestCase {
	name := "Should return KapbitError when RegisterWorkflow returns an error"

	var (
		workflw = mock.NewWorkflowMock()
		err     = errors.New("registration error")

		factory = mock.NewFactoryMock().RegisterNew(
			func(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error) {
				return workflw, nil
			},
		)
		service = mock.NewEngineMock().RegisterRegisterWorkflow(
			func(ctx context.Context, workflow wfl.Workflow) (ref *rtm.WorkflowRef, error error) {
				return nil, err
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
				Gate: &support.IngressGate{},
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
			Error: kapbit.NewKapbitError(err),
		},
	}
}
