package handler

import "sync"

type OffsetSupplier struct {
	offsets []int64
	mu      sync.Mutex
}

func NewOffsetSupplier(partitions int) *OffsetSupplier {
	return &OffsetSupplier{
		offsets: make([]int64, partitions),
	}
}

func (s *OffsetSupplier) Next(partition int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	offset := s.offsets[partition]
	s.offsets[partition]++
	return offset
}
