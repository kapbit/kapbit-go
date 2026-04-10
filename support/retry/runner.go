package retry

import (
	"context"
	"time"
)

type Task func(ctx context.Context) (state int, stop bool, err error)

type Runner struct {
	policy Policy
}

func NewRunner(policy Policy) Runner {
	return Runner{policy: policy}
}

// Run executes 'task' repeatedly.
// The task returns a 'state' (int) which is passed to the policy to determine the next wait.
func (r Runner) Run(ctx context.Context, task Task) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	// Initial wait is 0 (immediate execution)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}

	for {
		// 1. Execute the Task
		state, stop, err := task(ctx)
		if stop {
			return err
		}

		// 2. Ask Policy for the delay based on the task's result state
		delay := r.policy.NextDelay(state)

		// 3. Wait (The Runner handles the timer mechanics)
		timer.Reset(delay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			// Loop continues
		}
	}
}
