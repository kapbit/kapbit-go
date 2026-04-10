// Package kapbit provides the primary entry point for the Kapbit workflow orchestrator.
package kapbit

import (
	"errors"
	"log/slog"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/persist"
	retry "github.com/kapbit/kapbit-go/event/retry"

	rtm "github.com/kapbit/kapbit-go/runtime"
	wrk "github.com/kapbit/kapbit-go/worker"
)

// TimeProvider is a type alias for the event package's TimeProvider.
type TimeProvider = evnt.TimeProvider

// Options configures a Kapbit instance.
type Options struct {
	Logger               *slog.Logger
	NodeID               string
	Runtime              []rtm.SetOption
	Worker               []wrk.SetOption
	Emitter              []retry.SetOption
	PersistHandler       []persist.SetOption
	TimeProvider         TimeProvider
	ShutdownPollInterval time.Duration
}

// DefaultOptions returns the default configuration for Kapbit.
func DefaultOptions() Options {
	return Options{
		Logger: slog.Default(),
		TimeProvider: evnt.TimeProviderFn(
			func() int64 { return time.Now().Unix() },
		),
		ShutdownPollInterval: 500 * time.Millisecond,
	}
}

// Validate checks if the options are valid.
func (o *Options) Validate() error {
	if o.NodeID == "" {
		return errors.New("kapbit: NodeID is required")
	}
	if o.Logger == nil {
		return errors.New("kapbit: Logger is required")
	}
	if o.TimeProvider == nil {
		return errors.New("kapbit: TimeProvider is required")
	}
	return nil
}

// Apply applies the given options to the Options struct and validates the result.
func Apply(o *Options, opts ...SetOption) error {
	for _, opt := range opts {
		opt(o)
	}
	return o.Validate()
}

// SetOption is a function that modifies the Options struct.
type SetOption func(*Options)

// WithLogger sets the logger for Kapbit.
func WithLogger(logger *slog.Logger) SetOption {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithNodeID sets the unique identifier for this Kapbit instance.
func WithNodeID(id string) SetOption {
	return func(o *Options) {
		o.NodeID = id
	}
}

// WithRuntime sets the options for Kapbit's runtime.
func WithRuntime(opts ...rtm.SetOption) SetOption {
	return func(o *Options) {
		o.Runtime = opts
	}
}

// WithWorker sets the options for Kapbit's worker.
func WithWorker(opts ...wrk.SetOption) SetOption {
	return func(o *Options) {
		o.Worker = opts
	}
}

// WithEmitter sets the options for Kapbit's event emitter.
func WithEmitter(opts ...retry.SetOption) SetOption {
	return func(o *Options) {
		o.Emitter = opts
	}
}

// WithPersistHandler sets the options for Kapbit's persistence handler.
func WithPersistHandler(opts ...persist.SetOption) SetOption {
	return func(o *Options) {
		o.PersistHandler = opts
	}
}

// WithTimeProvider sets the time provider for Kapbit.
func WithTimeProvider(provider TimeProvider) SetOption {
	return func(o *Options) {
		o.TimeProvider = provider
	}
}

// WithShutdownPollInterval sets the duration between checks during shutdown.
func WithShutdownPollInterval(interval time.Duration) SetOption {
	return func(o *Options) {
		o.ShutdownPollInterval = interval
	}
}
