package mem

import (
	"testing"

	evs "github.com/kapbit/kapbit-go/event/store"
	repmem "github.com/kapbit/kapbit-go/repository/mem"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

type LoadRecentTestCase struct {
	Name             string
	PartitionCount   int
	PayloadSizeLimit int
	Partition        int
	Records          []evs.EventRecord
	Chpt             int64
	StartCondition   evs.StartCondition
	WantRecords      []evs.EventRecord
	WantChpt         int64
	WantError        error
}

func RunLoadRecentTest(t *testing.T, tc LoadRecentTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		s := repmem.New(tc.PartitionCount, tc.PayloadSizeLimit)
		var err error
		for _, r := range tc.Records {
			_, err = s.SaveEvent(t.Context(), tc.Partition, r)
			assertfatal.EqualError(t, err, nil)
		}
		if tc.WantError == nil {
			err = s.SaveCheckpoint(t.Context(), tc.Partition, tc.Chpt)
			assertfatal.EqualError(t, err, nil)
		}

		scanned := []evs.EventRecord{}
		err = s.LoadRecent(t.Context(), tc.Partition, tc.StartCondition,
			func(chpt int64) error {
				asserterror.Equal(t, chpt, tc.WantChpt)
				return nil
			},
			func(offset int64, beforeChpt bool, record evs.EventRecord) error {
				scanned = append(scanned, record)
				return nil
			},
		)
		asserterror.EqualError(t, err, tc.WantError)
		if err == nil {
			asserterror.EqualDeep(t, scanned, tc.WantRecords)
		}
	})
}
