package mem

import (
	"testing"

	repmem "github.com/kapbit/kapbit-go/repository/mem"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type SaveCheckpointTestCase struct {
	Name           string
	PartitionCount int
	Partition      int
	Chpt           int64
	WantError      error
}

func RunSaveCheckpointTest(t *testing.T, tc SaveCheckpointTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		s := repmem.New(tc.PartitionCount, 0)
		err := s.SaveCheckpoint(t.Context(), tc.Partition, tc.Chpt)
		asserterror.EqualError(t, err, tc.WantError)
	})
}
