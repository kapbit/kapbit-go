package workflow

import (
	"context"
	"errors"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"

	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func FinErrorTestCase() TestCase {
	name := "Should return a result builder error"

	var (
		err      = errors.New("can't build a result")
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(1,
			func() bool { return false },
		)
		executeAction = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return nil, err
			},
		)
		timeProvider = mock.NewTimeProviderMock()
		emitter      = mock.NewEmitterMock()
	)

	progress := wfl.NewProgress("wfl-1", wfl.Type("my-wfl"))
	progress.RecordExecutionOutcome(0, outcome1)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID:    wfl.ID("wfl-id"),
				Def:   wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder, wfl.Step{Execute: executeAction}),
				Input: wfl.Input("input"),
			},
			Progress: progress,
			Tools: &evnt.Kit{
				Time:    timeProvider,
				Emitter: emitter,
			},
			Logger: test.NopLogger,
		},
		Want: Want{
			State:  wfl.Finalizing,
			Result: nil,
			Error:  wfl.NewCompletionError(err),
		},
		mock: []*mok.Mock{
			outcome1.Mock, timeProvider.Mock, emitter.Mock,
		},
	}
}

func FinNoResultTestCase(t *testing.T) TestCase {
	name := "Should be able to return nil result"

	var (
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return nil, nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNow(
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
					Result:    nil,
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
			Result: nil,
			Error:  nil,
		},
		mock: []*mok.Mock{timeProvider.Mock, emitter.Mock},
	}
}
