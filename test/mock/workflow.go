package mock

import (
	"context"

	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

// WorkflowMock is a mock implementation of the Workflow interface.
type WorkflowMock struct {
	*mok.Mock
}

// NewWorkflowMock returns a new instance of WorkflowMock.
func NewWorkflowMock() WorkflowMock {
	return WorkflowMock{mok.New("Workflow")}
}

// --- Function type definitions for mocking ---

type (
	WorkflowNodeIDFn       func() string
	WorkflowRunFn          func(ctx context.Context) (wfl.Result, error)
	WorkflowStateFn        func() wfl.State
	WorkflowRefreshStateFn func()
	WorkflowSpecFn         func() wfl.Spec
	WorkflowProgressFn     func() *wfl.Progress
	WorkflowEqualFn        func(other wfl.Workflow) bool
)

// --- Register methods ---

func (m WorkflowMock) RegisterNodeID(fn WorkflowNodeIDFn) WorkflowMock {
	m.Register("NodeID", fn)
	return m
}

func (m WorkflowMock) RegisterRun(fn WorkflowRunFn) WorkflowMock {
	m.Register("Run", fn)
	return m
}

func (m WorkflowMock) RegisterState(fn WorkflowStateFn) WorkflowMock {
	m.Register("State", fn)
	return m
}

func (m WorkflowMock) RegisterRefreshState(fn WorkflowRefreshStateFn) WorkflowMock {
	m.Register("RefreshState", fn)
	return m
}

func (m WorkflowMock) RegisterSpec(fn WorkflowSpecFn) WorkflowMock {
	m.Register("Spec", fn)
	return m
}

func (m WorkflowMock) RegisterNSpec(n int, fn WorkflowSpecFn) WorkflowMock {
	m.RegisterN("Spec", n, fn)
	return m
}

func (m WorkflowMock) RegisterProgress(fn WorkflowProgressFn) WorkflowMock {
	m.Register("Progress", fn)
	return m
}

func (m WorkflowMock) RegisterNProgress(n int, fn WorkflowProgressFn) WorkflowMock {
	m.RegisterN("Progress", n, fn)
	return m
}

func (m WorkflowMock) RegisterWorkflow(fn WorkflowEqualFn) WorkflowMock {
	m.Register("Equal", fn)
	return m
}

// --- Interface method implementations ---

func (m WorkflowMock) NodeID() string {
	results, err := m.Call("NodeID")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(string)
	return res0
}

func (m WorkflowMock) Run(ctx context.Context) (wfl.Result, error) {
	results, err := m.Call("Run", ctx)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(wfl.Result)
	res1, _ := results[1].(error)
	return res0, res1
}

func (m WorkflowMock) State() wfl.State {
	results, err := m.Call("State")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(wfl.State)
	return res0
}

func (m WorkflowMock) RefreshState() {
	_, err := m.Call("RefreshState")
	if err != nil {
		panic(err)
	}
}

func (m WorkflowMock) Spec() wfl.Spec {
	results, err := m.Call("Spec")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(wfl.Spec)
	return res0
}

func (m WorkflowMock) Progress() *wfl.Progress {
	results, err := m.Call("Progress")
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(*wfl.Progress)
	return res0
}

func (m WorkflowMock) Equal(other wfl.Workflow) bool {
	results, err := m.Call("Equal", other)
	if err != nil {
		panic(err)
	}
	res0, _ := results[0].(bool)
	return res0
}
