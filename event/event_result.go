package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type WorkflowResultEvent struct {
	NodeID    string
	Key       wfl.ID
	Type      wfl.Type
	Result    wfl.Result
	Timestamp int64
}

func (e WorkflowResultEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e WorkflowResultEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e WorkflowResultEvent) IsTerminal() bool {
	return true
}
