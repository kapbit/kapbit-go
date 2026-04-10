package workflow

// Progress tracks the execution and compensation history of a workflow.
// It is the source of truth used for recovery, idempotency, and state
// transitions.
type Progress struct {
	wid          ID
	wtype        Type
	execution    []Outcome
	compensation []Outcome
	result       Result
}

// NewProgress creates a new, empty Progress tracker for a specific workflow.
func NewProgress(wid ID, wtype Type) *Progress {
	return &Progress{
		wid:          wid,
		wtype:        wtype,
		execution:    []Outcome{},
		compensation: []Outcome{},
	}
}

func MustProgress(wid ID, wtype Type, execution,
	compensation []Outcome, result Result,
) *Progress {
	return &Progress{
		wid:          wid,
		wtype:        wtype,
		execution:    execution,
		compensation: compensation,
		result:       result,
	}
}

// RecordExecutionOutcome adds or updates a forward step outcome.
//
// Idempotency:
//   - If seq matches the next available slot, the outcome is appended.
//   - If seq matches the last slot, the outcome is updated (resumption).
//
// Returns ProtocolViolationError if the sequence is invalid or if the
// workflow is already compensating or completed.
func (p *Progress) RecordExecutionOutcome(seq OutcomeSeq, outcome Outcome) error {
	err := p.validateExecutionPreconditions()
	if err != nil {
		return NewProtocolViolationError(p.wid, p.wtype, err)
	}
	var (
		lenExec = len(p.execution)
		n       = int(seq)
	)
	if n == lenExec {
		p.execution = append(p.execution, outcome)
		return nil
	}
	// We expect the last event could be duplicated, so rewrite it!
	if lenExec > 0 && n == lenExec-1 {
		p.execution[seq] = outcome
		return nil
	}
	err = NewUnexpectedExecutionOutcomeSeq(seq)
	return NewProtocolViolationError(p.wid, p.wtype, err)
}

// RecordCompensationOutcome adds or updates a rollback step outcome.
//
// Idempotency:
//   - If seq matches the next available slot, the outcome is appended.
//   - If seq matches the last slot, the outcome is updated (resumption).
//
// Returns ProtocolViolationError if the sequence is invalid or if the
// compensation rules are violated.
func (p *Progress) RecordCompensationOutcome(seq OutcomeSeq, outcome Outcome) error {
	var (
		lenComp = len(p.compensation)
		n       = int(seq)
	)
	if n == lenComp {
		if err := p.validateCompensationPreconditions(); err != nil {
			return NewProtocolViolationError(p.wid, p.wtype, err)
		}
		p.compensation = append(p.compensation, outcome)
		return nil
	}
	// We expect the last event could be duplicated, so rewrite it!
	if lenComp > 0 && n == lenComp-1 {
		p.compensation[n] = outcome
		return nil
	}
	err := NewUnexpectedCompensationOutcomeSeq(seq)
	return NewProtocolViolationError(p.wid, p.wtype, err)
}

// SetResult executes a result builder function and stores its output.
//
// It ensures the workflow has reached a valid terminal state (either complete
// success or a terminal failure state) before allowing a result to be set.
//
// Returns ProtocolViolationError if the workflow is in an invalid state.
func (p *Progress) SetResult(result func() (Result, error)) (
	Result, error,
) {
	err := p.validateTerminalState()
	if err != nil {
		return nil, NewProtocolViolationError(p.wid, p.wtype, err)
	}
	p.result, err = result()
	return p.result, err
}

// LastExecutionOutcome returns the index and data of the final forward step.
func (p *Progress) LastExecutionOutcome() (index int, outcome Outcome,
	ok bool,
) {
	l := len(p.execution)
	if l == 0 {
		return
	}
	index = l - 1
	return index, p.execution[index], true
}

// LastCompensationOutcome returns the index and data of the final rollback step.
func (p *Progress) LastCompensationOutcome() (index int, outcome Outcome,
	ok bool,
) {
	l := len(p.compensation)
	if l == 0 {
		return
	}
	index = l - 1
	return index, p.compensation[index], true
}

func (p *Progress) ForEachExecution(fn func(Outcome) error) (err error) {
	for _, r := range p.execution {
		if err = fn(r); err != nil {
			return
		}
	}
	return
}

func (p *Progress) ForEachCompensation(fn func(Outcome) error) (err error) {
	for _, r := range p.compensation {
		if err = fn(r); err != nil {
			return
		}
	}
	return
}

func (p *Progress) Result() Result {
	return p.result
}

// ExecutionLen returns the number of forward steps executed.
func (p *Progress) ExecutionLen() int {
	return len(p.execution)
}

// CompensationLen returns the number of rollback steps executed.
func (p *Progress) CompensationLen() int {
	return len(p.compensation)
}

// CompensationFailed returns true if the final compensation attempt was
// unsuccessful.
func (p *Progress) CompensationFailed() bool {
	_, outcome, ok := p.LastCompensationOutcome()
	if !ok {
		return false
	}
	return outcome.Failure()
}

func (p *Progress) WorkflowID() ID {
	return p.wid
}

// Empty returns true if no execution progress has been made yet.
func (p *Progress) Empty() bool {
	return len(p.execution) == 0
}

func (p *Progress) validateExecutionPreconditions() error {
	if len(p.compensation) > 0 || p.result != nil {
		return ErrExecutionClosed
	}
	_, outcome, ok := p.LastExecutionOutcome()
	if ok && outcome != NoOutcome && outcome.Failure() {
		return ErrExecutionClosed
	}
	return nil
}

// validateCompensationPreconditions checks if adding a new compensation step
// is allowed. Rules:
// 1. Cannot exceed execution length (can't undo what hasn't happened).
// 2. Cannot compensate if the workflow hasn't actually failed yet.
func (p *Progress) validateCompensationPreconditions() error {
	if p.result != nil {
		return ErrCompensationClosed
	}

	var (
		lenComp = len(p.compensation)
		lenExec = len(p.execution)
	)

	// Rule 1: We cannot compensate if any previous compensation step failed.
	if lenComp > 0 {
		lastComp := p.compensation[lenComp-1]
		if lastComp != NoOutcome && lastComp.Failure() {
			return ErrCompensationClosed
		}
	}

	// Rule 2: Boundary Check
	if lenComp >= lenExec {
		return NewCompensationOverflowError(OutcomeSeq(lenComp+1), lenExec)
	}

	// Rule 3: Trigger Check (Must be in a Failure state)
	_, lastExec, ok := p.LastExecutionOutcome()
	if ok && (lastExec == NoOutcome || !lastExec.Failure()) {
		return ErrInvalidRollback
	}
	return nil
}

func (p *Progress) validateTerminalState() error {
	var (
		lenExec = len(p.execution)
		lenComp = len(p.compensation)
	)
	// lenComp >= lenExec checked in validateCompensationPreconditions
	if lenComp > 0 {
		// Valid Case A: Full Rollback (lenComp == lenExec+1)
		if lenComp == lenExec-1 {
			return nil
		}

		// Valid Case B: Stuck Rollback (lenComp < lenExec, but last comp failed)
		lastComp := p.compensation[lenComp-1]
		if lastComp.Failure() {
			return nil
		}
		return NewIncompleteCompensationError(lenExec, lenComp)
	}
	return nil
}
