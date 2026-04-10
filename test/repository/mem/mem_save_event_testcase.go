package mem

import (
	"testing"

	evs "github.com/kapbit/kapbit-go/event/store"
	repmem "github.com/kapbit/kapbit-go/repository/mem"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

type SaveEventTestCase struct {
	Name             string
	PartitionCount   int
	PayloadSizeLimit int
	Partition        int
	Record           evs.EventRecord
	WantOffset       int64
	WantError        error
}

func RunSaveEventTest(t *testing.T, tc SaveEventTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		s := repmem.New(tc.PartitionCount, tc.PayloadSizeLimit)
		offset, err := s.SaveEvent(t.Context(), tc.Partition, tc.Record)
		asserterror.EqualError(t, err, tc.WantError)
		if err == nil {
			assertfatal.Equal(t, offset, tc.WantOffset)
		}
	})
}
