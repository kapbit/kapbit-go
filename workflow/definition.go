package workflow

import (
	"errors"
	"fmt"
)

// Definition defines the static structure and logic of a workflow.
// It acts as a blueprint, specifying the sequence of steps and how to
// construct the final result.
type Definition struct {
	t             Type
	resultBuilder Action[Result]
	steps         []Step
}

// NewDefinition creates a new workflow definition and validates its components.
//
// Validation Rules:
//   - A ResultBuilder must be provided.
//   - Every step in the workflow must have an Execute action.
func NewDefinition(t Type, resultBuilder Action[Result], steps ...Step) (
	Definition, error,
) {
	if resultBuilder == nil {
		return Definition{}, errors.New("result builder is required")
	}
	for i, s := range steps {
		if s.Execute == nil {
			return Definition{}, fmt.Errorf("step %d is missing execute action", i)
		}
	}
	return Definition{
		t:             t,
		resultBuilder: resultBuilder,
		steps:         steps,
	}, nil
}

// MustDefinition is a helper for creating definitions in tests or static
// initializers where inputs are guaranteed to be valid.
func MustDefinition(t Type, resultBuilder Action[Result], steps ...Step) Definition {
	return Definition{
		t:             t,
		resultBuilder: resultBuilder,
		steps:         steps,
	}
}

// Type returns the unique category identifier of the workflow.
func (d Definition) Type() Type {
	return d.t
}

// ResultBuilder returns the action used to construct the workflow's final output.
func (d Definition) ResultBuilder() Action[Result] {
	return d.resultBuilder
}

// Steps returns the sequence of execution steps defined for the workflow.
func (d Definition) Steps() []Step {
	return d.steps
}

// Definitions is a collection of workflow blueprints.
type Definitions []Definition

// ToMap converts the slice of definitions into a map keyed by workflow Type
// for efficient lookup.
func (d Definitions) ToMap() map[Type]Definition {
	m := make(map[Type]Definition, len(d))
	for _, def := range d {
		m[def.t] = def
	}
	return m
}
