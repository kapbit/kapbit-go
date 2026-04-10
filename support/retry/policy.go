package retry

import "time"

// pkg/support/retry/policy.go

type Policy interface {
	// NextDelay calculates the delay based on an input state (e.g., failure count, mode)
	NextDelay(state int) time.Duration
}
