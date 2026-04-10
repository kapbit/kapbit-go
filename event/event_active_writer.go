package event

type ActiveWriterEvent struct {
	NodeID         string
	PartitionIndex int
	Timestamp      int64
}

func (e ActiveWriterEvent) Partition() int {
	return e.PartitionIndex
}

func (e ActiveWriterEvent) isSystem() {}
