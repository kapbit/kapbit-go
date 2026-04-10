package exponential

import (
	"fmt"
	"time"
)

// Options configures the exponential backoff policy.
type Options struct {
	InitialInterval time.Duration // Wait time after the first failure
	MaxInterval     time.Duration // Upper bound for the wait time
	Multiplier      float64       // Factor to multiply the interval by
	// MaxAttempts     int           // Maximum retries (0 = infinite)
	UseJitter bool // Add randomization to prevents thundering herd
}

// Validate checks for configuration errors.
func (o Options) Validate() error {
	if o.InitialInterval <= 0 {
		return fmt.Errorf("option InitialInterval must be positive (got %s)", o.InitialInterval)
	}
	if o.MaxInterval <= 0 {
		return fmt.Errorf("option MaxInterval must be positive (got %s)", o.MaxInterval)
	}
	if o.MaxInterval < o.InitialInterval {
		return fmt.Errorf("option MaxInterval (%s) cannot be less than initial interval (%s)",
			o.MaxInterval, o.InitialInterval)
	}
	if o.Multiplier <= 1.0 {
		return fmt.Errorf("option Multiplier must be greater than 1.0 (got %.2f)", o.Multiplier)
	}
	return nil
}

// DefaultOptions returns the default configuration for exponential backoff.
func DefaultOptions() Options {
	return Options{
		InitialInterval: 200 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      1.5,
	}
}

// Apply executes the functional options against the struct.
func Apply(o *Options, opts ...SetOption) error {
	for _, opt := range opts {
		opt(o)
	}
	return o.Validate()
}

// SetOption defines the functional option type.
type SetOption func(*Options)

func WithInitialInterval(d time.Duration) SetOption {
	return func(o *Options) { o.InitialInterval = d }
}

func WithMaxInterval(d time.Duration) SetOption {
	return func(o *Options) { o.MaxInterval = d }
}

func WithMultiplier(f float64) SetOption {
	return func(o *Options) { o.Multiplier = f }
}

// func WithMaxAttempts(n int) SetExponentialOption {
// 	return func(o *ExponentialOptions) { o.MaxAttempts = n }
// }

func WithJitter() SetOption {
	return func(o *Options) { o.UseJitter = true }
}
