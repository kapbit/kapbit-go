package mock

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/ymz-ncnk/mok"
)

// EventHandlerMock is a mock implementation of the retry.EventHandler interface.
type EventHandlerMock struct {
	*mok.Mock
}

// NewEventHandlerMock returns a new instance of EventHandlerMock.
func NewEventHandlerMock() EventHandlerMock {
	return EventHandlerMock{mok.New("EventHandler")}
}

// --- Function type for registering mock behavior ---

// EventHandlerHandleFn defines the function signature for mocking the Handle() method.
type EventHandlerHandleFn func(ctx context.Context, event evnt.Event) error

// --- Register methods ---

// RegisterHandle registers a single mock function for the Handle() method.
func (m EventHandlerMock) RegisterHandle(fn EventHandlerHandleFn) EventHandlerMock {
	m.Register("Handle", fn)
	return m
}

// RegisterNHandle registers a mock function for the next n calls to Handle().
func (m EventHandlerMock) RegisterNHandle(n int, fn EventHandlerHandleFn) EventHandlerMock {
	m.RegisterN("Handle", n, fn)
	return m
}

// UnregisterHandle unregisters all mock functions for Handle().
func (m EventHandlerMock) UnregisterHandle() EventHandlerMock {
	m.Unregister("Handle")
	return m
}

// --- Interface method implementation ---

// Handle delegates to the registered mock function.
func (m EventHandlerMock) Handle(ctx context.Context, event evnt.Event) error {
	results, err := m.Call("Handle", ctx, event)
	if err != nil {
		panic(err)
	}
	e, _ := results[0].(error)
	return e
}
