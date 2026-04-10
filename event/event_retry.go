package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type WorkflowRetryEvent struct {
	NodeID    string
	Key       wfl.ID
	Type      wfl.Type
	Timestamp int64
}

func (e WorkflowRetryEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e WorkflowRetryEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e WorkflowRetryEvent) IsTerminal() bool {
	return false
}
