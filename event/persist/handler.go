package persist

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/circbrk-go"
)

const EventStoreServiceName = "event-store"

type PersistHandler struct {
	logger    *slog.Logger
	store     evs.EventStore
	chptman   *chptsup.Manager
	resolver  support.PartitionResolver
	tracker   *runtime.PositionTracker
	states    []*partitionState
	handlerID string
	mu        sync.Mutex
}

func NewPersistHandler(store evs.EventStore, chptman *chptsup.Manager,
	tracker *runtime.PositionTracker,
	opts ...SetOption,
) (*PersistHandler, error) {
	o := DefaultOptions()
	if err := Apply(&o, opts...); err != nil {
		return nil, err
	}

	h := &PersistHandler{
		logger:    o.Logger.With("comp", "persist_handler"),
		store:     store,
		chptman:   chptman,
		tracker:   tracker,
		handlerID: uuid.New().String(),
	}

	var (
		l      = store.PartitionCount()
		states = make([]*partitionState, l)
		cbs    = make([]*circbrk.CircuitBreaker, l)
	)
	for i := range l {
		var (
			partition = i
			cb        *circbrk.CircuitBreaker
		)
		cb = circbrk.New(append(o.CircuitBreaker,
			circbrk.WithChangeStateCallback(func(state circbrk.State) {
				h.reportCircuitBreakerStateChange(partition, state)
			}),
		)...)
		states[i] = &partitionState{
			cb: cb,
			mu: sync.Mutex{},
		}
		cbs[i] = cb
	}
	resolver, err := support.NewPartitionResolver(EventStoreServiceName, cbs)
	if err != nil {
		panic(err)
	}
	h.states = states
	h.resolver = resolver
	return h, nil
}

func (h *PersistHandler) Handle(ctx context.Context, event evnt.Event) (err error) {
	switch e := event.(type) {

	case evnt.SystemEvent:
		return h.handleSystemEvent(ctx, e)

	case evnt.WorkflowEvent:
		return h.handleWorkflowEvent(ctx, e)

	default:
		h.reportUnhandledEventType(event)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: unhandled event type: %T", event))
	}
}

func (h *PersistHandler) handleSystemEvent(ctx context.Context,
	event evnt.SystemEvent,
) (err error) {
	var (
		partition = event.Partition()
		state     = h.states[partition]
	)
	if !state.cb.Allow() {
		return codes.NewCircuitBreakerOpenError("storage")
	}
	offset, err := h.store.SaveEvent(ctx, partition, event)
	if err != nil {
		state.cb.Fail()
		return
	}
	state.cb.Success()
	h.reportEventSaved(partition, event, offset)
	return
}

func (h *PersistHandler) handleWorkflowEvent(ctx context.Context,
	event evnt.WorkflowEvent,
) (err error) {
	wid := event.WorkflowID()
	if _, ok := event.(evnt.WorkflowCreatedEvent); ok {
		return h.saveWorkflowCreatedEvent(ctx, wid, event)
	}
	return h.saveNotWorkflowCreatedEvent(ctx, wid, event)
}

func (h *PersistHandler) saveWorkflowCreatedEvent(ctx context.Context,
	wid wfl.ID, event evnt.WorkflowEvent,
) (err error) {
	partition, err := h.resolver.Resolve(wid)
	if err != nil {
		// CircuitBreakerError - when all the partitions blocked.
		return NewCantResolvePartitionError(err)
	}
	return h.saveAndFillPosition(ctx, partition, event)
}

func (h *PersistHandler) saveNotWorkflowCreatedEvent(ctx context.Context,
	wid wfl.ID, event evnt.WorkflowEvent,
) (err error) {
	position, pst := h.tracker.Position(wid)
	if !pst {
		h.reportWorkflowPositionMissingForEvent(event)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: workflow %s position missing", wid))
	}
	state := h.states[position.Partition]
	if !state.cb.Allow() {
		return codes.NewCircuitBreakerOpenError("storage")
	}
	h.enrich(event, position)
	offset, err := h.store.SaveEvent(ctx, position.Partition, event)
	if err != nil {
		state.cb.Fail()
		return
	}
	state.cb.Success()
	h.reportEventSaved(position.Partition, event, offset)
	if event.IsTerminal() {
		h.saveCheckpointIfAny(ctx, wid)
		h.untrackPosition(wid)
	}
	return
}

func (h *PersistHandler) saveAndFillPosition(ctx context.Context, partition int,
	event evnt.WorkflowEvent,
) (err error) {
	// We MUST lock to ensure the physical log order matches the logical
	// sequence reservation.
	state := h.states[partition]
	state.mu.Lock()
	defer state.mu.Unlock()

	offset, err := h.store.SaveEvent(ctx, partition, event)
	if err != nil {
		state.cb.Fail()
		return
	}
	state.cb.Success()
	h.reportEventSaved(partition, event, offset)
	h.tracker.ReserveNext(event.WorkflowID(), partition, offset)
	return
}

func (h *PersistHandler) enrich(event evnt.WorkflowEvent, position runtime.Position) {
	if e, ok := event.(OffsetEnricher); ok {
		e.SetWorkflowOffset(position.PhysicalOffset)
	}
}

func (h *PersistHandler) saveCheckpointIfAny(ctx context.Context, wid wfl.ID) {
	position, pst := h.tracker.Position(wid)
	if !pst {
		h.reportWorkflowPositionMissingForCheckpoint(wid)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: workflow %s position missing", wid))
	}

	state := h.states[position.Partition]
	state.mu.Lock()
	defer state.mu.Unlock()

	chpt, err := h.chptman.Add(position.Partition, chptsup.Tuple{
		N: position.LogicalSeq,
		V: position.PhysicalOffset,
	})
	if err != nil {
		h.reportCheckpointManagerRejectedPosition(wid, position, err)
		panic(fmt.Sprintf("KAPBIT INTERNAL BUG: checkpoint manager rejected position %d: %v",
			position.Partition, err))
	}
	if chpt.Ok {
		err := h.store.SaveCheckpoint(ctx, position.Partition, chpt.Value.V)
		if err != nil {
			h.reportFailedToPersistCheckpoint(wid, chpt.Value, err)
			return
		}
		h.reportCheckpointSaved(position.Partition, chpt.Value.V)
	}
}

func (h *PersistHandler) untrackPosition(wid wfl.ID) {
	h.tracker.Untrack(wid)
}

func (h *PersistHandler) reportUnhandledEventType(event evnt.Event) {
	h.logger.Error("KAPBIT INTERNAL BUG: unhandled event type",
		"event_type", fmt.Sprintf("%T", event),
	)
}

func (h *PersistHandler) reportWorkflowPositionMissingForEvent(event evnt.WorkflowEvent) {
	h.logger.Error("KAPBIT INTERNAL BUG: workflow position missing",
		"wid", event.WorkflowID(),
		"event_type", fmt.Sprintf("%T", event),
		"details", "Can't save event.",
	)
}

func (h *PersistHandler) reportWorkflowPositionMissingForCheckpoint(wid wfl.ID) {
	h.logger.Error("KAPBIT INTERNAL BUG: workflow position missing",
		"wid", wid,
		"details", "Can't save checkpoint.",
	)
}

func (h *PersistHandler) reportCheckpointManagerRejectedPosition(wid wfl.ID,
	position runtime.Position, err error,
) {
	h.logger.Error("KAPBIT INTERNAL BUG: checkpoint manager rejected position",
		"wid", wid,
		"partition", position.Partition,
		"logical_seq", position.LogicalSeq,
		"physical_offset", position.PhysicalOffset,
		"err", err,
	)
}

func (h *PersistHandler) reportFailedToPersistCheckpoint(wid wfl.ID,
	chpt chptsup.Tuple, err error,
) {
	h.logger.Warn("failed to persist checkpoint",
		"checkpoint", chpt,
		"wid", wid,
		"cause", err,
	)
}

func (h *PersistHandler) reportEventSaved(partition int, event evnt.Event,
	offset int64,
) {
	h.logger.Debug("event saved",
		"partition", partition,
		"event_type", fmt.Sprintf("%T", event),
		"offset", offset,
	)
}

func (h *PersistHandler) reportCheckpointSaved(partition int, chpt int64) {
	h.logger.Debug("checkpoint saved",
		"partition", partition,
		"offset", chpt,
	)
}

func (h *PersistHandler) reportCircuitBreakerStateChange(partition int,
	state circbrk.State,
) {
	switch state {
	case circbrk.Open:
		h.logger.Warn("partition blocked: circuit breaker is open",
			"partition", partition,
			"handler-id", h.handlerID,
			"comp", "persist_handler",
		)
	case circbrk.HalfOpen:
		h.logger.Info("partition testing: circuit breaker is half-open",
			"partition", partition,
			"handler-id", h.handlerID,
			"comp", "persist_handler",
		)
	case circbrk.Closed:
		h.logger.Info("partition recovered: circuit breaker is closed",
			"partition", partition,
			"handler-id", h.handlerID,
			"comp", "persist_handler",
		)
	}
}
