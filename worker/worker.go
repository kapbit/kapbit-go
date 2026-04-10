package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/kapbit/kapbit-go/codes"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	supretry "github.com/kapbit/kapbit-go/support/retry"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type Engine interface {
	RetryQueue() <-chan *rtm.WorkflowRef
	EmitWorkflowCreated(ctx context.Context, ref *rtm.WorkflowRef) error
	ExecWorkflow(ctx context.Context, ref *rtm.WorkflowRef) (wfl.Result, error)
	EmitDeadLetter(ctx context.Context, ref *rtm.WorkflowRef) error
}

type Worker struct {
	options Options
	runner  supretry.Runner
	level   atomic.Int32
	engine  Engine
	gate    *support.IngressGate
	wg      sync.WaitGroup
	logger  *slog.Logger
}

func New(engine Engine, gate *support.IngressGate,
	opts ...SetOption,
) (worker *Worker, err error) {
	o := Options{}
	Apply(&o, opts...)
	o.SetDefaults()
	if err = o.Validate(); err != nil {
		return
	}
	worker = &Worker{
		options: o,
		runner:  supretry.NewRunner(o.Policy),
		engine:  engine,
		gate:    gate,
		logger:  o.Logger.With("comp", "retry_worker"),
	}
	worker.level.Store(int32(swtch.NormalLevel))
	return
}

func (w *Worker) Start(ctx context.Context) {
	w.reportStarted()
	w.runner.Run(ctx, w.process)
	w.wg.Wait()
}

func (w *Worker) Level() int {
	return int(w.level.Load())
}

// The task logic is agnostic to the policy type
func (w *Worker) process(ctx context.Context) (level int, stop bool, err error) {
	select {

	case <-ctx.Done():
		w.reportStopped()
		return swtch.UndefinedLevel, true, context.Canceled

	case ref, ok := <-w.engine.RetryQueue():
		if !ok {
			w.reportStopped()
			return swtch.UndefinedLevel, true, errors.New("retry queue closed")
		}
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			var err error
			if ref.IncrementRetriesCount(w.options.MaxAttempts) {
				err = w.processRef(ctx, ref)
			} else {
				err = w.emitDeadLetter(ctx, ref)
			}
			w.handleError(err, ref)
		}()
		return w.Level(), false, nil
	}
}

func (w *Worker) processRef(ctx context.Context, ref *rtm.WorkflowRef) (err error) {
	if ref.Saved() {
		err = w.execWorkflow(ctx, ref)
	} else {
		err = w.saveAndExecWorkflow(ctx, ref)
	}
	return
}

func (w *Worker) saveAndExecWorkflow(ctx context.Context,
	ref *rtm.WorkflowRef,
) (err error) {
	w.reportEmittingWorkflowCreated(ref)
	err = w.engine.EmitWorkflowCreated(ctx, ref)
	if err != nil {
		return
	}
	ref.MarkSaved()
	return w.execWorkflow(ctx, ref)
}

func (w *Worker) execWorkflow(ctx context.Context, ref *rtm.WorkflowRef) (err error) {
	w.reportExecutingWorkflow(ref)
	_, err = w.engine.ExecWorkflow(ctx, ref)
	return
}

// emitDeadLetter emits a dead letter event.
//
// Could return ErrCircuitBreakerOpen.
func (w *Worker) emitDeadLetter(ctx context.Context, ref *rtm.WorkflowRef) (err error) {
	w.reportEmittingDeadLetter(ref)
	err = w.engine.EmitDeadLetter(ctx, ref)
	return
}

func (w *Worker) handleError(err error, ref *rtm.WorkflowRef) {
	if err != nil {
		w.reportWorkflowExecutionFailed(ref, err)
	}

	var cberr *codes.CircuitBreakerOpenError
	if err == nil || !errors.As(err, &cberr) {
		// In a case of a non-circuit breaker error, we opens the gate and
		// resume normal execution.
		if w.gate.Open() {
			w.reportIngressGateOpened(ref)
			w.runNormal()
		}
		return
	}
	w.slowDown()
}

func (w *Worker) slowDown() {
	w.level.Store(int32(swtch.SlowLevel))
}

func (w *Worker) runNormal() {
	w.level.Store(int32(swtch.NormalLevel))
}

func (w *Worker) reportStarted() {
	w.logger.Info("worker started", "max_attempts", w.options.MaxAttempts)
}

func (w *Worker) reportStopped() {
	w.logger.Info("worker stopped")
}

func (w *Worker) reportEmittingWorkflowCreated(ref *rtm.WorkflowRef) {
	spec := ref.Workflow().Spec()
	w.logger.Debug("workflow created event emitting",
		"wid", spec.ID,
		"wtype", spec.Def.Type(),
		"attempt", ref.RetryCount(),
	)
}

func (w *Worker) reportExecutingWorkflow(ref *rtm.WorkflowRef) {
	spec := ref.Workflow().Spec()
	w.logger.Debug("workflow executing",
		"wid", spec.ID,
		"wtype", spec.Def.Type(),
		"attempt", ref.RetryCount(),
	)
}

func (w *Worker) reportEmittingDeadLetter(ref *rtm.WorkflowRef) {
	spec := ref.Workflow().Spec()
	w.logger.Debug("max attempts reached, dead letter event emitting",
		"wid", spec.ID,
		"wtype", spec.Def.Type(),
		"attempt", ref.RetryCount(),
	)
}

func (w *Worker) reportWorkflowExecutionFailed(ref *rtm.WorkflowRef, err error) {
	spec := ref.Workflow().Spec()
	w.logger.Debug("workflow execution failed",
		"wid", spec.ID,
		"wtype", spec.Def.Type(),
		"attempt", ref.RetryCount(),
		"err", err,
	)
}

func (w *Worker) reportIngressGateOpened(ref *rtm.WorkflowRef) {
	spec := ref.Workflow().Spec()
	w.logger.Info("ingress gate opened, resuming normal execution",
		"wid", spec.ID,
		"wtype", spec.Def.Type(),
		"attempt", ref.RetryCount(),
	)
}
