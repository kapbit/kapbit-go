package workflow

import (
	"errors"
	"fmt"
)

// -- Factory errors --

type WorkflowTypeNotRegisteredError struct {
	wid   ID
	wtype Type
}

func NewWorkflowTypeNotRegisteredError(wid ID, wtype Type) error {
	return &WorkflowTypeNotRegisteredError{
		wid:   wid,
		wtype: wtype,
	}
}

func (e *WorkflowTypeNotRegisteredError) Error() string {
	return fmt.Sprintf("workflow type %q not registered, wid=%q", e.wtype, e.wid)
}

func (e *WorkflowTypeNotRegisteredError) WorkflowID() ID {
	return e.wid
}

func (e *WorkflowTypeNotRegisteredError) WorkflowType() Type {
	return e.wtype
}

// -- Workflow errors

func NewExecutionError(cause error) error {
	return fmt.Errorf("execution error: %w", cause)
}

func NewCompensationError(cause error) error {
	return fmt.Errorf("compensation error: %w", cause)
}

func NewCompletionError(cause error) error {
	return fmt.Errorf("completion error: %w", cause)
}

// -- Progress errors --

type ProtocolViolationError struct {
	wid   ID
	wtype Type
	cause error
}

func NewProtocolViolationError(wid ID, wtype Type, cause error) error {
	return &ProtocolViolationError{
		wid:   wid,
		wtype: wtype,
		cause: cause,
	}
}

func (e *ProtocolViolationError) WorkflowID() ID {
	return e.wid
}

func (e *ProtocolViolationError) Unwrap() error {
	return e.cause
}

func (e *ProtocolViolationError) Error() string {
	return fmt.Sprintf("protocol violation in workflow %s: %s", e.wid, e.cause)
}

var (
	ErrInvalidRollback = errors.New("invalid rollback: compensation triggered " +
		"but the last execution step was successful")
	ErrExecutionClosed = errors.New("cannot record execution outcome: workflow " +
		"is already in a compensation or complete state")
	ErrCompensationClosed = errors.New("cannot record compensation outcome: " +
		"workflow is already in a terminal state or a failed compensation")
)

func NewUnexpectedExecutionOutcomeSeq(seq OutcomeSeq) error {
	return fmt.Errorf("unexpected execution outcome. seq %d", seq)
}

func NewUnexpectedCompensationOutcomeSeq(seq OutcomeSeq) error {
	return fmt.Errorf("unexpected compensation outcome seq %d", seq)
}

func NewCompensationOverflowError(compSeq OutcomeSeq, lenExec int) error {
	format := "compensation overflow: attempt to compensate step %d, but only " +
		"%d steps have executed"
	return fmt.Errorf(format, compSeq, lenExec)
}

func NewIncompleteCompensationError(lenExec, lenComp int) error {
	format := "illegal terminal state: partial compensation (exec=%d, comp=%d) " +
		"but last compensation success"

	return fmt.Errorf(format, lenExec, lenComp)
}
