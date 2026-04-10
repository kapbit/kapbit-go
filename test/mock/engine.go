package mock

import (
	"context"

	rtm "github.com/kapbit/kapbit-go/runtime"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

// EngineMock is a mock implementation of the executor service.
type EngineMock struct {
	*mok.Mock
}

// NewEngineMock returns a new instance of EngineMock.
func NewEngineMock() EngineMock {
	return EngineMock{mok.New("Engine")}
}

// --- Function type definitions for mocking ---

type (
	EngineRegisterWorkflowFn func(ctx context.Context,
		workflow wfl.Workflow) (ref *rtm.WorkflowRef, err error)
	EngineExecWorkflowFn func(ctx context.Context,
		ref *rtm.WorkflowRef) (result wfl.Result, err error)
	EngineEmitWorkflowCreatedFn func(ctx context.Context,
		ref *rtm.WorkflowRef) (err error)
	EngineEmitDeadLetterFn func(ctx context.Context,
		ref *rtm.WorkflowRef) error
	EngineEmitRejectionFn func(ctx context.Context,
		ref *rtm.WorkflowRef, reason string) error
	EngineActiveSlotsFn func() int
	EngineRetryQueueFn  func() <-chan *rtm.WorkflowRef
)

// --- Register methods ---

func (m EngineMock) RegisterRegisterWorkflow(fn EngineRegisterWorkflowFn) EngineMock {
	m.Register("RegisterWorkflow", fn)
	return m
}

func (m EngineMock) RegisterExecWorkflow(fn EngineExecWorkflowFn) EngineMock {
	m.Register("ExecWorkflow", fn)
	return m
}

func (m EngineMock) RegisterEmitWorkflowCreated(fn EngineEmitWorkflowCreatedFn) EngineMock {
	m.Register("EmitWorkflowCreated", fn)
	return m
}

func (m EngineMock) RegisterEmitDeadLetter(fn EngineEmitDeadLetterFn) EngineMock {
	m.Register("EmitDeadLetter", fn)
	return m
}

func (m EngineMock) RegisterEmitRejection(fn EngineEmitRejectionFn) EngineMock {
	m.Register("EmitRejection", fn)
	return m
}

func (m EngineMock) RegisterActiveSlots(fn EngineActiveSlotsFn) EngineMock {
	m.Register("ActiveSlots", fn)
	return m
}

func (m EngineMock) RegisterRetryQueue(fn EngineRetryQueueFn) EngineMock {
	m.Register("RetryQueue", fn)
	return m
}

func (m EngineMock) RegisterNRetryQueue(n int, fn EngineRetryQueueFn) EngineMock {
	m.RegisterN("RetryQueue", n, fn)
	return m
}

// --- Interface method implementations ---

func (m EngineMock) RegisterWorkflow(ctx context.Context,
	workflow wfl.Workflow,
) (ref *rtm.WorkflowRef, err error) {
	results, err := m.Call("RegisterWorkflow", ctx, workflow)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(*rtm.WorkflowRef)
	res1, _ := results[1].(error)
	return res0, res1
}

func (m EngineMock) ExecWorkflow(ctx context.Context,
	ref *rtm.WorkflowRef,
) (result wfl.Result, err error) {
	results, err := m.Call("ExecWorkflow", ctx, ref)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(wfl.Result)
	res1, _ := results[1].(error)
	return res0, res1
}

func (m EngineMock) EmitWorkflowCreated(ctx context.Context,
	ref *rtm.WorkflowRef,
) (err error) {
	results, err := m.Call("EmitWorkflowCreated", ctx, ref)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(error)
	return res0
}

func (m EngineMock) EmitDeadLetter(ctx context.Context,
	ref *rtm.WorkflowRef,
) error {
	results, err := m.Call("EmitDeadLetter", ctx, ref)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(error)
	return res0
}

func (m EngineMock) EmitRejection(ctx context.Context,
	ref *rtm.WorkflowRef, reason string,
) error {
	results, err := m.Call("EmitRejection", ctx, ref, reason)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(error)
	return res0
}

func (m EngineMock) ActiveSlots() int {
	results, err := m.Call("ActiveSlots")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(int)
	return res0
}

func (m EngineMock) RetryQueue() <-chan *rtm.WorkflowRef {
	results, err := m.Call("RetryQueue")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(<-chan *rtm.WorkflowRef)
	return res0
}
