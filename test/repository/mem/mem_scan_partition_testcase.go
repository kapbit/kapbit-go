package mem

import (
	"context"
	"testing"

	evs "github.com/kapbit/kapbit-go/event/store"
	repmem "github.com/kapbit/kapbit-go/repository/mem"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

type ScanPartitionTestCase struct {
	Name           string
	PartitionCount int
	Partition      int
	Records        []evs.EventRecord
	Chpt           int64
	WantRecords    []evs.EventRecord
}

func RunScanPartitionTest(t *testing.T, tc ScanPartitionTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		repo := repmem.New(tc.PartitionCount, 0)
		for _, record := range tc.Records {
			_, err := repo.SaveEvent(context.Background(), tc.Partition, record)
			assertfatal.EqualError(t, err, nil)
		}
		if tc.Chpt >= 0 {
			err := repo.SaveCheckpoint(context.Background(), tc.Partition, tc.Chpt)
			assertfatal.EqualError(t, err, nil)
		}

		var gotRecords []evs.EventRecord
		err := repo.ScanPartition(t.Context(), tc.Partition,
			func(chpt int64) error { return nil },
			func(offset int64, beforeChpt bool, record evs.EventRecord) error {
				gotRecords = append(gotRecords, record)
				return nil
			})

		if tc.Partition < 0 || tc.Partition >= tc.PartitionCount {
			if err == nil {
				t.Error("expected error for invalid partition")
			}
			return
		}

		asserterror.EqualError(t, err, nil)
		asserterror.EqualDeep(t, gotRecords, tc.WantRecords)
	})
}
