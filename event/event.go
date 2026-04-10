package event

import wfl "github.com/kapbit/kapbit-go/workflow"

type Event any

type SystemEvent interface {
	Partition() int
	isSystem()
}

type WorkflowEvent interface {
	WorkflowID() wfl.ID
	WorkflowType() wfl.Type
	IsTerminal() bool
}
