package mock

import (
	"time"

	"github.com/ymz-ncnk/mok"
)

// PolicyMock is a mock implementation of the retry.Policy interface.
type PolicyMock struct {
	*mok.Mock
}

// NewPolicyMock returns a new instance of PolicyMock.
func NewPolicyMock() PolicyMock {
	return PolicyMock{mok.New("Policy")}
}

// --- Function type for registering mock behavior ---

// PolicyNextDelayFn defines the function signature for mocking the NextDelay() method.
type PolicyNextDelayFn func(state int) time.Duration

// --- Register methods ---

// RegisterNextDelay registers a single mock function for the NextDelay() method.
func (m PolicyMock) RegisterNextDelay(fn PolicyNextDelayFn) PolicyMock {
	m.Register("NextDelay", fn)
	return m
}

// RegisterNNextDelay registers a mock function for the next n calls to NextDelay().
func (m PolicyMock) RegisterNNextDelay(n int, fn PolicyNextDelayFn) PolicyMock {
	m.RegisterN("NextDelay", n, fn)
	return m
}

// UnregisterNextDelay unregisters all mock functions for NextDelay().
func (m PolicyMock) UnregisterNextDelay() PolicyMock {
	m.Unregister("NextDelay")
	return m
}

// --- Interface method implementation ---

// NextDelay delegates to the registered mock function.
func (m PolicyMock) NextDelay(state int) time.Duration {
	results, err := m.Call("NextDelay", state)
	if err != nil {
		panic(err)
	}
	d, _ := results[0].(time.Duration)
	return d
}
