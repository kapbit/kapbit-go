package workflow

import "fmt"

// ValueOutcome is a generic implementation of the Outcome interface that
// holds a value of type T.
type ValueOutcome[T any] struct {
	Value  T
	IsFail bool
}

// Failure returns true if this outcome represents a failed operation.
func (o ValueOutcome[T]) Failure() bool {
	return o.IsFail
}

// String returns a string representation of the outcome.
// If it's a failure outcome, it includes a [FAIL] prefix.
func (o ValueOutcome[T]) String() string {
	if o.IsFail {
		return fmt.Sprintf("[FAIL] %v", o.Value)
	}
	return fmt.Sprintf("%v", o.Value)
}

// Success creates a successful ValueOutcome with the provided value.
func Success[T any](v T) ValueOutcome[T] {
	return ValueOutcome[T]{Value: v, IsFail: false}
}

// Fail creates a failed ValueOutcome with the provided value.
func Fail[T any](v T) ValueOutcome[T] {
	return ValueOutcome[T]{Value: v, IsFail: true}
}
