package kapbit

import (
	"context"

	kapbit "github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

func ClosedTestCase() TestCase {
	name := "Should return ErrClosed when Kapbit is closed"

	var (
		timeProvider = mock.NewTimeProviderMock()
		mcks         = []*mok.Mock{timeProvider.Mock}
	)
	return TestCase{
		Name: name,
		Setup: Setup{
			Tools: kapbit.Tools{
				Factory: mock.NewFactoryMock(),
				Engine: mock.NewEngineMock(),
				Worker: mock.NewWorkerMock().RegisterStart(
					func(ctx context.Context) {},
				),
			},
			Options: kapbit.Options{
				TimeProvider: timeProvider,
				Logger:       test.NopLogger,
			},
			mock: mcks,
			Prepare: func(k *kapbit.Kapbit) {
				k.Close()
			},
		},
		Params: wfl.Params{
			ID:    wfl.ID("wfl-1"),
			Type:  wfl.Type("type-a"),
			Input: wfl.Input("input-1"),
		},
		Want: Want{
			Error: kapbit.NewKapbitError(kapbit.ErrClosed),
		},
	}
}
