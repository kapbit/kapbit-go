package impl

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

// Workflow is the internal implementation of the wfl.Workflow interface.
// It manages the execution state machine, progressing through steps and
// handling automatic rollbacks.
type Workflow struct {
	nodeID   string
	spec     wfl.Spec
	state    wfl.State
	progress *wfl.Progress
	tools    *evnt.Kit
	logger   *slog.Logger
}

// NewWorkflow creates a new workflow instance ready for execution or resumption.
//
// The engine follows an event-driven execution model:
//  1. Before executing a step, the current progress is checked.
//  2. If the step's outcome is already in storage (restored state), the
//     engine proceeds to the next step immediately.
//  3. If not, the engine executes the step and emits a persistence event
//     to record the outcome before advancing.
func NewWorkflow(nodeID string, spec wfl.Spec, progress *wfl.Progress,
	tools *evnt.Kit, logger *slog.Logger,
) wfl.Workflow {
	workflow := &Workflow{
		nodeID:   nodeID,
		spec:     spec,
		state:    calcState(spec.Def, progress),
		progress: progress,
		tools:    tools,
		logger:   logger.With("comp", "workflow", "wid", spec.ID),
	}
	return workflow
}

func (w *Workflow) NodeID() string {
	return w.nodeID
}

// Run executes the workflow state machine until it reaches a terminal state.
//
// It returns the final result on success, or an error if any part of the
// process fails.
//
// Types of Errors Returned:
//   - wfl.ExecutionError: Failure during a forward step execution.
//   - wfl.CompensationError: Failure while attempting to roll back a step.
//   - wfl.CompletionError: Failure in the final result construction logic.
//   - wfl.ProtocolViolationError: Internal inconsistency in workflow progress.
//   - Persistence/Codec errors: Failures from the underlying storage layer.
func (w *Workflow) Run(ctx context.Context) (result wfl.Result, err error) {
	for {
		switch w.state {
		case wfl.Executing:
			err = w.execute(ctx)
		case wfl.Compensating:
			err = w.compensate(ctx)
		case wfl.Finalizing:
			err = w.finalize(ctx)
			if err == nil {
				w.reportCompletedSuccessfully()
				return w.progress.Result(), nil
			}
		case wfl.Completed:
			return w.progress.Result(), nil
		default:
			panic(fmt.Sprintf("Workflow.Run: unexpected state %d", w.state))
		}
		if err != nil {
			w.reportRunFailed(w.state, err)
			return
		}
	}
}

func (w *Workflow) State() wfl.State {
	return w.state
}

func (w *Workflow) RefreshState() {
	w.state = calcState(w.spec.Def, w.progress)
}

func (w *Workflow) Spec() wfl.Spec {
	return w.spec
}

func (w *Workflow) Progress() *wfl.Progress {
	return w.progress
}

func (w *Workflow) Equal(other wfl.Workflow) bool {
	o, ok := other.(*Workflow)
	if !ok {
		return false
	}
	if w == nil || o == nil {
		return w == o
	}
	return w.nodeID == o.nodeID &&
		reflect.DeepEqual(w.spec, o.spec) &&
		reflect.DeepEqual(w.state, o.state) &&
		reflect.DeepEqual(w.progress, o.progress)
}

// execute handles the Executing state: it identifies the next step to run,
// executes its forward logic, and records the outcome.
func (w *Workflow) execute(ctx context.Context) (
	err error,
) {
	var outcome wfl.Outcome
	index, _, ok := w.progress.LastExecutionOutcome()
	if ok {
		index++
		if index >= len(w.spec.Def.Steps()) {
			w.state = wfl.Finalizing
			w.reportAllStepsExecuted()
			return
		}
	}
	w.reportExecutingStep(index)
	outcome, err = w.executeStep(ctx, index)
	if err != nil {
		err = wfl.NewExecutionError(err)
		return
	}
	seq := wfl.OutcomeSeq(w.progress.ExecutionLen())
	err = w.emitOutcome(ctx, seq, wfl.OutcomeTypeExecution, outcome)
	if err != nil {
		return
	}

	err = w.progress.RecordExecutionOutcome(seq, outcome)
	if err != nil {
		return
	}

	w.reportStepOutcome(index, seq, outcome)

	if outcome != wfl.NoOutcome && outcome.Failure() {
		w.state = wfl.Compensating
		w.reportOutcomeFailure(index)
	}
	return
}

// compensate handles the Compensating state: it rolls back steps in reverse
// order until all executed work is undone or a compensation fails.
func (w *Workflow) compensate(ctx context.Context) (
	err error,
) {
	index, _, ok := w.progress.LastCompensationOutcome()
	if !ok {
		index = w.progress.ExecutionLen() - 2
	} else {
		index = w.progress.ExecutionLen() - 2 - (index + 1)
	}
	if index < 0 {
		w.state = wfl.Finalizing
		w.reportAllStepsCompensated()
		return
	}

	w.reportExecutingStep(index)
	outcome, err := w.compensateStep(ctx, index)
	if err != nil {
		err = wfl.NewCompensationError(err)
		return
	}
	seq := wfl.OutcomeSeq(w.progress.CompensationLen())
	err = w.emitOutcome(ctx, seq, wfl.OutcomeTypeCompensation, outcome)
	if err != nil {
		return
	}

	err = w.progress.RecordCompensationOutcome(seq, outcome)
	if err != nil {
		return
	}

	w.reportStepOutcome(index, seq, outcome)

	if outcome != wfl.NoOutcome && outcome.Failure() {
		w.state = wfl.Finalizing
		w.reportCompensationStepFailed(index)
	}
	return
}

// finalize handles the Finalizing state: it constructs the final workflow
// result and emits a terminal event to close the workflow.
func (w *Workflow) finalize(ctx context.Context) (
	err error,
) {
	result, err := w.progress.SetResult(func() (wfl.Result, error) {
		return w.buildResult(ctx)
	})
	if err != nil {
		err = wfl.NewCompletionError(err)
		return
	}
	w.reportWorkflowResult(result)
	event := evnt.WorkflowResultEvent{
		NodeID:    w.nodeID,
		Key:       w.spec.ID,
		Type:      w.Spec().Def.Type(),
		Result:    result,
		Timestamp: w.tools.Time.Now(),
	}
	return w.tools.Emitter.Emit(ctx, event)
}

func (w *Workflow) emitOutcome(ctx context.Context, seq wfl.OutcomeSeq,
	tp wfl.OutcomeType, outcome wfl.Outcome,
) error {
	event := evnt.StepOutcomeEvent{
		NodeID:      w.nodeID,
		Key:         w.spec.ID,
		Type:        w.spec.Def.Type(),
		OutcomeSeq:  seq,
		OutcomeType: tp,
		Outcome:     outcome,
		Timestamp:   w.tools.Time.Now(),
	}
	return w.tools.Emitter.Emit(ctx, event)
}

func (w *Workflow) executeStep(ctx context.Context, index int) (
	outcome wfl.Outcome, err error,
) {
	step := w.spec.Def.Steps()[index]
	return step.Execute.Run(ctx, w.spec.Input, w.progress)
}

func (w *Workflow) compensateStep(ctx context.Context, index int) (
	outcome wfl.Outcome, err error,
) {
	step := w.spec.Def.Steps()[index]
	if step.Compensate == nil {
		outcome = wfl.NoOutcome
		return
	}
	return step.Compensate.Run(ctx, w.spec.Input, w.progress)
}

func (w *Workflow) buildResult(ctx context.Context) (resutl wfl.Result, err error) {
	return w.spec.Def.ResultBuilder().Run(ctx, w.spec.Input, w.progress)
}

// calcState determines the workflow's next logical state by analyzing
// its current progress history.
//
// State Determination Rules:
//  1. Completed: A result has already been successfully persisted.
//  2. Finalizing:
//     - All execution steps completed successfully.
//     - All required compensation steps finished successfully.
//     - A step failed, and its compensation also failed (frozen state).
//  3. Compensating:
//     - The last execution step failed.
//     - A compensation process is currently underway.
//  4. Executing:
//     - The workflow is starting or moving through its forward steps.
func calcState(def wfl.Definition, progress *wfl.Progress) wfl.State {
	if progress.Result() != nil {
		return wfl.Completed
	}

	if progress.CompensationLen() > 0 {
		index, outcome, ok := progress.LastCompensationOutcome()
		if !ok {
			panic(
				fmt.Sprintf("kapbit: compensation length is %d but LastCompensationOutcome "+
					"returned !ok (wid: %s)",
					progress.CompensationLen(), progress.WorkflowID()),
			)
		}
		if outcome != wfl.NoOutcome && outcome.Failure() {
			return wfl.Finalizing
		}
		// If we've compensated all executed steps (excluding the failed one).
		if index+1 == progress.ExecutionLen()-1 {
			return wfl.Finalizing
		}
		return wfl.Compensating
	}

	if progress.ExecutionLen() > 0 {
		index, outcome, ok := progress.LastExecutionOutcome()
		if !ok {
			panic(
				fmt.Sprintf("kapbit: execution length is %d but LastExecutionOutcome "+
					"returned !ok (wid: %s)",
					progress.ExecutionLen(), progress.WorkflowID()),
			)
		}
		if outcome != wfl.NoOutcome && outcome.Failure() {
			return wfl.Compensating
		}
		if index+1 == len(def.Steps()) {
			return wfl.Finalizing
		}
		return wfl.Executing
	}

	if len(def.Steps()) == 0 {
		return wfl.Finalizing
	}

	// If the workflow doesn't have any output.
	return wfl.Executing
}

func (w *Workflow) reportCompletedSuccessfully() {
	w.logger.Debug("workflow completed successfully")
}

func (w *Workflow) reportRunFailed(state wfl.State, err error) {
	w.logger.Debug("workflow run failed", "state", state, "err", err)
}

func (w *Workflow) reportExecutingStep(index int) {
	w.logger.Debug("step executing", "index", index)
}

func (w *Workflow) reportStepOutcome(index int, seq wfl.OutcomeSeq,
	outcome wfl.Outcome,
) {
	if outcome != wfl.NoOutcome {
		w.logger.Debug("step outcome", "index", index, "seq", seq,
			"failure", outcome.Failure())
	} else {
		w.logger.Debug("step outcome", "index", index, "seq", seq, "no_outcome", true)
	}
}

func (w *Workflow) reportOutcomeFailure(index int) {
	w.logger.Debug("step execution failed, transiting to compensation",
		"index", index)
}

func (w *Workflow) reportAllStepsExecuted() {
	w.logger.Debug("all steps executed, transiting to finalizing")
}

func (w *Workflow) reportAllStepsCompensated() {
	w.logger.Debug("all steps compensated, transiting to finalizing")
}

func (w *Workflow) reportCompensationStepFailed(index int) {
	w.logger.Debug("step compensation failed, transiting to finalizing",
		"index", index)
}

func (w *Workflow) reportWorkflowResult(result wfl.Result) {
	if result != nil {
		w.logger.Debug("workflow result",
			"result_type", fmt.Sprintf("%T", result),
		)
	} else {
		w.logger.Debug("workflow result",
			"nil", true,
		)
	}
}
