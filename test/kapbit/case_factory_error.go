package kapbit

import (
	"context"
	"errors"

	kapbit "github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func FactoryErrorTestCase() TestCase {
	name := "Should return KapbitError when Factory.New returns an error"

	var (
		err     = errors.New("factory error")
		factory = mock.NewFactoryMock().RegisterNew(
			func(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error) {
				return nil, err
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNow(
			func() int64 { return 0 },
		)

		mcks = []*mok.Mock{
			factory.Mock, timeProvider.Mock,
		}
	)
	return TestCase{
		Name: name,
		Setup: Setup{
			Tools: kapbit.Tools{
				Factory: factory,
				Engine:  mock.NewEngineMock(),
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
			Error: kapbit.NewKapbitError(err),
		},
	}
}
