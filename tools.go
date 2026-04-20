package kapbit

import (
	"context"

	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type Tools struct {
	Factory wfl.Factory
	Engine  Engine
	Worker  Worker
	Gate    *support.EntryGate
}

type Engine interface {
	RegisterWorkflow(ctx context.Context, workflow wfl.Workflow) (
		ref *rtm.WorkflowRef, err error)
	ExecWorkflow(ctx context.Context, ref *rtm.WorkflowRef) (result wfl.Result,
		err error)
	ActiveSlots() int

	EmitWorkflowCreated(ctx context.Context, ref *rtm.WorkflowRef) error
	EmitDeadLetter(ctx context.Context, ref *rtm.WorkflowRef) error
	RetryQueue() <-chan *rtm.WorkflowRef
}

// Worker executes pending workflows periodically.
type Worker interface {
	Start(ctx context.Context)
}
