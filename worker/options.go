package worker

import (
	"fmt"
	"log/slog"

	supretry "github.com/kapbit/kapbit-go/support/retry"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
)

type Options struct {
	MaxAttempts int
	Policy      supretry.Policy
	PolicyOpts  []swtch.SetOption
	Logger      *slog.Logger
}

func (o *Options) SetDefaults() {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 10
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.Policy == nil {
		var err error
		if o.Policy, err = swtch.NewSwitch(o.PolicyOpts...); err != nil {
			panic(err)
		}
	}
}

func (o *Options) Validate() error {
	if o.MaxAttempts <= 0 {
		return fmt.Errorf("option RetriesCount must be positive (got %d)",
			o.MaxAttempts)
	}
	return nil
}

func Apply(o *Options, opts ...SetOption) {
	for _, opt := range opts {
		opt(o)
	}
}

type SetOption func(*Options)

func WithMaxAttempts(n int) SetOption {
	return func(o *Options) {
		o.MaxAttempts = n
	}
}

func WithPolicy(policy supretry.Policy) SetOption {
	return func(o *Options) {
		o.Policy = policy
	}
}

func WithPolicyOptions(opts ...swtch.SetOption) SetOption {
	return func(o *Options) {
		o.PolicyOpts = opts
	}
}

func WithLogger(logger *slog.Logger) SetOption {
	return func(o *Options) {
		o.Logger = logger
	}
}
