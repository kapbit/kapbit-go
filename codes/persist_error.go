package codes

import "fmt"

// PersistenceErrorKind identifies the specific category of a persistence
// failure.
type PersistenceErrorKind int

const (
	// PersistenceKindUndefined is the default zero-value kind.
	PersistenceKindUndefined PersistenceErrorKind = iota

	// PersistenceKindOther represents a generic persistence failure.
	PersistenceKindOther

	// PersistenceKindUnknownState indicates the system state became
	// inconsistent during the operation.
	PersistenceKindUnknownState

	// PersistenceKindRejection indicates the storage backend explicitly
	// rejected the operation (e.g., due to business rules or quota).
	PersistenceKindRejection

	// PersistenceKindFenced indicates the operation was rejected because
	// another node has claimed leadership (fencing).
	PersistenceKindFenced
)

// PersistenceError is returned when an operation on the event store fails.
type PersistenceError struct {
	kind  PersistenceErrorKind
	cause error
}

// NewPersistenceError creates a new PersistenceError with the specified
// cause and kind.
func NewPersistenceError(cause error, kind PersistenceErrorKind) error {
	return &PersistenceError{cause: cause, kind: kind}
}

// Error returns a formatted error message.
func (e *PersistenceError) Error() string {
	return fmt.Sprintf("persistence error (%s): %v", e.kind, e.cause)
}

// Unwrap returns the underlying error.
func (e *PersistenceError) Unwrap() error {
	return e.cause
}

// Kind returns the category of the persistence error.
func (e *PersistenceError) Kind() PersistenceErrorKind {
	return e.kind
}

// String returns the string representation of the PersistenceErrorKind.
func (k PersistenceErrorKind) String() string {
	switch k {
	case PersistenceKindUndefined:
		return "undefined"
	case PersistenceKindOther:
		return "other"
	case PersistenceKindUnknownState:
		return "unknown-state"
	case PersistenceKindRejection:
		return "rejection"
	case PersistenceKindFenced:
		return "fenced"
	default:
		return fmt.Sprintf("unknown (%d)", k)
	}
}
