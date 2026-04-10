package runtime

import (
	"sync"

	wfl "github.com/kapbit/kapbit-go/workflow"
)

// WorkflowRef is a runtime workflow ref.
//
// retriesCount and outputSaveRejected are not persisted anywhere, so on
// restart they will be reset.
type WorkflowRef struct {
	workflow     wfl.Workflow
	retryCount   int
	saved        bool
	deadLettered bool
	rejected     bool
	mu           sync.Mutex
}

func NewWorkflowRef(workflow wfl.Workflow, retriesCount int) *WorkflowRef {
	return &WorkflowRef{workflow: workflow, retryCount: retriesCount}
}

func NewSavedWorkflowRef(workflow wfl.Workflow, retriesCount int) *WorkflowRef {
	return &WorkflowRef{workflow: workflow, retryCount: retriesCount, saved: true}
}

func (r *WorkflowRef) Workflow() wfl.Workflow {
	return r.workflow
}

func (r *WorkflowRef) IncrementRetriesCount(threshold int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.retryCount+1 >= threshold {
		return false
	}
	r.retryCount++
	return true
}

func (r *WorkflowRef) MarkSaved() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = true
}

// MarkDeadLettered marks the workflow as dead lettered.
//
// Usefull for restored refs.
func (r *WorkflowRef) MarkDeadLettered() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deadLettered = true
}

// MarkRejected marks the workflow as rejected.
//
// Usefull for restored refs.
func (r *WorkflowRef) MarkRejected() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rejected = true
}

func (r *WorkflowRef) RetryCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.retryCount
}

func (r *WorkflowRef) Saved() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saved
}

func (r *WorkflowRef) DeadLettered() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.deadLettered
}

func (r *WorkflowRef) Rejected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rejected
}

func (r *WorkflowRef) Equal(other *WorkflowRef) bool {
	if r == nil || other == nil {
		return r == other
	}
	if r == other {
		return true // Same memory address
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	other.mu.Lock()
	defer other.mu.Unlock()

	return r.retryCount == other.retryCount &&
		r.saved == other.saved &&
		r.deadLettered == other.deadLettered &&
		r.rejected == other.rejected &&
		r.workflow.Equal(other.workflow) // Delegate to workflow's Equal
}
