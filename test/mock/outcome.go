package mock

import (
	"github.com/ymz-ncnk/mok"
)

// OutcomeMock is a mock for the Outcome interface.
type OutcomeMock struct {
	*mok.Mock
}

// NewOutcomeMock returns a new OutcomeMock instance.
func NewOutcomeMock() OutcomeMock {
	return OutcomeMock{mok.New("Outcome")}
}

// --- Function type for registering mock behavior ---

type (
	FailureFn func() bool
)

// --- Register methods ---

func (m OutcomeMock) RegisterFailure(fn FailureFn) OutcomeMock {
	m.Register("Failure", fn)
	return m
}

func (m OutcomeMock) RegisterNFailure(n int, fn FailureFn) OutcomeMock {
	m.RegisterN("Failure", n, fn)
	return m
}

// --- Unregister methods ---

func (m OutcomeMock) UnregisterFailure() OutcomeMock {
	m.Unregister("Failure")
	return m
}

// --- Interface methods ---

func (m OutcomeMock) Failure() bool {
	results, err := m.Call("Failure")
	if err != nil {
		panic(err)
	}
	v, _ := results[0].(bool)
	return v
}
