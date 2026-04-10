package retry

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	supretry "github.com/kapbit/kapbit-go/support/retry"
	expnt "github.com/kapbit/kapbit-go/support/retry/exponential"
)

const (
	// maxRetryLevel prevents the retry index from growing indefinitely.
	// i=60 is used as a safety cap before passing to the exponential backoff runner.
	maxRetryLevel = 60

	// warnRetryLevel defines when a transient failure is considered "notable"
	// enough to warrant a log message.
	warnRetryLevel = 5
)

type RetryEmitter struct {
	handler          EventHandler
	runner           supretry.Runner
	logger           *slog.Logger
	degraded         *atomic.Bool
	onTransientError OnTransientError
}

func NewRetryEmitter(handler EventHandler, opts ...SetOption) (emitter RetryEmitter,
	err error,
) {
	o := DefaultOptions()
	if err = Apply(&o, opts...); err != nil {
		return
	}
	emitter = RetryEmitter{
		handler:          handler,
		runner:           supretry.NewRunner(o.Policy),
		logger:           o.Logger.With("comp", "event_emitter"),
		degraded:         &atomic.Bool{},
		onTransientError: o.OnTransientError,
	}
	return
}

func (e RetryEmitter) Emit(ctx context.Context, event evnt.Event) (err error) {
	var i int
	task := func(ctx context.Context) (level int, stop bool, err error) {
		i = e.nextLevel(i)
		err = e.handler.Handle(ctx, event)
		if err != nil {
			// Decide if we should stop retrying.
			if e.shouldStopTrying(err) {
				return expnt.UndefinedLevel, true, err
			}
			// Transient error
			e.reportTransientError(i, err)
			if e.onTransientError != nil {
				e.onTransientError(err)
			}
			return i, false, err
		}
		e.reportRecovery(i)
		return expnt.UndefinedLevel, true, nil
	}
	return e.runner.Run(ctx, task)
}

func (e RetryEmitter) shouldStopTrying(err error) bool {
	var (
		eerr *codes.EncodeError
		perr *codes.PersistenceError
	)
	return errors.As(err, &eerr) ||
		(errors.As(err, &perr) &&
			(perr.Kind() == codes.PersistenceKindFenced || perr.Kind() == codes.PersistenceKindRejection))
}

func (e RetryEmitter) nextLevel(current int) int {
	if current < maxRetryLevel {
		return current + 1
	}
	return current
}

func (e RetryEmitter) reportTransientError(i int, err error) {
	if i >= warnRetryLevel && e.degraded.CompareAndSwap(false, true) {
		e.logger.Warn("events emit is struggling; retries are accumulating",
			"err", err)
	}
}

func (e RetryEmitter) reportRecovery(i int) {
	if i >= warnRetryLevel && e.degraded.CompareAndSwap(true, false) {
		e.logger.Info("events emit recovered")
	}
}
