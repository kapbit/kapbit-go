package mem

import (
	"time"

	evs "github.com/kapbit/kapbit-go/event/store"
)

// StoredRecord wraps the event with its assigned offset and ingestion time.
type StoredRecord struct {
	Offset    int64
	Record    evs.EventRecord
	CreatedAt time.Time
}
