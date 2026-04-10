package restorer_test

import (
	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
	wflimpl "github.com/kapbit/kapbit-go/workflow/impl"
)

var (
	// wfl-1
	workflow1 = newWorkflow(w1, defs[0], func(p *wfl.Progress) {
		p.RecordExecutionOutcome(wfl.OutcomeSeq(0), e11.Outcome)
		p.RecordExecutionOutcome(wfl.OutcomeSeq(1), e12.Outcome)
	})

	// wfl-2
	workflow2 = newWorkflow(w2, defs[1], func(p *wfl.Progress) {
		p.RecordExecutionOutcome(wfl.OutcomeSeq(0), e21.Outcome)
		p.SetResult(func() (wfl.Result, error) { return r2, nil })
	})

	// wfl-3
	workflow3 = newWorkflow(w3, defs[0], func(p *wfl.Progress) {
		p.RecordExecutionOutcome(wfl.OutcomeSeq(0), e31.Outcome)
		p.RecordExecutionOutcome(wfl.OutcomeSeq(1), e32.Outcome)
		p.RecordExecutionOutcome(wfl.OutcomeSeq(2), e33.Outcome)
		p.RecordCompensationOutcome(wfl.OutcomeSeq(0), c31.Outcome)
		p.RecordCompensationOutcome(wfl.OutcomeSeq(1), c32.Outcome)
	})

	// wfl-4
	workflow4 = newWorkflow(w4, defs[2], func(p *wfl.Progress) {
		p.RecordExecutionOutcome(wfl.OutcomeSeq(0), e41.Outcome)
		p.RecordExecutionOutcome(wfl.OutcomeSeq(1), e42.Outcome)
		p.RecordCompensationOutcome(wfl.OutcomeSeq(0), c41.Outcome)
		p.SetResult(func() (wfl.Result, error) { return r4, nil })
	})

	// wfl-5
	workflow5 = newWorkflow(w5, defs[3], func(p *wfl.Progress) {
		p.SetResult(func() (wfl.Result, error) { return r5, nil })
	})

	// wfl-6
	workflow6 = newWorkflow(w6, defs[3], nil)
)

func newWorkflow(event evnt.WorkflowCreatedEvent, definition wfl.Definition,
	setupProgress func(p *wfl.Progress),
) wfl.Workflow {
	spec := wfl.Spec{
		ID:    event.WorkflowID(),
		Def:   definition,
		Input: event.Input,
	}
	progress := wfl.NewProgress(event.WorkflowID(), event.WorkflowType())
	if setupProgress != nil {
		setupProgress(progress)
	}
	return wflimpl.NewWorkflow(NodeID, spec, progress, eventTools, logger)
}
