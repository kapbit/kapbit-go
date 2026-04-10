package executor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	rtm "github.com/kapbit/kapbit-go/runtime"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type OnFenced func()

type ErrorHandler func(ref *rtm.WorkflowRef, err error)

type Engine struct {
	logger   *slog.Logger
	runtime  *rtm.Runtime
	tools    *evnt.Kit
	onFenced OnFenced
	full     atomic.Bool
	mu       sync.Mutex
}

func NewEngine(runtime *rtm.Runtime, tools *evnt.Kit, onFenced OnFenced,
	logger *slog.Logger,
) (engine *Engine) {
	return &Engine{
		logger:   logger.With("comp", "engine"),
		runtime:  runtime,
		tools:    tools,
		onFenced: onFenced,
	}
}

// RegisterWorkflow reserves a slot for the workflow and emits a workflow
// created event.
//
// Returns the workflow ref and nil error if registration succeeds.
//
// Returns ErrIdempotencyViolation if the workflow was already handled.
//
// Returns ErrNoSlotsAvailable if the max workflow limit is reached.
//
// During persistence expects to receive context.Canceled, codes.EncodeError,
// or codes.PersistenceError (with PersistenceKindNoActiveWriter or
// PersistenceKindRejection kinds). Returns any of these errors. On
// PersistenceKindNoActiveWriter also calls onFenced callback.
//
// Panics on unexpected error types or unexpected/undefined persistence error
// kinds.
func (e *Engine) RegisterWorkflow(ctx context.Context,
	workflow wfl.Workflow,
) (ref *rtm.WorkflowRef, err error) {
	var (
		spec  = workflow.Spec()
		wid   = spec.ID
		wtype = spec.Def.Type()
	)
	err = e.reserveSlot(wid)
	if err != nil {
		e.reportSlotReservationFailed(wid, err)
		return
	}

	ref = rtm.NewWorkflowRef(workflow, 0)
	err = e.EmitWorkflowCreated(ctx, ref)
	if err != nil {
		return nil, err
	}
	ref.MarkSaved()

	e.reportWorkflowRegistered(wid, wtype)
	return
}

// ExecWorkflow executes the workflow.
//
// Returns the workflow result and nil error if execution succeeds.
//
// On wfl.ProtocolViolationError or codes.EncodeError, emits a
// workflow rejected event and returns the error (or joined error if emission
// fails).
//
// On codes.PersistenceError with PersistenceKindNoActiveWriter, calls the
// onFenced callback and returns the error.
//
// On codes.PersistenceError with PersistenceKindRejection, emits a
// workflow rejected event and returns the error (or joined error if emission
// fails).
//
// On any other error (wfl.ExecutionError, wfl.CompensationError,
// wfl.CompletionError, which may wrap codes.CircuitBreakerOpenError from
// user-defined step functions), schedules the workflow for retry and returns
// WillRetryLaterError wrapping the original error.
//
// Panics on unexpected/undefined persistence error kinds.
func (e *Engine) ExecWorkflow(ctx context.Context,
	ref *rtm.WorkflowRef,
) (result wfl.Result, err error) {
	var (
		workflow = ref.Workflow()
		spec     = workflow.Spec()
		wid      = spec.ID
		wtype    = spec.Def.Type()
	)
	e.reportExecutingWorkflow(wid, wtype)

	result, err = workflow.Run(ctx)
	if err != nil {
		err = e.onExecError(err, ref).Do(ctx, e)
		return
	}
	e.completeSlot(wid)

	e.reportWorkflowExecutionCompleted(wid, wtype)
	return
}

func (e *Engine) EmitWorkflowCreated(ctx context.Context,
	ref *rtm.WorkflowRef,
) (err error) {
	var (
		workflow = ref.Workflow()
		spec     = workflow.Spec()
		wid      = spec.ID
		event    = evnt.WorkflowCreatedEvent{
			NodeID:    ref.Workflow().NodeID(),
			Key:       wid,
			Type:      spec.Def.Type(),
			Input:     spec.Input,
			Timestamp: e.tools.Time.Now(),
		}
	)
	err = e.tools.Emitter.Emit(ctx, event)
	if err != nil {
		e.releaseSlot(spec.ID)
		return e.onWorkflowCreatedError(err, ref).Do(ctx, e)
	}

	e.reportWorkflowCreatedEventEmitted(wid, spec.Def.Type())
	return
}

func (e *Engine) EmitDeadLetter(ctx context.Context,
	ref *rtm.WorkflowRef,
) (err error) {
	var (
		spec  = ref.Workflow().Spec()
		event = &evnt.DeadLetterEvent{
			NodeID:    ref.Workflow().NodeID(),
			Key:       spec.ID,
			Type:      spec.Def.Type(),
			Timestamp: e.tools.Time.Now(),
		}
	)
	err = e.tools.Emitter.Emit(ctx, event)
	if err != nil {
		return e.onDeadLetterError(err, ref).Do(ctx, e)
	}
	e.completeSlot(spec.ID)

	e.reportDeadLetterEventEmitted(spec.ID, spec.Def.Type())
	return
}

func (e *Engine) EmitRejection(ctx context.Context, ref *rtm.WorkflowRef,
	reason string,
) (err error) {
	// TODO
	if len(reason) > 256 {
		reason = reason[:256] + "... (truncated)"
	}

	var (
		spec   = ref.Workflow().Spec()
		wid    = spec.ID
		nodeID = ref.Workflow().NodeID()
		wtype  = spec.Def.Type()
	)
	event := &evnt.WorkflowRejectedEvent{
		NodeID:    nodeID,
		Key:       wid,
		Type:      wtype,
		Reason:    reason,
		Timestamp: e.tools.Time.Now(),
	}
	err = e.tools.Emitter.Emit(ctx, event)
	if err != nil {
		return e.onRejectionError(err, ref).Do(ctx, e)
	}
	e.completeSlot(wid)

	e.reportWorkflowRejectionEventEmitted(wid, wtype, reason)
	return
}

func (e *Engine) ActiveSlots() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.runtime.Running().Count()
}

func (e *Engine) RetryQueue() <-chan *rtm.WorkflowRef {
	return e.runtime.RetryQueue()
}

func (e *Engine) reserveSlot(wid wfl.ID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.runtime.IdempotencyWindow().Contains(wid) {
		return ErrIdempotencyViolation
	}
	if !e.runtime.Running().Increment() {
		return ErrNoSlotsAvailable
	}
	e.runtime.IdempotencyWindow().Add(wid)

	e.reportSlotReserved(wid)
	return nil
}

func (e *Engine) releaseSlot(wid wfl.ID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.runtime.Running().Decrement() {
		e.reportFailedToReleaseSlot(wid)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: can't release slot, execution counter is already 0, wid=%s",
			wid))
	}
	e.runtime.IdempotencyWindow().Remove(wid)

	e.reportSlotsAvailable()
	e.reportSlotReleased(wid)
}

func (e *Engine) completeSlot(wid wfl.ID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.runtime.Running().Decrement() {
		e.reportFailedToCompleteSlot(wid)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: can't complete slot, execution counter is already 0, wid=%s",
			wid))
	}

	e.reportSlotsAvailable()
	e.reportSlotCompleted(wid)
}

func (e *Engine) onExecError(err error, ref *rtm.WorkflowRef) OnErrorAction {
	// TODO context.Deadline
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return DoNothing{err}
	}

	var (
		spec = ref.Workflow().Spec()
		wid  = spec.ID
	)
	// 1. Protocol Violation: Internal Framework Bug.
	var verr *wfl.ProtocolViolationError
	if errors.As(err, &verr) {
		e.reportProtocolViolation(wid, spec.Def.Type(), err)
		return EmitWorkflowRejection{e, ref, err, e.logger}
	}

	// 2. Encode Error: User Implementation Warning (Bad Serializer/Data).
	var eerr *codes.EncodeError
	if errors.As(err, &eerr) {
		e.reportWorkflowResultEncodingFailed(wid, err)
		return EmitWorkflowRejection{e, ref, err, e.logger}
	}

	// 3. Persistence Errors: Check contract.
	var perr *codes.PersistenceError
	if errors.As(err, &perr) {
		switch perr.Kind() {
		// Another kapbit instance is running.
		case codes.PersistenceKindFenced:
			return FenceKapbit{perr}

		// Failed to save event.
		case codes.PersistenceKindRejection:
			return EmitWorkflowRejection{e, ref, err, e.logger}

		default:
			e.reportUndefinedPersistenceErrorKind(wid, perr.Kind())
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected persistence error kind %d for workflow %s",
				perr.Kind(), wid))
		}
	}

	// 4. Other errors: Retry.
	return RetryWorkflowExecution{ref, err, e.logger}
}

func (e *Engine) onWorkflowCreatedError(err error, ref *rtm.WorkflowRef) OnErrorAction {
	if errors.Is(err, context.Canceled) {
		return DoNothing{err}
	}

	wid := ref.Workflow().Spec().ID
	// 1. Encode Error: User Implementation Warning (Bad Serializer/Data).
	var eerr *codes.EncodeError
	if errors.As(err, &eerr) {
		e.reportWorkflowInputEncodingFailed(wid, err)
		return DoNothing{err}
	}

	// 2. Persistence Errors: Check contract.
	var perr *codes.PersistenceError
	if errors.As(err, &perr) {
		switch perr.Kind() {
		// Another kapbit instance is running.
		case codes.PersistenceKindFenced:
			return FenceKapbit{err}

		// Failed to save event.
		case codes.PersistenceKindRejection:
			e.reportFailedToPersistWorkflow(err)
			return DoNothing{err}

		default:
			e.reportUndefinedPersistenceErrorKindDuringCreation(wid, perr.Kind())
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected persistence error kind %d for workflow %s",
				perr.Kind(), wid))
		}
	}

	// 3. Unhandled Error Type: Internal Framework Bug.
	e.reportUnexpectedErrorTypeDuringCreation(wid, err)
	panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected error type for workflow %s: %T (%v)",
		wid, err, err))
}

func (e *Engine) onDeadLetterError(err error, ref *rtm.WorkflowRef) OnErrorAction {
	if errors.Is(err, context.Canceled) {
		return DoNothing{err}
	}

	wid := ref.Workflow().Spec().ID
	// 1. Encode Error: Internal Framework Bug.
	var eerr *codes.EncodeError
	if errors.As(err, &eerr) {
		e.reportCannotEncodeDeadLetter(wid, err)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: cannot encode dead letter for workflow %s: %v",
			wid, err))
	}

	// 2. Persistence Errors: Check contract.
	var perr *codes.PersistenceError
	if errors.As(err, &perr) {
		switch perr.Kind() {
		// Another kapbit is running.
		case codes.PersistenceKindFenced:
			return FenceKapbit{err}

		// Failed to save event.
		case codes.PersistenceKindRejection:
			e.reportPersistenceRejectedDeadLetter(wid, perr)
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: failed to save dead letter for workflow %s: %v",
				wid, perr))

		default:
			e.reportUndefinedPersistenceErrorKindDuringDeadLetter(wid, perr.Kind())
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected persistence error kind %d for workflow %s",
				perr.Kind(), wid))
		}
	}

	// 3. Unhandled Error Type: Internal Framework Bug.
	e.reportUnexpectedErrorTypeForDeadLetter(wid, err)
	panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected error type for dead letter %s: %T (%v)", wid, err, err))
}

func (e *Engine) onRejectionError(err error, ref *rtm.WorkflowRef) OnErrorAction {
	if errors.Is(err, context.Canceled) {
		return DoNothing{err}
	}

	wid := ref.Workflow().Spec().ID
	// 1. Encode Error: Internal Framework Bug.
	var eerr *codes.EncodeError
	if errors.As(err, &eerr) {
		e.reportCannotEncodeRejection(wid, err)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: cannot encode rejection for workflow %s: %v", wid, err))
	}

	// 2. Persistence Errors: Check contract.
	var perr *codes.PersistenceError
	if errors.As(err, &perr) {
		switch perr.Kind() {
		// Another kapbit is running.
		case codes.PersistenceKindFenced:
			return FenceKapbit{err}

		// Failed to save event.
		case codes.PersistenceKindRejection:
			e.reportPersistenceRejectedRejectionEvent(wid, perr)
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: failed to save rejection for workflow %s: %v",
				wid, perr))

		default:
			e.reportUndefinedPersistenceErrorKindDuringRejection(wid, perr.Kind())
			panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected persistence error kind %d for workflow %s",
				perr.Kind(), wid))
		}
	}

	// 3. Unhandled Error Type: Internal Framework Bug.
	e.reportUnexpectedErrorTypeForRejection(wid, err)
	panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unexpected error type for rejection %s: %T (%v)",
		wid, err, err))
}

func (e *Engine) reportWorkflowRegistered(wid wfl.ID, wtype wfl.Type) {
	e.logger.Debug("workflow registered", "wid", wid, "wtype", wtype)
}

func (e *Engine) reportExecutingWorkflow(wid wfl.ID, wtype wfl.Type) {
	e.logger.Debug("workflow executing", "wid", wid, "wtype", wtype)
}

func (e *Engine) reportWorkflowExecutionCompleted(wid wfl.ID, wtype wfl.Type) {
	e.logger.Debug("workflow execution completed", "wid", wid, "wtype", wtype)
}

func (e *Engine) reportWorkflowCreatedEventEmitted(wid wfl.ID, wtype wfl.Type) {
	e.logger.Debug("workflow created event emitted", "wid", wid, "wtype", wtype)
}

func (e *Engine) reportDeadLetterEventEmitted(wid wfl.ID, wtype wfl.Type) {
	e.logger.Debug("dead letter event emitted", "wid", wid, "wtype", wtype)
}

func (e *Engine) reportWorkflowRejectionEventEmitted(wid wfl.ID, wtype wfl.Type, reason string) {
	e.logger.Debug("workflow rejection event emitted",
		"wid", wid,
		"wtype", wtype,
		"reason", reason,
	)
}

func (e *Engine) reportSlotReserved(wid wfl.ID) {
	e.logger.Debug("slot reserved", "wid", wid)
}

func (e *Engine) reportSlotReleased(wid wfl.ID) {
	e.logger.Debug("slot released", "wid", wid)
}

func (e *Engine) reportSlotCompleted(wid wfl.ID) {
	e.logger.Debug("slot completed", "wid", wid)
}

func (e *Engine) reportSlotsAvailable() {
	if e.full.CompareAndSwap(true, false) {
		e.logger.Info("slots available")
	}
}

func (e *Engine) reportSlotReservationFailed(wid wfl.ID, err error) {
	if errors.Is(err, ErrNoSlotsAvailable) {
		if e.full.CompareAndSwap(false, true) {
			e.logger.Warn("slot reservation failed: no slots available", "wid", wid)
		}
		return
	}
	e.logger.Warn("slot reservation failed", "wid", wid, "err", err)
}

func (e *Engine) reportFailedToReleaseSlot(wid wfl.ID) {
	e.logger.Error("internal bug: can't release slot, execution counter is already 0", "wid", wid)
}

func (e *Engine) reportFailedToCompleteSlot(wid wfl.ID) {
	e.logger.Error("internal bug: can't complete slot, execution counter is already 0", "wid", wid)
}

func (e *Engine) reportProtocolViolation(wid wfl.ID, wtype wfl.Type, err error) {
	e.logger.Error("internal bug: protocol violation detected",
		"wid", wid,
		"wtype", wtype,
		"err", err,
	)
}

func (e *Engine) reportWorkflowResultEncodingFailed(wid wfl.ID, err error) {
	e.logger.Warn("workflow output/result encoding failed", "wid", wid, "err", err)
}

func (e *Engine) reportUndefinedPersistenceErrorKind(wid wfl.ID, kind codes.PersistenceErrorKind) {
	e.logger.Error("internal bug: undefined persistence error kind",
		"wid", wid,
		"kind", kind,
	)
}

func (e *Engine) reportWorkflowInputEncodingFailed(wid wfl.ID, err error) {
	e.logger.Warn("workflow input encoding failed", "wid", wid, "err", err)
}

func (e *Engine) reportFailedToPersistWorkflow(err error) {
	e.logger.Error("failed to persist workflow", "err", err)
}

func (e *Engine) reportUndefinedPersistenceErrorKindDuringCreation(wid wfl.ID, kind codes.PersistenceErrorKind) {
	e.logger.Error("internal bug: undefined persistence error kind during creation",
		"wid", wid,
		"kind", kind,
	)
}

func (e *Engine) reportUnexpectedErrorTypeDuringCreation(wid wfl.ID, err error) {
	e.logger.Error("internal bug: unexpected error type during creation",
		"wid", wid,
		"type", fmt.Sprintf("%T", err),
		"err", err,
	)
}

func (e *Engine) reportCannotEncodeDeadLetter(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: cannot encode dead letter", "wid", wid, "err", err)
}

func (e *Engine) reportPersistenceRejectedDeadLetter(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: persistence rejected dead letter", "wid", wid, "err", err)
}

func (e *Engine) reportUndefinedPersistenceErrorKindDuringDeadLetter(wid wfl.ID, kind codes.PersistenceErrorKind) {
	e.logger.Error("KAPBIT INTERNAL BUG: undefined persistence error kind during dead letter",
		"wid", wid,
		"kind", kind,
	)
}

func (e *Engine) reportUnexpectedErrorTypeForDeadLetter(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: unexpected error type for dead letter",
		"wid", wid,
		"type", fmt.Sprintf("%T", err),
		"err", err,
	)
}

func (e *Engine) reportCannotEncodeRejection(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: cannot encode rejection", "wid", wid, "err", err)
}

func (e *Engine) reportPersistenceRejectedRejectionEvent(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: persistence rejected rejection event", "wid", wid, "err", err)
}

func (e *Engine) reportUndefinedPersistenceErrorKindDuringRejection(wid wfl.ID, kind codes.PersistenceErrorKind) {
	e.logger.Error("KAPBIT INTERNAL BUG: undefined persistence error kind during rejection",
		"wid", wid,
		"kind", kind,
	)
}

func (e *Engine) reportUnexpectedErrorTypeForRejection(wid wfl.ID, err error) {
	e.logger.Error("KAPBIT INTERNAL BUG: unexpected error type for rejection",
		"wid", wid,
		"type", fmt.Sprintf("%T", err),
		"err", err,
	)
}
