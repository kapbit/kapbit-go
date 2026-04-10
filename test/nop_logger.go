package test

import (
	"io"
	"log/slog"
)

var NopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
