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

func CompTestCase(t *testing.T) TestCase {
	name := "Should perform compensation actions and return result"

	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome3 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome4 = mock.NewOutcomeMock().RegisterNFailure(5,
			func() bool { return true },
		)

		compOutcome1 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		compOutcome2 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		executeAction3 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome3, nil
			},
		)
		compensateAction3 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome2, nil
			},
		)
		executeAction4 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome4, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(8,
			func() int64 {
				i++
				return i
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
					Timestamp:   1,
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
					Timestamp:   2,
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
					OutcomeSeq:  2,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome3,
					Timestamp:   3,
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
					OutcomeSeq:  3,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome4,
					Timestamp:   4,
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
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     compOutcome2,
					Timestamp:   5,
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
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     wfl.NoOutcome,
					Timestamp:   6,
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
					OutcomeSeq:  2,
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     compOutcome1,
					Timestamp:   7,
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
					Timestamp: 8,
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
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
					wfl.Step{Execute: executeAction2},
					wfl.Step{Execute: executeAction3, Compensate: compensateAction3},
					wfl.Step{Execute: executeAction4},
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
			outcome1.Mock, outcome2.Mock, compOutcome1.Mock, compOutcome2.Mock,
			timeProvider.Mock, emitter.Mock,
		},
	}
}

func CompContinueTestCase(t *testing.T) TestCase {
	name := "Should continue compensation and return result"

	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(1,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(1,
			func() bool { return false },
		)
		outcome3 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return true },
		)

		compOutcome1 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)
		compOutcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		compensateAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome2, nil
			},
		)
		executeAction3 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome3, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				i++
				return i
			},
		)

		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  1,
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     compOutcome1,
					Timestamp:   1,
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
					Timestamp: 2,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		)
	)
	progress := wfl.NewProgress("wfl-1", "my-wfl")
	progress.RecordExecutionOutcome(0, outcome1)
	progress.RecordExecutionOutcome(1, outcome2)
	progress.RecordExecutionOutcome(2, outcome3)

	progress.RecordCompensationOutcome(0, compOutcome2)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
					wfl.Step{Execute: executeAction2, Compensate: compensateAction2},
					wfl.Step{Execute: executeAction3},
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
			outcome1.Mock, outcome2.Mock, outcome3.Mock, compOutcome1.Mock,
			compOutcome2.Mock, timeProvider.Mock, emitter.Mock,
		},
	}
}

func CompEmitErrorTestCase(t *testing.T) TestCase {
	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return true },
		)
		compOutcome1 = mock.NewOutcomeMock()

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(3,
			func() int64 {
				i++
				return i
			},
		)
		err     = errors.New("can't emit compensation outcome")
		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome1,
					Timestamp:   1,
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
					Timestamp:   2,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		).RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				return err
			},
		)
	)

	return TestCase{
		Name: "Should return an emit error during the compensation",
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), nil,
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
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
			State:  wfl.Compensating,
			Result: nil,
			Error:  err,
		},
		mock: []*mok.Mock{
			outcome1.Mock, outcome2.Mock, compOutcome1.Mock, timeProvider.Mock,
			emitter.Mock,
		},
	}
}

func CompErrorTestCase(t *testing.T) TestCase {
	name := "Should return a step compensation error"

	var (
		err      = errors.New("can't compensate")
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return true },
		)

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return nil, err
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				i++
				return i
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
					Timestamp:   1,
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
					Timestamp:   2,
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
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), nil,
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
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
			State:  wfl.Compensating,
			Result: nil,
			Error:  wfl.NewCompensationError(err),
		},
		mock: []*mok.Mock{
			outcome1.Mock, outcome2.Mock, timeProvider.Mock,
			emitter.Mock,
		},
	}
}

func CompNoOutcomeStateTestCase(t *testing.T) TestCase {
	name := "Should be in Compensating state when last compensation outcome is NoOutcome"

	var (
		execOutcome1 = mock.NewOutcomeMock().RegisterFailure(
			func() bool { return false },
		)
		execOutcome2 = mock.NewOutcomeMock().RegisterFailure(
			func() bool { return true },
		)
		compOutcome1 = wfl.NoOutcome

		executeAction1    = mock.NewActionMock[wfl.Outcome]() // already executed
		executeAction2    = mock.NewActionMock[wfl.Outcome]() // already executed
		compensateAction1 = mock.NewActionMock[wfl.Outcome]() // already compensated with NoOutcome

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

	progress := wfl.NewProgress("wfl-1", wfl.Type("my-wfl"))
	progress.RecordExecutionOutcome(0, execOutcome1)
	progress.RecordExecutionOutcome(1, execOutcome2)
	progress.RecordCompensationOutcome(0, compOutcome1)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
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
		mock: []*mok.Mock{timeProvider.Mock, emitter.Mock},
	}
}

func CompPartialTestCase(t *testing.T) TestCase {
	name := "Should stop on failed compensation"

	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return false },
		)
		outcome3 = mock.NewOutcomeMock().RegisterNFailure(4,
			func() bool { return true },
		)

		compOutcome1 = mock.NewOutcomeMock()
		compOutcome2 = mock.NewOutcomeMock().RegisterNFailure(3,
			func() bool { return true },
		)

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome1, nil
			},
		)
		executeAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome2, nil
			},
		)
		compensateAction2 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome2, nil
			},
		)
		executeAction3 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome3, nil
			},
		)
		resultBuilder = mock.NewActionMock[wfl.Result]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Result, error) {
				return wfl.Result("result"), nil
			},
		)

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(5,
			func() int64 {
				i++
				return i
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
					Timestamp:   1,
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
					Timestamp:   2,
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
					OutcomeSeq:  2,
					OutcomeType: wfl.OutcomeTypeExecution,
					Outcome:     outcome3,
					Timestamp:   3,
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
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     compOutcome2,
					Timestamp:   4,
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
					Timestamp: 5,
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
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
					wfl.Step{Execute: executeAction2, Compensate: compensateAction2},
					wfl.Step{Execute: executeAction3},
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
			outcome1.Mock, outcome2.Mock, compOutcome1.Mock, compOutcome2.Mock,
			timeProvider.Mock, emitter.Mock,
		},
	}
}

func CompStartTestCase(t *testing.T) TestCase {
	name := "Should start compensation"
	var (
		outcome1 = mock.NewOutcomeMock().RegisterNFailure(1,
			func() bool { return false },
		)
		outcome2 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return true },
		)

		compOutcome1 = mock.NewOutcomeMock().RegisterNFailure(2,
			func() bool { return false },
		)

		executeAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return outcome1, nil
			},
		)
		compensateAction1 = mock.NewActionMock[wfl.Outcome]().RegisterRun(
			func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (wfl.Outcome, error) {
				return compOutcome1, nil
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

		i            int64 = 0
		timeProvider       = mock.NewTimeProviderMock().RegisterNNow(2,
			func() int64 {
				i++
				return i
			},
		)

		emitter = mock.NewEmitterMock().RegisterEmit(
			func(ctx context.Context, event evnt.Event) error {
				want := evnt.StepOutcomeEvent{
					NodeID:      "node_id",
					Key:         wfl.ID("wfl-id"),
					Type:        wfl.Type("my-wfl"),
					OutcomeSeq:  0,
					OutcomeType: wfl.OutcomeTypeCompensation,
					Outcome:     compOutcome1,
					Timestamp:   1,
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
					Timestamp: 2,
				}
				asserterror.EqualDeep(t, event, want)
				return nil
			},
		)
	)
	progress := wfl.NewProgress("wfl-1", "my-wfl")
	progress.RecordExecutionOutcome(0, outcome1)
	progress.RecordExecutionOutcome(1, outcome2)

	return TestCase{
		Name: name,
		Setup: Setup{
			NodeID: "node_id",

			Spec: wfl.Spec{
				ID: wfl.ID("wfl-id"),
				Def: wfl.MustDefinition(wfl.Type("my-wfl"), resultBuilder,
					wfl.Step{Execute: executeAction1, Compensate: compensateAction1},
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
			outcome1.Mock, outcome2.Mock, compOutcome1.Mock,
			timeProvider.Mock, emitter.Mock,
		},
	}
}
