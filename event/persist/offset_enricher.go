package persist

// OffsetEnricher is the specific contract for events that require
// physical log positioning data, typically for archival or tracking.
type OffsetEnricher interface {
	SetWorkflowOffset(offset int64)
}
