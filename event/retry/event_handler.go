package retry

import (
	"context"

	evnt "github.com/kapbit/kapbit-go/event"
)

type EventHandler interface {
	// Handle handles an event.
	//
	// Return EncodeErorr or PersistenceError if any.
	Handle(ctx context.Context, event evnt.Event) error
}
