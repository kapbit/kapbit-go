package mock

import (
	"github.com/ymz-ncnk/mok"
)

// TimeProviderMock is a mock implementation of the TimeProvider interface.
type TimeProviderMock struct {
	*mok.Mock
}

// NewTimeProviderMock returns a new instance of TimeProviderMock.
func NewTimeProviderMock() TimeProviderMock {
	return TimeProviderMock{mok.New("TimeProvider")}
}

// --- Function type for registering mock behavior ---

// TimeProviderNowFn defines the function signature for mocking the Now() method.
type TimeProviderNowFn func() int64

// --- Register methods ---

// RegisterNow registers a single mock function for the Now() method.
func (m TimeProviderMock) RegisterNow(fn TimeProviderNowFn) TimeProviderMock {
	m.Register("Now", fn)
	return m
}

// TODO Should be RegisterNowN

// RegisterNNow registers a mock function for the next n calls to Now().
func (m TimeProviderMock) RegisterNNow(n int, fn TimeProviderNowFn) TimeProviderMock {
	m.RegisterN("Now", n, fn)
	return m
}

// UnregisterNow unregisters all mock functions for Now().
func (m TimeProviderMock) UnregisterNow() TimeProviderMock {
	m.Unregister("Now")
	return m
}

// --- Interface method implementation ---

// Now delegates to the registered mock function.
func (m TimeProviderMock) Now() int64 {
	results, err := m.Call("Now")
	if err != nil {
		panic(err)
	}
	t, _ := results[0].(int64)
	return t
}
