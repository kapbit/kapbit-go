package workflow

import (
	"fmt"
)

const (
	// OutcomeTypeExecution represent a regular step execution outcome.
	OutcomeTypeExecution OutcomeType = iota + 1

	// OutcomeTypeCompensation represent a rollback step execution outcome.
	OutcomeTypeCompensation
)

// NoOutcome is a placeholder for when no execution or compensation outcome has
// been recorded yet.
var NoOutcome Outcome = nil

// OutcomeSeq represents the sequential order of outcomes within a workflow.
type OutcomeSeq int64

// OutcomeType defines the category of effort (forward execution or rollback).
type OutcomeType int

func (t OutcomeType) String() string {
	switch t {
	case OutcomeTypeExecution:
		return "Execution"
	case OutcomeTypeCompensation:
		return "Compensation"
	default:
		return fmt.Sprintf("unknown outcome type %d", int(t))
	}
}

// Outcome represents the result of a single task or step within a workflow.
// It can signify either a successful completion or a failure that requires
// attention or compensation.
type Outcome interface {
	// Failure returns true if the outcome represents a failed operation.
	Failure() bool
}
