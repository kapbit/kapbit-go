package mock

import (
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

// FactoryMock is a mock implementation of the Factory interface.
type FactoryMock struct {
	*mok.Mock
}

// NewFactoryMock returns a new instance of FactoryMock.
func NewFactoryMock() FactoryMock {
	return FactoryMock{mok.New("Factory")}
}

// --- Function type for registering mock behavior ---

// FactoryNewFn defines the function signature for mocking the New() method.
type FactoryNewFn func(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error)

// --- Register methods ---

// RegisterNew registers a single mock function for the New() method.
func (m FactoryMock) RegisterNew(fn FactoryNewFn) FactoryMock {
	m.Register("New", fn)
	return m
}

// RegisterNNew registers a mock function for the next n calls to New().
func (m FactoryMock) RegisterNNew(n int, fn FactoryNewFn) FactoryMock {
	m.RegisterN("New", n, fn)
	return m
}

// UnregisterNew unregisters all mock functions for New().
func (m FactoryMock) UnregisterNew() FactoryMock {
	m.Unregister("New")
	return m
}

// --- Interface method implementation ---

// New delegates to the registered mock function.
func (m FactoryMock) New(nodeID string, params wfl.Params, progress *wfl.Progress) (wfl.Workflow, error) {
	results, err := m.Call("New", nodeID, params, mok.SafeVal[*wfl.Progress](progress))
	if err != nil {
		panic(err)
	}
	w, _ := results[0].(wfl.Workflow)
	e, _ := results[1].(error)
	return w, e
}
