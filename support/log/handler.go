package log

import (
	"context"
	"log/slog"
)

var uniqueKeys = map[string]struct{}{
	"app":       {},
	"comp":      {},
	"mod":       {},
	"node-id":   {},
	"wid":       {},
	"partition": {},
	"txn-id":    {},
}

// UniqueHandler ensures that specific keys are unique in the log output.
// It keeps only the last value provided for these keys.
type UniqueHandler struct {
	handler slog.Handler
	unique  map[string]slog.Attr
}

func NewUniqueHandler(h slog.Handler) slog.Handler {
	return &UniqueHandler{
		handler: h,
		unique:  make(map[string]slog.Attr),
	}
}

func (h *UniqueHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *UniqueHandler) Handle(ctx context.Context, r slog.Record) error {
	// If the record has any unique keys, we need to handle them.
	// However, we want the ones in h.unique to take precedence over
	// previous ones, but ones in the Record itself usually represent
	// the most specific context.

	// Actually, the simplest way is to just add h.unique attributes to the record.
	// Most slog handlers (like JSON) will show all of them if they are duplicated,
	// but UniqueHandler's goal is to prevent them from being duplicated in the first place
	// by intercepting WithAttrs.

	// To handle attributes passed directly to the log call:
	var finalUnique = make(map[string]slog.Attr)
	for k, v := range h.unique {
		finalUnique[k] = v
	}

	r.Attrs(func(a slog.Attr) bool {
		if _, ok := uniqueKeys[a.Key]; ok {
			finalUnique[a.Key] = a
		}
		return true
	})

	// Create a new record with filtered attributes.
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	// Add all non-unique attributes from the original record.
	r.Attrs(func(a slog.Attr) bool {
		if _, ok := uniqueKeys[a.Key]; !ok {
			newRecord.AddAttrs(a)
		}
		return true
	})

	// Add the final unique attributes.
	for _, a := range finalUnique {
		newRecord.AddAttrs(a)
	}

	return h.handler.Handle(ctx, newRecord)
}

func (h *UniqueHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newUnique := make(map[string]slog.Attr)
	for k, v := range h.unique {
		newUnique[k] = v
	}

	var other []slog.Attr
	for _, a := range attrs {
		if _, ok := uniqueKeys[a.Key]; ok {
			newUnique[a.Key] = a
		} else {
			other = append(other, a)
		}
	}

	return &UniqueHandler{
		handler: h.handler.WithAttrs(other),
		unique:  newUnique,
	}
}

func (h *UniqueHandler) WithGroup(name string) slog.Handler {
	// Groups make it complicated. For simplicity, we just pass it through.
	return &UniqueHandler{
		handler: h.handler.WithGroup(name),
		unique:  h.unique,
	}
}
