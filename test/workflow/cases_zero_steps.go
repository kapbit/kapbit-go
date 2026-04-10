package workflow

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"

	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func ZeroStepsTestCase(t *testing.T) TestCase {
	name := "Should perform a Workflow with zero steps"

	var (
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(1,
			func() int64 {
				return 0
			},
		)
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.WorkflowResultEvent{
					NodeID:    "node_id",
					Key:       wfl.ID("wfl-id"),
					Type:      wfl.Type("my-wfl"),
					Result:    wfl.Result("result"),
					Timestamp: 0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		)
	)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID:    wfl.ID("wfl-id"),
				Def:   wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder),
				Input: wfl.Input("input"),
			},
			Progress: wfl.NewProgress("wfl-1", wfl.Type("my-wfl")),
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
			Logger: test.NopLogger,
		},
		Want: Want{
			State:  wfl.Finalizing,
			Result: wfl.Result("result"),
			Error:  nil,
		},
		mock: []*mok.Mock{timeProvider.Mock, emitter.Mock},
	}
}
