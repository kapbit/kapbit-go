package runtime

import "fmt"

type Position struct {
	Partition      int
	LogicalSeq     int64
	PhysicalOffset int64
}

func (p Position) String() string {
	return fmt.Sprintf("{%d, %d, %d}", p.Partition, p.LogicalSeq,
		p.PhysicalOffset)
}
