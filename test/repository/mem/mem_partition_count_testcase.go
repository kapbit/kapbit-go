package mem

import (
	"testing"

	repmem "github.com/kapbit/kapbit-go/repository/mem"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type PartitionCountTestCase struct {
	Name           string
	PartitionCount int
}

func RunPartitionCountTest(t *testing.T, tc PartitionCountTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		s := repmem.New(tc.PartitionCount, 0)
		asserterror.Equal(t, s.PartitionCount(), tc.PartitionCount)
	})
}
