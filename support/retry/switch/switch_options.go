package swtch

import (
	"fmt"
	"time"
)

// Options configures the two-state (fast/slow) policy.
type Options struct {
	NormalInterval time.Duration // The interval when in "Fast" or "Normal" mode
	SlowInterval   time.Duration // The interval when in "Slow" or "Idle" mode
	UseJitter      bool          // Whether to apply jitter to the intervals
}

// SetDefaults applies sensible default values to zero fields.
func (o *Options) SetDefaults() {
	if o.NormalInterval <= 0 {
		o.NormalInterval = 100 * time.Millisecond
	}
	if o.SlowInterval <= 0 {
		o.SlowInterval = 5 * time.Second
	}
}

// Validate checks for configuration errors.
func (o Options) Validate() error {
	if o.NormalInterval <= 0 {
		return fmt.Errorf("option FastInterval must be positive (got %s)", o.NormalInterval)
	}
	if o.SlowInterval <= 0 {
		return fmt.Errorf("option SlowInterval must be positive (got %s)", o.SlowInterval)
	}
	// It is technically valid for Fast > Slow, though unusual, so we don't enforce Fast < Slow.
	return nil
}

// Apply executes the functional options against the struct.
func Apply(o *Options, opts ...SetOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// SetOption defines the functional option type.
type SetOption func(*Options)

func WithNormalInterval(d time.Duration) SetOption {
	return func(o *Options) { o.NormalInterval = d }
}

func WithSlowInterval(d time.Duration) SetOption {
	return func(o *Options) { o.SlowInterval = d }
}

func WithJitter() SetOption {
	return func(o *Options) { o.UseJitter = true }
}
