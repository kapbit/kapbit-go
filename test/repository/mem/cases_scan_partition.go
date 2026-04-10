package mem

import (
	evs "github.com/kapbit/kapbit-go/event/store"
)

func ScanPartitionAllRecordsCase() ScanPartitionTestCase {
	records := []evs.EventRecord{
		{Key: "r1", Payload: []byte("p1")},
		{Key: "r2", Payload: []byte("p2")},
	}
	return ScanPartitionTestCase{
		Name:           "Should return all records with payloads regardless of checkpoint",
		PartitionCount: 1,
		Partition:      0,
		Records:        records,
		Chpt:           0,
		WantRecords:    records, // ScanPartition always returns full records
	}
}

func ScanPartitionInvalidPartitionCase() ScanPartitionTestCase {
	return ScanPartitionTestCase{
		Name:           "Should return error for invalid partition",
		PartitionCount: 1,
		Partition:      5,
		Chpt:           -1,
		WantRecords:    []evs.EventRecord{},
	}
}
