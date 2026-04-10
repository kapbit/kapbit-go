package mem

import (
	"github.com/kapbit/kapbit-go/codes"
	evs "github.com/kapbit/kapbit-go/event/store"
	repmem "github.com/kapbit/kapbit-go/repository/mem"
)

func SaveEventSuccessCase() SaveEventTestCase {
	return SaveEventTestCase{
		Name:           "Should successfully save a record",
		PartitionCount: 1,
		Partition:      0,
		Record:         evs.EventRecord{Key: "k1", Payload: []byte("p1")},
		WantOffset:     0,
	}
}

func SaveEventPayloadTooLargeCase() SaveEventTestCase {
	return SaveEventTestCase{
		Name:             "Should return error when payload is too large",
		PartitionCount:   1,
		PayloadSizeLimit: 10,
		Partition:        0,
		Record:           evs.EventRecord{Key: "k1", Payload: make([]byte, 11)},
		WantError: codes.NewPersistenceError(
			repmem.NewTooLargePayloadError(11),
			codes.PersistenceKindRejection,
		),
	}
}

func SaveEventInvalidPartitionCase() SaveEventTestCase {
	return SaveEventTestCase{
		Name:           "Should return error for invalid partition",
		PartitionCount: 1,
		Partition:      5,
		Record:         evs.EventRecord{Key: "k1"},
		WantError:      repmem.NewInvalidPartitionError(5, 1),
	}
}
