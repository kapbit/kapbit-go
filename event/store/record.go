package store

import "strconv"

const (
	MetaKeyEventType MetaKey = "event_type"

	MetaKeyNodeID         MetaKey = "node_id"
	MetaKeyWorkflowType   MetaKey = "workflow_type"
	MetaKeyOutcomeSeq     MetaKey = "outcome_seq"
	MetaKeyOutcomeType    MetaKey = "outcome_type"
	MetaKeyFailure        MetaKey = "failure"
	MetaKeyWorkflowOffset MetaKey = "workflow_offset"
	MetaKeyReason         MetaKey = "reason"
	MetaKeyEmptyPayload   MetaKey = "no_outcome"
	MetaKeyTimestamp      MetaKey = "timestamp"

	MetaKeyPartition MetaKey = "partition"
)

const (
	EventTypeUndefined EventType = iota
	EventTypeActiveWriter
	EventTypeWorkflowCreated
	EventTypeStepOutcome
	EventTypeWorkflowResult
	EventTypeDeadLetter
	EventTypeRejection
	EventTypeCheckpoint
)

// EventType is the type of the event.
//
// Part of the EventRecord.Meta.
type EventType int

// Value returns the string representation of the underlying integer.
// Example: EventTypeWorkflowCreated -> "2"
func (t EventType) Value() string {
	return strconv.Itoa(int(t))
}

func (t EventType) String() string {
	switch t {
	case EventTypeUndefined:
		return "UndefinedEvent"
	case EventTypeActiveWriter:
		return "ActiveWriterEvent"
	case EventTypeWorkflowCreated:
		return "WorkflowCreatedEvent"
	case EventTypeStepOutcome:
		return "StepOutcomeEvent"
	case EventTypeWorkflowResult:
		return "WorkflowResultEvent"
	case EventTypeDeadLetter:
		return "DeadLetterEvent"
	case EventTypeRejection:
		return "RejectionEvent"
	case EventTypeCheckpoint:
		return "CheckpointEvent"
	default:
		return "UnknownEvent"
	}
}

// ParseEventType converts a numeric string back into an EventType.
func ParseEventType(s string) EventType {
	val, err := strconv.Atoi(s)
	if err != nil {
		return EventTypeUndefined
	}
	return EventType(val)
}

type MetaKey string

type EventRecord struct {
	Key     string
	Meta    map[MetaKey]string
	Payload []byte
}

func (r EventRecord) ActiveWriterEvent() bool {
	return r.Meta[MetaKeyEventType] == EventTypeActiveWriter.Value()
}

func (r EventRecord) WorkflowCreatedEvent() bool {
	return r.Meta[MetaKeyEventType] == EventTypeWorkflowCreated.Value()
}

func (r EventRecord) OutcomeEvent() bool {
	return r.Meta[MetaKeyEventType] == EventTypeStepOutcome.Value()
}

func (r EventRecord) TerminalEvent() bool {
	t := r.Meta[MetaKeyEventType]
	return t == EventTypeWorkflowResult.Value() ||
		t == EventTypeDeadLetter.Value() ||
		t == EventTypeRejection.Value()
}

func (r EventRecord) EventType() EventType {
	return ParseEventType(r.Meta[MetaKeyEventType])
}
