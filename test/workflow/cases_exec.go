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

func ExecTestCase(t *testing.T) TestCase {
	name := "Should perform execution actions and return result"

	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(3,
			func() int64 {
				return 0
			},
		)
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome1,
					Timestamp:   0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  1,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome2,
					Timestamp:   0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
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
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
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
		mock: []*mok.Mock{
			outcome1.Mock, outcome2.Mock, timeProvider.Mock,
			emitter.Mock,
		},
	}
}

func ExecContinueTestCase(t *testing.T) TestCase {
	name := "Should continue execution and return result"

	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				return 0
			},
		)
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  1,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome2,
					Timestamp:   0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
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
	progress := wfl.NewProgress("wfl-1", wfl.Type("my-wfl"))
	progress.RecordExecutionOutcome(wfl.OutcomeSeq(0), outcome1)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
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
			Result: wfl.Result("result"),
			Error:  nil,
		},
		mock: []*mok.Mock{
			outcome1.Mock, outcome2.Mock, timeProvider.Mock,
			emitter.Mock,
		},
	}
}

func ExecEmitErrorTestCase() TestCase {
	name := "Should return an emit error during the execution"

	var (
		outcome1       = mock.NewOutcomeMock()
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				panic("should not be called")
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(1,
			func() int64 {
				return 0
			},
		)
		err     = errors.New("emit error")
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				return err
			},
		)
	)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), nil,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
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
			State:  wfl.Executing,
			Result: nil,
			Error:  err,
		},
		mock: []*mok.Mock{
			outcome1.Mock, timeProvider.Mock, emitter.Mock,
		},
	}
}

func ExecErrorCase() TestCase {
	name := "Should return a step execution error"

	var (
		err            = errors.New("can't execute")
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return nil, err
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				panic("should not be called")
			},
		)
	)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), nil,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
				Input: wfl.Input("input"),
			},
			Progress: wfl.NewProgress("wfl-1", wfl.Type("my-wfl")),
			Tools:    &evnt.Kit{},
			Logger:   test.NopLogger,
		},
		Want: Want{
			State:  wfl.Executing,
			Result: nil,
			Error:  wfl.NewExecutionError(err),
		},
		mock: []*mok.Mock{},
	}
}

func ExecNoOutcomeTestCase(t *testing.T) TestCase {
	name := "Should be able to return nil execution outcome"

	var (
		outcome1       = wfl.NoOutcome
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				return 0
			},
		)
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome1,
					Timestamp:   0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
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
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1},
				),
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

func ExecNoOutcomeStateTestCase(t *testing.T) TestCase {
	name := "Should be in Executing state when last execution outcome is NoOutcome"

	var (
		outcome1 = wfl.NoOutcome
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		executeAction1 = mock.NewActionMock[wfl.Outcome]() // already executed
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)
		timeProvider = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				return 0
			},
		)
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  1,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome2,
					Timestamp:   0,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
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

	progress := wfl.NewProgress("wfl-1", wfl.Type("my-wfl"))
	progress.RecordExecutionOutcome(0, outcome1)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
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
			Result: wfl.Result("result"),
			Error:  nil,
		},
		mock: []*mok.Mock{outcome2.Mock, timeProvider.Mock, emitter.Mock},
	}
}

func ExecPartialTestCase(t *testing.T) TestCase {
	name := "Should stop on failed execution"

	var (
		result   = wfl.Result("result")
		outcome1 = mock.NewOutcomeMock().RegisterFailure(
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterFailure(
			func() bool { return false },
		)
		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				panic("should not be called")
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				panic("should not be called")
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return result, nil
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
	progress := wfl.NewProgress("wfl-1", wfl.Type("my-wfl"))
	progress.RecordExecutionOutcome(0, outcome1)
	progress.RecordExecutionOutcome(1, outcome2)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1},
					wfl.Step{Execute: executeAction2},
				),
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
			Result: result,
			Error:  nil,
		},
		mock: []*mok.Mock{
			outcome1.Mock, outcome2.Mock, timeProvider.Mock,
			emitter.Mock,
		},
	}
}
