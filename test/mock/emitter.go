package mock

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/ymz-ncnk/mok"
)

// EmitterMock is a mock implementation of the Emitter interface.
type EmitterMock struct {
	*mok.Mock
}

// NewEmitterMock returns a new instance of EmitterMock.
func NewEmitterMock() EmitterMock {
	return EmitterMock{mok.New("Emitter")}
}

// --- Function type for registering mock behavior ---

// EmitterEmitFn defines the function signature for mocking the Emit() method.
// Using a generic 'any' or your specific Event interface type here.
type EmitterEmitFn func(ctx context.Context, event evnt.Event) error

// --- Register methods ---

// RegisterEmit registers a single mock function for the Emit() method.
func (m EmitterMock) RegisterEmit(fn EmitterEmitFn) EmitterMock {
	m.Register("Emit", fn)
	return m
}

// RegisterNEmit registers a mock function for the next n calls to Emit().
func (m EmitterMock) RegisterNEmit(n int, fn EmitterEmitFn) EmitterMock {
	m.RegisterN("Emit", n, fn)
	return m
}

// UnregisterEmit unregisters all mock functions for Emit().
func (m EmitterMock) UnregisterEmit() EmitterMock {
	m.Unregister("Emit")
	return m
}

// --- Interface method implementation ---

// Emit delegates to the registered mock function.
func (m EmitterMock) Emit(ctx context.Context, event evnt.Event) error {
	results, err := m.Call("Emit", ctx, mok.SafeVal[any](event))
	if err != nil {
		panic(err)
	}
	e, _ := results[0].(error)
	return e
}
