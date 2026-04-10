package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type DeadLetterEvent struct {
	NodeID         string
	Key            wfl.ID
	Type           wfl.Type
	WorkflowOffset int64
	Timestamp      int64
}

func (e *DeadLetterEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e *DeadLetterEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e *DeadLetterEvent) IsTerminal() bool {
	return true
}

func (e *DeadLetterEvent) SetWorkflowOffset(offset int64) {
	e.WorkflowOffset = offset
}
