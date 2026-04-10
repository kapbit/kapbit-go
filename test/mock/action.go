package mock

import (
	"context"

	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

// ActionMock is a mock implementation of the Action interface.
// It delegates all method calls to an embedded mok.Mock.
type ActionMock[T any] struct {
	*mok.Mock
}

// NewActionMock returns a new instance of ActionMock.
func NewActionMock[T any]() ActionMock[T] {
	return ActionMock[T]{mok.New("Action")}
}

// --- Function type for registering mock behavior ---

// ActionRunFn defines the function signature for mocking the Run() method.
type ActionRunFn[T any] func(ctx context.Context, input wfl.Input, progress *wfl.Progress) (T, error)

// --- Register methods ---

// RegisterRun registers a single mock function for the Run() method.
func (m ActionMock[T]) RegisterRun(fn ActionRunFn[T]) ActionMock[T] {
	m.Register("Run", fn)
	return m
}

// RegisterNRun registers a mock function for the next n calls to Run().
func (m ActionMock[T]) RegisterNRun(n int, fn ActionRunFn[T]) ActionMock[T] {
	m.RegisterN("Run", n, fn)
	return m
}

// UnregisterRun unregisters all mock functions for Run().
func (m ActionMock[T]) UnregisterRun() ActionMock[T] {
	m.Unregister("Run")
	return m
}

// --- Interface method implementation ---

// Run delegates to the registered mock function.
func (m ActionMock[T]) Run(ctx context.Context, input wfl.Input, progress *wfl.Progress) (T, error) {
	results, err := m.Call("Run", ctx, input, mok.SafeVal[*wfl.Progress](progress))
	if err != nil {
		panic(err)
	}
	out, _ := results[0].(T)
	e, _ := results[1].(error)
	return out, e
}
