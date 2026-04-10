package workflow

// Step represents a single unit of execution within a workflow.
// It consists of a primary action and an optional compensation action for
// rolling back changes in case of failure.
type Step struct {
	Execute    Action[Outcome]
	Compensate Action[Outcome]
}
