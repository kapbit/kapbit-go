package retry

import (
	"fmt"
	"log/slog"

	supretry "github.com/kapbit/kapbit-go/support/retry"
	expnt "github.com/kapbit/kapbit-go/support/retry/exponential"
)

type OnTransientError func(err error)

type Options struct {
	Logger           *slog.Logger
	Policy           supretry.Policy
	OnTransientError OnTransientError
}

func (o *Options) Validate() error {
	if o.Logger == nil {
		return fmt.Errorf("option Logger is required")
	}
	if o.Policy == nil {
		return fmt.Errorf("option Policy is required")
	}
	return nil
}

func DefaultOptions() Options {
	return Options{
		Logger: slog.Default(),
		Policy: expnt.MustNewExponential(),
	}
}

func Apply(o *Options, opts ...SetOption) error {
	for _, opt := range opts {
		opt(o)
	}
	return o.Validate()
}

type SetOption func(*Options)

func WithPolicy(policy supretry.Policy) SetOption {
	return func(o *Options) {
		o.Policy = policy
	}
}

func WithLogger(logger *slog.Logger) SetOption {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithOnTransientError(onTransientError OnTransientError) SetOption {
	return func(o *Options) {
		o.OnTransientError = onTransientError
	}
}
