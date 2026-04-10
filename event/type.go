package event

const (
	// Workflow-level events
	EventTypeWorkflowCreated EventType = "workflow.created"
	EventTypeWorkfowResult   EventType = "workflow.result"
	// TODO Remove
	EventTypeWorkflowCompensationTriggered EventType = "workflow.compensation_triggered"

	// Step-level events (Emitted during workflow execution)
	EventTypeStepOutcome EventType = "workflow.step_outcome"

	// Infrastructure/Worker events
	// TODO Remove
	EventTypeRetryAttempt EventType = "worker.retry_attempt"

	// Protocol/System events
	EventTypeRejected   EventType = "system.rejected"
	EventTypeDeadLetter EventType = "system.dead_letter"
)

type EventType string
