package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/support"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type Restorer struct {
	logger  *slog.Logger
	nodeID  string
	options RestorerOptions
	store   evs.EventStore
	factory wfl.Factory
}

func NewRestorer(nodeID string, factory wfl.Factory, store evs.EventStore,
	opts ...SetOption,
) (Restorer, error) {
	o := DefaultRestorerOptions()
	if err := ApplyRestorer(&o, opts...); err != nil {
		return Restorer{}, err
	}
	return Restorer{
		logger:  o.Logger.With("comp", "restorer"),
		options: o,
		nodeID:  nodeID,
		store:   store,
		factory: factory,
	}, nil
}

func (r Restorer) Restore(ctx context.Context, aweOffsets []int64) (rt *Runtime,
	chptman *chptsup.Manager,
	tracker *PositionTracker,
	err error,
) {
	startTime := time.Now()
	r.reportReconstructionStarting(r.options.IdempotencyDepth)

	rt, chptman, refs, tracker, err := r.load(ctx, aweOffsets)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, ref := range refs.Slice() {
		workflow := ref.Workflow()
		workflow.RefreshState()
		if workflow.Progress().Result() == nil && !ref.Rejected() && !ref.DeadLettered() {
			rt.ScheduleRetry(ref)
			rt.Running().Increment()
		}
	}
	r.reportReconstructionComplete(time.Since(startTime), chptman)
	r.reportRestorationStateSnapshot(chptman, tracker)
	return
}

func (r Restorer) load(ctx context.Context, aweOffsets []int64) (rt *Runtime,
	chptman *chptsup.Manager,
	refs *support.OrderedMap[*WorkflowRef, wfl.ID],
	tracker *PositionTracker,
	err error,
) {
	numPartitions := r.store.PartitionCount()
	rt = New(
		r.logger,
		support.NewCounter(r.options.MaxWorkflows),
		support.NewFIFOSet[wfl.ID](numPartitions*r.options.IdempotencyDepth),
		make(chan *WorkflowRef, r.options.MaxWorkflows),
	)
	var (
		chpts      = make([]chptsup.Tuple, numPartitions)
		terminated = struct{ wids []wfl.ID }{
			wids: []wfl.ID{},
		}
	)
	for partition := range numPartitions {
		chpts[partition] = chptsup.Tuple{N: -1, V: -1}
	}
	refs = support.NewOrderedMap[*WorkflowRef, wfl.ID]()
	tracker = NewPositionTracker(chpts)
	for partition := range numPartitions {
		var (
			chptFn  = r.makeCheckpointFn(partition, chpts)
			eventFn = r.makeEventFn(partition, aweOffsets[partition], rt, refs,
				&terminated, tracker)
		)
		err = r.store.LoadRecent(ctx, partition, r.options.IdempotencyDepth,
			chptFn, eventFn)
		if err != nil {
			err = NewRestoreError(NewPartitionError(partition, err))
			return
		}
	}
	chptman = chptsup.NewManager(r.logger, chpts, r.options.CheckpointSize)
	for _, wid := range terminated.wids {
		pos, pst := tracker.Position(wid)
		if !pst {
			r.reportMissingTrackingData(wid)
			panic(fmt.Sprintf("restorer: tracker knows nothing about the terminated wid %q",
				wid))
		}
		chptman.Add(pos.Partition, chptsup.Tuple{
			N: pos.LogicalSeq,
			V: pos.PhysicalOffset,
		})
		tracker.Untrack(wid)
	}
	return
}

func (r Restorer) makeCheckpointFn(partition int,
	chpts []chptsup.Tuple,
) evs.CheckpointCallback {
	return func(chpt int64) (err error) {
		if chpt < -1 {
			return NewInvalidCheckpointError(chpt)
		}
		chpts[partition] = chptsup.Tuple{N: -1, V: chpt}
		return nil
	}
}

func (r Restorer) makeEventFn(partition int, aweOffset int64, rt *Runtime,
	workflows *support.OrderedMap[*WorkflowRef, wfl.ID],
	terminated *struct{ wids []wfl.ID },
	tracker *PositionTracker,
) evs.EventCallback {
	return func(offset int64, beforeChpt bool, event evnt.Event) (err error) {
		switch e := event.(type) {

		case evnt.ActiveWriterEvent:
			if offset > aweOffset && e.NodeID != r.nodeID {
				return NewFencedError(partition, offset, e.NodeID)
			}
			return
		case evnt.WorkflowCreatedEvent:
			return r.handleCreatedEvent(rt, workflows, partition, offset, beforeChpt,
				tracker, e)
		case evnt.StepOutcomeEvent:
			return r.handleOutcomeEvent(workflows, partition, offset, e)
		case evnt.WorkflowResultEvent:
			return r.handleResultEvent(workflows, terminated, partition, offset, e)
		case evnt.WorkflowRetryEvent:
			return
		case *evnt.DeadLetterEvent:
			return r.handleDeadLetteredEvent(workflows, terminated, partition, offset, e)
		case *evnt.WorkflowRejectedEvent:
			return r.handleRejectedEvent(workflows, terminated, partition, offset, e)

		default:
			return NewUnknownEventTypeError(partition, offset, event)
		}
	}
}

// func (r Restorer) handleActiveWriterEvent(event evnt.ActiveWriterEvent) bool {
// 	return event.NodeID == r.nodeID
// }

func (r Restorer) handleCreatedEvent(rt *Runtime,
	refs *support.OrderedMap[*WorkflowRef, wfl.ID],
	partition int,
	offset int64,
	beforeChpt bool,
	tracker *PositionTracker,
	event evnt.WorkflowCreatedEvent,
) (err error) {
	rt.idempotencyWindow.Add(event.WorkflowID())
	if beforeChpt {
		r.reportIdempotencyWindowOnly(event.WorkflowID())
		return
	}
	tracker.ReserveNext(event.WorkflowID(), partition, offset)
	params := wfl.Params{
		ID:    event.WorkflowID(),
		Type:  event.WorkflowType(),
		Input: event.Input,
	}

	if _, pst := refs.Get(params.ID); pst {
		r.reportDuplicateEvent(evnt.EventTypeWorkflowCreated, params.ID, partition,
			offset)
		return
	}

	progress := wfl.NewProgress(params.ID, params.Type)
	workflow, err := r.factory.New(r.nodeID, params, progress)
	if err != nil {
		return
	}

	ref := NewWorkflowRef(workflow, 0)
	ref.MarkSaved()
	refs.Add(params.ID, ref)

	r.reportEvent(evnt.EventTypeWorkflowCreated, params.ID, partition, offset)
	return
}

func (r Restorer) handleOutcomeEvent(
	refs *support.OrderedMap[*WorkflowRef, wfl.ID],
	partition int,
	offset int64,
	event evnt.StepOutcomeEvent,
) (err error) {
	var (
		wid     = event.WorkflowID()
		seq     = event.OutcomeSeq
		outcome = event.Outcome
		ref     *WorkflowRef
	)
	ref, pst := refs.Get(wid)
	if !pst {
		r.reportSkipEvent(evnt.EventTypeStepOutcome, wid)
		return
	}

	switch event.OutcomeType {
	case wfl.OutcomeTypeExecution:
		err = ref.Workflow().Progress().RecordExecutionOutcome(seq, outcome)
		if err != nil {
			return
		}
		r.reportStepOutcomeEvent(wfl.OutcomeTypeExecution, seq, wid, partition,
			offset)

	case wfl.OutcomeTypeCompensation:
		err = ref.Workflow().Progress().RecordCompensationOutcome(seq, outcome)
		if err != nil {
			return
		}
		r.reportStepOutcomeEvent(wfl.OutcomeTypeCompensation, seq, wid, partition,
			offset)

	default:
		err = NewUnknownOutcomeTypeError(partition, offset, event.OutcomeType, wid)
	}
	return
}

// handleWorkflowResult restores the final result of the workflow.
func (r Restorer) handleResultEvent(
	workflows *support.OrderedMap[*WorkflowRef, wfl.ID],
	terminated *struct{ wids []wfl.ID },
	partition int,
	offset int64,
	event evnt.WorkflowResultEvent,
) (err error) {
	var (
		wid    = event.WorkflowID()
		result = event.Result
	)
	ref, pst := workflows.Get(wid)
	if !pst {
		r.reportSkipEvent(evnt.EventTypeWorkfowResult, wid)
		return
	}
	_, err = ref.Workflow().Progress().SetResult(func() (wfl.Result, error) {
		return result, nil
	})
	if err != nil {
		return
	}
	terminated.wids = append(terminated.wids, wid)
	r.reportEvent(evnt.EventTypeWorkfowResult, wid, partition, offset)
	return
}

func (r Restorer) handleDeadLetteredEvent(
	workflows *support.OrderedMap[*WorkflowRef, wfl.ID],
	terminated *struct{ wids []wfl.ID },
	partition int,
	offset int64,
	event *evnt.DeadLetterEvent,
) (err error) {
	wid := event.WorkflowID()
	ref, pst := workflows.Get(wid)
	if !pst {
		r.reportSkipEvent(evnt.EventTypeDeadLetter, wid)
		return
	}
	ref.MarkDeadLettered()
	terminated.wids = append(terminated.wids, wid)
	r.reportEvent(evnt.EventTypeDeadLetter, wid, partition, offset)
	return
}

func (r Restorer) handleRejectedEvent(
	workflows *support.OrderedMap[*WorkflowRef, wfl.ID],
	terminated *struct{ wids []wfl.ID },
	partition int,
	offset int64,
	event *evnt.WorkflowRejectedEvent,
) (err error) {
	wid := event.WorkflowID()
	ref, pst := workflows.Get(wid)
	if !pst {
		r.reportSkipEvent(evnt.EventTypeRejected, wid)
		return
	}
	ref.MarkRejected()
	terminated.wids = append(terminated.wids, wid)
	r.reportEvent(evnt.EventTypeRejected, wid, partition, offset)
	return
}

func (r Restorer) reportReconstructionStarting(count int) {
	r.logger.Info("runtime reconstruction starting", "count", count)
}

func (r Restorer) reportReconstructionComplete(duration time.Duration,
	chptman *chptsup.Manager,
) {
	r.logger.Info("runtime reconstruction complete",
		"duration", duration, "chpt", chptman)
}

func (r Restorer) reportRestorationStateSnapshot(chptman *chptsup.Manager,
	tracker *PositionTracker,
) {
	r.logger.Debug("restoration state snapshot",
		"chptman", chptman,
		"tracker", tracker,
	)
}

func (r Restorer) reportMissingTrackingData(wid wfl.ID) {
	r.logger.Error("tracker knows nothing about the wid", "wid", wid)
}

func (r Restorer) reportSkipEvent(eventType evnt.EventType, wid wfl.ID) {
	msg := fmt.Sprintf("skip %s event for historical/unknown workflow",
		eventType)
	r.logger.Debug(msg, "wid", wid)
}

func (r Restorer) reportStepOutcomeEvent(outcomeType wfl.OutcomeType,
	outcomeSeq wfl.OutcomeSeq,
	wid wfl.ID,
	partition int,
	offset int64,
) {
	r.logger.Debug("found event",
		"event_type", evnt.EventTypeStepOutcome,
		"wid", wid,
		"outcome_type", outcomeType,
		"outcome_seq", outcomeSeq,
		"partition", partition,
		"offset", offset,
	)
}

func (r Restorer) reportEvent(eventType evnt.EventType, wid wfl.ID, partition int,
	offset int64,
) {
	r.logger.Debug("found event",
		"event_type", eventType,
		"wid", wid,
		"partition", partition,
		"offset", offset,
	)
}

func (r Restorer) reportIdempotencyWindowOnly(wid wfl.ID) {
	r.logger.Debug("idempotency window only",
		"event_type", evnt.EventTypeWorkflowCreated,
		"wid", wid,
	)
}

func (r Restorer) reportDuplicateEvent(eventType evnt.EventType, wid wfl.ID,
	partition int,
	offset int64,
) {
	r.logger.Debug("skip duplicate event",
		"event_type", eventType,
		"wid", wid,
		"partition", partition,
		"offset", offset,
	)
}
