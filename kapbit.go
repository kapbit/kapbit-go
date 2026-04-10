// Package kapbit provides the primary entry point for the Kapbit workflow
// orchestrator.
//
// Kapbit is an event-driven, durable workflow orchestrator designed for reliable
// execution of multi-step processes. It uses event sourcing to maintain state
// and optimistic concurrency control for distributed safety.
package kapbit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/persist"
	"github.com/kapbit/kapbit-go/event/retry"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/executor"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	"github.com/kapbit/kapbit-go/support/log"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
	wflimpl "github.com/kapbit/kapbit-go/workflow/impl"
)

// Kapbit is the core orchestrator instance that orchestrates workflow lifecycles.
// It manages the coordination between event storage, the internal runtime,
// and background workers.
type Kapbit struct {
	ctx         context.Context
	cancelCause context.CancelCauseFunc
	logger      *slog.Logger
	options     Options
	tools       Tools
	wg          sync.WaitGroup
}

// New creates and initializes a new Kapbit instance.
//
// The initialization follows a strict sequence to ensure distributed safety:
//  1. Apply configuration options.
//  2. Establish leadership by claiming "Active Writer" status for all
//     event partitions.
//  3. Restore the internal runtime state by replaying historical events.
//  4. Initialize the retry worker to handle background tasks.
//
// Parameters:
//   - ctx: The parent context for Kapbit.
//   - defs: Registry of workflow types and their implementation logic.
//   - codec: The serializer used for event data.
//   - repository: The underlying event storage backend.
//   - opts: Optional configuration setters for fine-tuning.
func New(ctx context.Context, defs wfl.Definitions,
	codec evs.Codec, repository evs.Repository, opts ...SetOption,
) (kapbit *Kapbit, err error) {
	o := DefaultOptions()
	if err = Apply(&o, opts...); err != nil {
		return
	}

	o.Logger = slog.New(log.NewUniqueHandler(o.Logger.Handler())).
		With("app", "kapbit", "node-id", nodeIDLogValue(o.NodeID))
	var (
		gate           = &support.IngressGate{}
		store          = evs.New(o.NodeID, codec, repository)
		partitionCount = store.PartitionCount()
		evntTools      = &evnt.Kit{Time: o.TimeProvider}
	)
	factory := wflimpl.NewFactory(defs, evntTools, o.Logger)
	// Create a temporary emitter to signal ownership before restoration.
	// This "Active Writer" event prevents race conditions by ensuring this
	// node is the unique producer for its assigned partitions.
	//
	// We use a nil gate here to avoid premature backpressure triggers during
	// the bootstrap phase.
	emitter, err := makeTempRetryEmitter(store, partitionCount, nil, o)
	if err != nil {
		return
	}
	evntTools.Emitter = emitter
	offsets, err := claimLeadership(ctx, repository, o.NodeID, partitionCount,
		evntTools)
	if err != nil {
		return
	}

	// Restore Kapbit's runtime state from the event log.
	// We've already established ourselves as an active writer at this point.
	runtime, chptman, tracker, err := restoreRuntime(ctx, factory, store, offsets,
		o)
	if err != nil {
		return
	}

	// Replace the temporary emitter with a fully-featured one connected to the
	// checkpoint manager and position tracker.
	//
	// This emitter will automatically close the ingress gate if storage becomes
	// unavailable (codes.CircuitBreakerOpenError), providing holistic
	// backpressure across Kapbit.
	emitter, err = makeRetryEmitter(store, chptman, tracker, gate, o)
	if err != nil {
		return
	}
	evntTools.Emitter = emitter
	var (
		ownCtx, cancelCause = context.WithCancelCause(context.Background())
		onFenced            = func() {
			cancelCause(ErrFenced)
		}
		engine = executor.NewEngine(runtime, evntTools, onFenced, o.Logger)
	)
	worker, err := makeRetryWorker(engine, gate, o)
	if err != nil {
		return
	}
	tools := Tools{
		Factory: factory,
		Engine:  engine,
		Worker:  worker,
		Gate:    gate,
	}
	return NewWithTools(ownCtx, cancelCause, tools, o)
}

// NewWithTools constructs a Kapbit instance from pre-initialized tools.
// Internal use or advanced initialization only.
func NewWithTools(ctx context.Context, cancelCause context.CancelCauseFunc,
	tools Tools,
	o Options,
) (kapbit *Kapbit, err error) {
	kapbit = &Kapbit{
		ctx:         ctx,
		cancelCause: cancelCause,
		logger:      o.Logger,
		options:     o,
		tools:       tools,
	}
	kapbit.wg.Add(1)
	go func() {
		defer kapbit.wg.Done()
		tools.Worker.Start(ctx)
	}()
	kapbit.reportStarted()
	return
}

// ExecWorkflow initiates the execution of a new workflow instance.
//
// It returns the workflow's final result or an error wrapped via kapbit.NewKapbitError.
//
// Lifecycle Errors:
//   - ErrClosed: If Kapbit is currently shutting down or closed.
//   - ErrIngressGateClosed: If Kapbit is in backpressure mode (circuit
//     breaker is open for storage or a required external engine).
//
// Creation Errors:
//   - wfl.WorkflowTypeNotRegisteredError: If the requested type is unknown.
//
// Registration Errors:
//   - executor.ErrIdempotencyViolation: If a workflow with the same ID was
//     previously seen.
//   - executor.ErrNoSlotsAvailable: If Kapbit's concurrency limit is reached.
//   - Persistence errors: codes.EncodeError, codes.PersistenceError, or
//     context.Canceled.
//
// Execution Errors:
//   - executor.WillRetryLaterError: If the workflow failed but is scheduled for
//     automatic retry. This usually wraps underlying step errors or
//     codes.CircuitBreakerOpenError from external engines.
//   - Protocol or storage failures: wfl.ProtocolViolationError,
//     codes.EncodeError, or codes.PersistenceError.
//
// Resilience Note:
// Transient storage failures (codes.CircuitBreakerOpenError during emission)
// are handled internally by the RetryEmitter and backpressure mechanism.
// They are never returned directly from this method.
func (k *Kapbit) ExecWorkflow(id wfl.ID, tp wfl.Type, input any) (
	result wfl.Result, err error,
) {
	if err = context.Cause(k.ctx); err != nil {
		err = NewKapbitError(err)
		return
	}
	k.wg.Add(1)
	defer k.wg.Done()

	l := k.logger.With("wid", id, "wtype", tp)
	if k.tools.Gate.Closed() {
		k.reportWorkflowSkippedGateClosed(l)
		err = NewKapbitError(ErrIngressGateClosed)
		return
	}

	start := k.options.TimeProvider.Now()
	k.reportWorkflowStarting(l)
	params := wfl.Params{ID: id, Type: tp, Input: input}
	progress := wfl.NewProgress(id, tp)
	workflow, err := k.tools.Factory.New(k.options.NodeID, params, progress)
	if err != nil {
		k.reportWorkflowCreationFailed(l, err)
		err = NewKapbitError(err)
		return
	}

	ref, err := k.tools.Engine.RegisterWorkflow(k.ctx, workflow)
	if err != nil {
		err = k.handleError("registration", err, l)
		return
	}

	result, err = k.tools.Engine.ExecWorkflow(k.ctx, ref)
	if err != nil {
		err = k.handleError("execution", err, l)
		return
	}

	// Holistic Recovery: The gate is only opened once a full workflow execution
	// succeeds. This ensures that both the storage (via emitter) and any
	// external engines (called by the workflow) are healthy.
	if k.tools.Gate.Open() {
		k.reportIngressGateOpened(l)
	}
	execTime := k.options.TimeProvider.Now() - start
	k.reportWorkflowFinished(l, execTime)
	return
}

// NodeID returns the identifier of this Kapbit instance.
func (k *Kapbit) NodeID() string {
	return k.options.NodeID
}

// WorkflowCount returns the number of currently active workflow slots.
func (k *Kapbit) WorkflowCount() int {
	return k.tools.Engine.ActiveSlots()
}

// Shutdown gracefully shuts down the Kapbit instance.
// It seals the ingress gate, waits for all active workflows to complete or
// for the context to timeout, and then closes the instance.
func (k *Kapbit) Shutdown(ctx context.Context) error {
	k.tools.Gate.Seal()
	ticker := time.NewTicker(k.options.ShutdownPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if k.tools.Engine.ActiveSlots() == 0 {
				k.Close()
				return nil
			}
		}
	}
}

// Close terminates the Kapbit instance immediately, canceling all operations.
func (k *Kapbit) Close() {
	k.cancelCause(ErrClosed)
	k.wg.Wait()
}

func (k *Kapbit) handleError(stage string, err error, l *slog.Logger) error {
	if IsFencedError(err) {
		k.reportWorkflowStageFailed(l, stage, err)
		return NewKapbitError(ErrFenced)
	}
	var cberr *codes.CircuitBreakerOpenError
	if errors.As(err, &cberr) {
		if k.tools.Gate.Close() {
			k.reportIngressGateClosed(l, cberr.Service())
		}
	} else {
		k.reportWorkflowStageFailed(l, stage, err)
	}
	return NewKapbitError(err)
}

func (k *Kapbit) reportStarted() {
	k.logger.Info("kapbit started")
}

func (k *Kapbit) reportWorkflowSkippedGateClosed(l *slog.Logger) {
	l.Debug("workflow skipped: ingress gate closed")
}

func (k *Kapbit) reportWorkflowStarting(l *slog.Logger) {
	l.Debug("workflow starting")
}

func (k *Kapbit) reportWorkflowCreationFailed(l *slog.Logger, err error) {
	l.Error("workflow creation failed", "err", err)
}

func (k *Kapbit) reportWorkflowFinished(l *slog.Logger, execTime int64) {
	l.Debug("workflow execution finished", "exec_time", execTime)
}

func (k *Kapbit) reportIngressGateClosed(l *slog.Logger, engine string) {
	reportIngressGateClosed(l, engine)
}

func reportIngressGateClosed(l *slog.Logger, engine string) {
	l.Debug("ingress gate closed", "trigger", "circuit_breaker", "engine", engine)
}

func (k *Kapbit) reportIngressGateOpened(l *slog.Logger) {
	reportIngressGateOpened(l)
}

func reportIngressGateOpened(l *slog.Logger) {
	l.Info("ingress gate opened, resuming normal execution")
}

func (k *Kapbit) reportWorkflowStageFailed(l *slog.Logger, stage string, err error) {
	l.Debug(fmt.Sprintf("workflow %s failed", stage), "err", err)
}

// claimLeadership establishes this instance as the active writer for all
// event partitions. It emits an ActiveWriterEvent for each partition and
// waits for it to be confirmed by the storage.
func claimLeadership(ctx context.Context, repository evs.Repository,
	nodeID string, partitionCount int, evntTools *evnt.Kit,
) (offsets []int64, err error) {
	offsets = make([]int64, partitionCount)
	for partition := range partitionCount {
		action := func() error {
			return emitActiveWriterEvent(ctx, nodeID, partition, evntTools)
		}
		callback := func(ctx context.Context, record evs.EventRecord) (bool, error) {
			if record.ActiveWriterEvent() && record.Meta[evs.MetaKeyNodeID] == string(nodeID) {
				return true, nil
			}
			return false, nil
		}
		var offset int64
		offset, err = repository.WatchAfter(ctx, partition, action, callback)
		if err != nil {
			return nil, err
		}
		offsets[partition] = offset
	}
	return offsets, nil
}

// emitActiveWriterEvent sends a leadership signal to the event store.
func emitActiveWriterEvent(ctx context.Context, nodeID string, partition int,
	evntTools *evnt.Kit,
) (err error) {
	event := evnt.ActiveWriterEvent{
		NodeID:         nodeID,
		PartitionIndex: partition,
		Timestamp:      evntTools.Time.Now(),
	}
	err = evntTools.Emitter.Emit(ctx, event)
	if err != nil {
		err = fmt.Errorf("failed to emit ActiveWriterEvent, partition %d: %w", partition,
			err)
		return
	}
	return
}

// restoreRuntime recreates Kapbit's runtime state from the event log starting
// from the given partition offsets.
func restoreRuntime(ctx context.Context, factory wfl.Factory,
	store evs.EventStore, offsets []int64, o Options,
) (runtime *rtm.Runtime, chptman *chptsup.Manager,
	tracker *rtm.PositionTracker, err error,
) {
	opts := append(
		[]rtm.SetOption{rtm.WithLogger(o.Logger)}, o.Runtime...,
	)
	restorer, err := rtm.NewRestorer(o.NodeID, factory, store, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	return restorer.Restore(ctx, offsets)
}

func makeTempRetryEmitter(store evs.EventStore, partitionCount int,
	gate *support.IngressGate,
	o Options,
) (emitter evnt.Emitter, err error) {
	// Create temporary checkpoint manager and position tracker for initial
	// active writer event emission. These will be replaced after restoration.
	chpts := make([]chptsup.Tuple, partitionCount)
	for i := range partitionCount {
		chpts[i] = chptsup.Tuple{N: -1, V: -1}
	}
	var (
		chptman = chptsup.NewManager(o.Logger, chpts, rtm.DefaultCheckpointSize)
		tracker = rtm.NewPositionTracker(chpts)
	)
	return makeRetryEmitter(store, chptman, tracker, gate, o)
}

func makeRetryEmitter(store evs.EventStore, chptman *chptsup.Manager,
	tracker *rtm.PositionTracker,
	gate *support.IngressGate,
	o Options,
) (emitter evnt.Emitter, err error) {
	handler, err := persist.NewPersistHandler(store, chptman, tracker,
		makePersistHandlerOpts(o)...)
	if err != nil {
		return
	}
	opts := append(
		[]retry.SetOption{retry.WithLogger(o.Logger)}, o.Emitter...,
	)
	if gate != nil {
		onTransientError := func(err error) {
			var cberr *codes.CircuitBreakerOpenError
			if errors.As(err, &cberr) {
				if gate.Close() {
					reportIngressGateClosed(o.Logger, cberr.Service())
				}
			}
		}
		opts = append(opts, retry.WithOnTransientError(onTransientError))
	}
	return retry.NewRetryEmitter(handler, opts...)
}

func makePersistHandlerOpts(o Options) []persist.SetOption {
	return append([]persist.SetOption{persist.WithLogger(o.Logger)}, o.PersistHandler...)
}

func makeRetryWorker(engine Engine, gate *support.IngressGate,
	o Options,
) (worker *wrk.Worker, err error) {
	opts := append(
		[]wrk.SetOption{wrk.WithLogger(o.Logger)}, o.Worker...,
	)
	return wrk.New(engine, gate, opts...)
}

func nodeIDLogValue(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
