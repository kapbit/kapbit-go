package workflow

import (
	"log/slog"
)

// Spec contains the complete definition and identity of a workflow instance.
type Spec struct {
	ID    ID
	Def   Definition
	Input Input
}

type (
	// ID represents a unique identifier for a specific workflow instance.
	ID string
	// Type represents the registered name of a workflow category.
	Type string
	// Input represents the initial data passed to a workflow.
	Input any
)

// NodeID represents a unique identifier for a Kapbit engine instance.
type NodeID string

// Short returns a truncated version of the NodeID for easier display.
func (id NodeID) Short() string {
	if len(id) > 8 {
		return string(id[:8])
	}
	return string(id)
}

// LogValue implements the slog.LogValuer interface, ensuring IDs are
// consistently formatted in logs.
func (id NodeID) LogValue() slog.Value {
	return slog.StringValue(id.Short())
}

func (id NodeID) String() string {
	return string(id)
}
