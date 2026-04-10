package persist

import (
	"fmt"
	"log/slog"

	"github.com/ymz-ncnk/circbrk-go"
)

type Options struct {
	Logger         *slog.Logger
	CircuitBreaker []circbrk.SetOption
}

func (o *Options) Validate() error {
	if o.Logger == nil {
		return fmt.Errorf("option Logger is required")
	}
	return nil
}

func DefaultOptions() Options {
	return Options{
		Logger: slog.Default(),
	}
}

func Apply(o *Options, opts ...SetOption) error {
	for _, opt := range opts {
		opt(o)
	}
	return o.Validate()
}

type SetOption func(*Options)

func WithLogger(logger *slog.Logger) SetOption {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithCircuitBreaker(opts ...circbrk.SetOption) SetOption {
	return func(o *Options) {
		o.CircuitBreaker = opts
	}
}
