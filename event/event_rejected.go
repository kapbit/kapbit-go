package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type WorkflowRejectedEvent struct {
	NodeID         string
	Key            wfl.ID
	Type           wfl.Type
	WorkflowOffset int64
	Reason         string
	Timestamp      int64
}

func (e *WorkflowRejectedEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e *WorkflowRejectedEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e *WorkflowRejectedEvent) IsTerminal() bool {
	return true
}

func (e *WorkflowRejectedEvent) SetWorkflowOffset(offset int64) {
	e.WorkflowOffset = offset
}
