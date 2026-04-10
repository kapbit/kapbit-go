package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type StepOutcomeEvent struct {
	NodeID      string
	Key         wfl.ID
	Type        wfl.Type
	OutcomeSeq  wfl.OutcomeSeq
	OutcomeType wfl.OutcomeType
	Outcome     wfl.Outcome
	Timestamp   int64
}

func (e StepOutcomeEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e StepOutcomeEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e StepOutcomeEvent) IsTerminal() bool {
	return false
}
