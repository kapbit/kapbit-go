package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
)

// WorkerMock is a mock implementation of the worker.Worker interface.
type WorkerMock struct {
	*mok.Mock
}

// NewWorkerMock returns a new instance of WorkerMock.
func NewWorkerMock() WorkerMock {
	return WorkerMock{mok.New("Worker")}
}

// --- Function type definitions for mocking ---

type WorkerStartFn func(ctx context.Context)

// --- Register methods ---

func (m WorkerMock) RegisterStart(fn WorkerStartFn) WorkerMock {
	m.Register("Start", fn)
	return m
}

// --- Interface method implementations ---

func (m WorkerMock) Start(ctx context.Context) {
	_, err := m.Call("Start", ctx)
	if err != nil {
		panic(err)
	}
}
