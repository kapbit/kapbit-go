package persist

type WorkflowPosition struct {
	LogicalSeq     int64
	Partition      int
	PhysicalOffset int64
}
