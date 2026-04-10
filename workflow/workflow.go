package workflow

import (
	"context"
)

const (
	// Executing indicates the workflow is currently running its forward steps.
	Executing State = iota + 1
	// Compensating indicates the workflow is rolling back due to a failure.
	Compensating
	// Finalizing indicates all steps (or compensations) have finished, and the
	// final result is being computed.
	Finalizing
	// Completed indicates the workflow has reached a terminal state.
	Completed
)

type State int

type Seq int64

// Workflow represents a durable, multi-step business process.
//
// It serves as the primary abstraction for managing state transitions and
// ensuring that a sequence of operations is either completed successfully
// or rolled back through a series of compensation actions.
//
// Workflows are inherently stateful and are typically restored from an
// event log to ensure reliability across node failures.
type Workflow interface {
	// NodeID returns the identifier of the engine instance currently owning
	// this workflow.
	NodeID() string

	// Run executes the workflow logic.
	//
	// It orchestrates the transition through individual steps and handles
	// automatic compensation if a step fails.
	//
	// Returns:
	//   - result: The final output of the workflow if it completes successfully.
	//   - err: A step execution error (ExecutionError), a compensation failure
	//     (CompensationError), a protocol violation (ProtocolViolationError),
	//     or a circuit breaker error (codes.CircuitBreakerOpenError) from
	//     external services.
	Run(ctx context.Context) (result Result, err error)

	// State returns the current lifecycle status of the workflow.
	State() State

	// RefreshState re-synchronizes the internal state based on the
	// recorded progress.
	RefreshState()

	// Spec returns the immutable specification of the workflow.
	Spec() Spec

	// Progress returns the tracker for the workflow's execution history.
	Progress() *Progress

	// Equal checks if this workflow instance represents the same business
	// process as another.
	Equal(other Workflow) bool
}
