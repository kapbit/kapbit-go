package runtime

import (
	"fmt"
	"log/slog"
)

const (
	DefaultMaxWorkflows          = 100
	DefaultCheckpointSize        = 100
	DefaultIdempotencyWindowSize = 1000
)

type RestorerOptions struct {
	MaxWorkflows     int
	CheckpointSize   int
	IdempotencyDepth int
	Logger           *slog.Logger
}

func DefaultRestorerOptions() RestorerOptions {
	return RestorerOptions{
		MaxWorkflows:     DefaultMaxWorkflows,
		CheckpointSize:   DefaultCheckpointSize,
		IdempotencyDepth: DefaultIdempotencyWindowSize,
		Logger:           slog.Default(),
	}
}

func ApplyRestorer(o *RestorerOptions, opts ...SetOption) error {
	for _, opt := range opts {
		opt(o)
	}
	return o.Validate()
}

// Validate checks for configuration errors and logical inconsistencies.
func (o *RestorerOptions) Validate() error {
	if o.MaxWorkflows <= 0 {
		return fmt.Errorf("option MaxWorkflows must be positive (got %d)",
			o.MaxWorkflows)
	}
	if o.CheckpointSize <= 0 {
		return fmt.Errorf("optiong CheckpointSize must be positive (got %d)",
			o.CheckpointSize)
	}
	if o.IdempotencyDepth <= 0 {
		return fmt.Errorf("optiong IdempotencyWindowSize must be positive (got %d)",
			o.IdempotencyDepth)
	}
	return nil
}

type SetOption func(*RestorerOptions)

func WithMaxWorkflows(n int) SetOption {
	return func(o *RestorerOptions) {
		o.MaxWorkflows = n
	}
}

// WithCheckpointSize sets the checkpoint size.
func WithCheckpointSize(s int) SetOption {
	return func(o *RestorerOptions) {
		o.CheckpointSize = s
	}
}

// WithIdempotencyDepth sets the idempotency window size.
func WithIdempotencyDepth(s int) SetOption {
	return func(o *RestorerOptions) {
		o.IdempotencyDepth = s
	}
}

func WithLogger(logger *slog.Logger) SetOption {
	return func(o *RestorerOptions) {
		o.Logger = logger
	}
}
