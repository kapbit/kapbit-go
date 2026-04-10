package event

import "context"

type Kit struct {
	Time    TimeProvider
	Emitter Emitter
}

type TimeProvider interface {
	Now() int64
}

type Emitter interface {
	// Emit emits an event.
	//
	// Returns EncodeError, PersistenceError (Fenced, Rejection).
	Emit(ctx context.Context, event Event) error
}

type TimeProviderFn func() int64

func (f TimeProviderFn) Now() int64 { return f() }
