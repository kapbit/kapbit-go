package kapbit

import (
	"context"

	kapbit "github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func GateClosedTestCase() TestCase {
	name := "Should return ErrEntryGateClosed when gate is closed"

	var (
		gate         = &support.EntryGate{}
		timeProvider = mock.NewTimeProviderMock()

		mcks = []*mok.Mock{
			timeProvider.Mock,
		}
	)
	gate.Close()
	return TestCase{
		Name: name,
		Setup: Setup{
			Tools: kapbit.Tools{
				Factory: mock.NewFactoryMock(),
				Engine:  mock.NewEngineMock(),
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
			Error: kapbit.NewKapbitError(kapbit.ErrEntryGateClosed),
		},
	}
}
