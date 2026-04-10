package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type WorkflowCreatedEvent struct {
	NodeID    string
	Key       wfl.ID
	Type      wfl.Type
	Input     wfl.Input
	Timestamp int64
}

func (e WorkflowCreatedEvent) WorkflowID() wfl.ID {
	return e.Key
}

func (e WorkflowCreatedEvent) WorkflowType() wfl.Type {
	return e.Type
}

func (e WorkflowCreatedEvent) IsTerminal() bool {
	return false
}
