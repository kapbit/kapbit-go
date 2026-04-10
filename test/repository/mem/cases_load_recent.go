package mem

import (
	evs "github.com/kapbit/kapbit-go/event/store"
	repmem "github.com/kapbit/kapbit-go/repository/mem"
)

func LoadRecentFullRecordsCase() LoadRecentTestCase {
	records := []evs.EventRecord{
		{Key: "r1", Payload: []byte("p1")},
		{Key: "r2", Payload: []byte("p2")},
	}
	return LoadRecentTestCase{
		Name:           "Should return full records when there is no checkpoint",
		PartitionCount: 1,
		Partition:      0,
		Records:        records,
		Chpt:           -1,
		WantChpt:       -1,
		WantRecords:    records,
	}
}

func LoadRecentLightRecordsCase() LoadRecentTestCase {
	return LoadRecentTestCase{
		Name:           "Should return light records for events before checkpoint",
		PartitionCount: 1,
		Partition:      0,
		Records: []evs.EventRecord{
			{Key: "r1", Payload: []byte("p1")},
			{Key: "r2", Payload: []byte("p2")},
			{Key: "r3", Payload: []byte("p3")},
		},
		Chpt:     1, // r1 (index 0) and r2 (index 1) are before/at checkpoint
		WantChpt: 1,
		WantRecords: []evs.EventRecord{
			{Key: "r1"},                        // light: only key is kept
			{Key: "r2"},                        // light: only key is kept
			{Key: "r3", Payload: []byte("p3")}, // full: after checkpoint
		},
	}
}

func LoadRecentStartConditionCase() LoadRecentTestCase {
	records := []evs.EventRecord{
		{Key: "r1", Payload: []byte("p1")},
		{Key: "r2", Payload: []byte("p2")},
		{Key: "r3", Payload: []byte("p3")},
	}
	// Start condition: find the record with key "r2"
	cond := func(r evs.EventRecord) (bool, error) {
		return r.Key == "r2", nil
	}
	return LoadRecentTestCase{
		Name:           "Should start from the record matching the start condition",
		PartitionCount: 1,
		Partition:      0,
		Records:        records,
		Chpt:           0, // r1 is at checkpoint; r2 and r3 after
		WantChpt:       0,
		StartCondition: cond,
		WantRecords: []evs.EventRecord{
			{Key: "r2", Payload: []byte("p2")},
			{Key: "r3", Payload: []byte("p3")},
		},
	}
}

func LoadRecentInvalidPartitionCase() LoadRecentTestCase {
	return LoadRecentTestCase{
		Name:           "Should return error for invalid partition",
		PartitionCount: 1,
		Partition:      5,
		Chpt:           -1,
		WantChpt:       0,
		WantRecords:    []evs.EventRecord{},
		WantError:      repmem.NewInvalidPartitionError(5, 1),
	}
}
