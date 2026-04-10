//go:generate go run gen/main.go

package runtime

import (
	"log/slog"

	"github.com/kapbit/kapbit-go/support"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type Runtime struct {
	running           *support.Counter
	idempotencyWindow *support.FIFOSet[wfl.ID]
	retryCh           chan *WorkflowRef
	logger            *slog.Logger
}

func New(logger *slog.Logger,
	counter *support.Counter, window *support.FIFOSet[wfl.ID],
	retryCh chan *WorkflowRef,
) *Runtime {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runtime{
		running:           counter,
		idempotencyWindow: window,
		retryCh:           retryCh,
		logger:            logger.With("comp", "runtime"),
	}
}

func (r *Runtime) ScheduleRetry(ref *WorkflowRef) {
	select {
	case r.retryCh <- ref:
		r.logger.Debug("scheduled workflow for retry",
			"wid", ref.Workflow().Progress().WorkflowID(),
			"seq", ref.RetryCount(),
		)
	default:
		err := NewRetryQueueFullError(len(r.retryCh))
		r.logger.Error("failed to schedule retry: queue is full",
			"err", err,
			"wid", ref.Workflow().Progress().WorkflowID(),
		)
		panic(err)
	}
}

func (r *Runtime) Running() *support.Counter {
	return r.running
}

func (r *Runtime) IdempotencyWindow() *support.FIFOSet[wfl.ID] {
	return r.idempotencyWindow
}

func (r *Runtime) RetryQueue() <-chan *WorkflowRef {
	return r.retryCh
}
