package kapbit

import (
	"context"
	"fmt"

	kapbit "github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/codes"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func FencedTestCase() TestCase {
	name := "Should return ErrFenced when Kapbit is fenced"

	var (
		timeProvider = mock.NewTimeProviderMock().RegisterNow(
			func() int64 { return 0 },
		)
		factory = mock.NewFactoryMock().RegisterNew(
			func(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error) {
				return mock.NewWorkflowMock(), nil
			},
		)
		service = mock.NewEngineMock()
		mcks    = []*mok.Mock{timeProvider.Mock, factory.Mock, service.Mock}
	)
	service.RegisterRegisterWorkflow(func(ctx context.Context,
		workflow wfl.Workflow,
	) (ref *rtm.WorkflowRef, err error) {
		err = codes.NewPersistenceError(fmt.Errorf("fenced"),
			codes.PersistenceKindFenced,
		)
		return
	})
	return TestCase{
		Name: name,
		Setup: Setup{
			Tools: kapbit.Tools{
				Factory: factory,
				Engine:  service,
				Worker: mock.NewWorkerMock().RegisterStart(
					func(ctx context.Context) {},
				),
				Gate: &support.EntryGate{},
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
			Error: kapbit.NewKapbitError(kapbit.ErrFenced),
		},
	}
}
