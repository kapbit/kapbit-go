package mem_test

import (
	"testing"

	"github.com/kapbit/kapbit-go/test/repository/mem"
)

func TestSaveEvent(t *testing.T) {
	for _, tc := range []mem.SaveEventTestCase{
		mem.SaveEventSuccessCase(),
		mem.SaveEventPayloadTooLargeCase(),
		mem.SaveEventInvalidPartitionCase(),
	} {
		mem.RunSaveEventTest(t, tc)
	}
}

func TestSaveCheckpoint(t *testing.T) {
	for _, tc := range []mem.SaveCheckpointTestCase{
		mem.SaveCheckpointSuccessCase(),
		mem.SaveCheckpointInvalidPartitionCase(),
	} {
		mem.RunSaveCheckpointTest(t, tc)
	}
}

func TestLoadRecent(t *testing.T) {
	for _, tc := range []mem.LoadRecentTestCase{
		mem.LoadRecentFullRecordsCase(),
		mem.LoadRecentLightRecordsCase(),
		mem.LoadRecentStartConditionCase(),
		mem.LoadRecentInvalidPartitionCase(),
	} {
		mem.RunLoadRecentTest(t, tc)
	}
}

func TestScanPartition(t *testing.T) {
	for _, tc := range []mem.ScanPartitionTestCase{
		mem.ScanPartitionAllRecordsCase(),
		mem.ScanPartitionInvalidPartitionCase(),
	} {
		mem.RunScanPartitionTest(t, tc)
	}
}

func TestPartitionCount(t *testing.T) {
	for _, tc := range []mem.PartitionCountTestCase{
		mem.PartitionCountCase(),
	} {
		mem.RunPartitionCountTest(t, tc)
	}
}
