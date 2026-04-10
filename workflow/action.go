package workflow

import "context"

// Action represents a generic task or step logic within a workflow.
// It receives the workflow's initial input and its current execution history
// (Progress), allowing for state-aware decision making.
type Action[T any] interface {
	// Run executes the action logic.
	Run(ctx context.Context, input Input, progress *Progress) (T, error)
}

// ActionFn is a functional adapter that allows ordinary functions to be used
// as Actions.
type ActionFn[T any] func(ctx context.Context, input Input, progress *Progress) (
	T, error)

// Run implements the Action interface by calling the underlying function.
func (a ActionFn[T]) Run(ctx context.Context, input Input, progress *Progress) (
	T, error,
) {
	return a(ctx, input, progress)
}
