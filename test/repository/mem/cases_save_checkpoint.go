package mem

import (
	repmem "github.com/kapbit/kapbit-go/repository/mem"
)

func SaveCheckpointSuccessCase() SaveCheckpointTestCase {
	return SaveCheckpointTestCase{
		Name:           "Should successfully save a checkpoint",
		PartitionCount: 1,
		Partition:      0,
		Chpt:           42,
	}
}

func SaveCheckpointInvalidPartitionCase() SaveCheckpointTestCase {
	return SaveCheckpointTestCase{
		Name:           "Should return error for invalid partition",
		PartitionCount: 1,
		Partition:      5,
		Chpt:           0,
		WantError:      repmem.NewInvalidPartitionError(5, 1),
	}
}
