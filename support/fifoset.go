package support

import (
	"container/list"
	"fmt"
	"strings"
)

// FIFOSet is a fixed-size set where the oldest item is evicted
// when capacity is exceeded.
type FIFOSet[T comparable] struct {
	capacity int
	ll       *list.List          // keeps insertion order
	m        map[T]*list.Element // lookup element → list node
}

// NewFIFOSet returns a new FIFOSet with the given maximum size.
func NewFIFOSet[T comparable](capacity int) *FIFOSet[T] {
	if capacity <= 0 {
		panic("FIFOSet capacity must be > 0")
	}
	return &FIFOSet[T]{
		capacity: capacity,
		ll:       list.New(),
		m:        make(map[T]*list.Element),
	}
}

// Add inserts an item. If it already exists, it does nothing.
// If capacity is exceeded, the oldest item is removed.
func (s *FIFOSet[T]) Add(v T) {
	// Already exists → do nothing
	if _, ok := s.m[v]; ok {
		return
	}

	// Push to the back (newest)
	elem := s.ll.PushBack(v)
	s.m[v] = elem

	// Remove oldest if over capacity
	if s.ll.Len() > s.capacity {
		oldest := s.ll.Front()
		if oldest != nil {
			val := oldest.Value.(T)
			s.ll.Remove(oldest)
			delete(s.m, val)
		}
	}
}

// Contains returns true if the value exists in the set.
func (s *FIFOSet[T]) Contains(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Remove deletes a value manually.
func (s *FIFOSet[T]) Remove(v T) {
	if elem, ok := s.m[v]; ok {
		s.ll.Remove(elem)
		delete(s.m, v)
	}
}

func (s *FIFOSet[T]) Cap() int {
	return s.capacity
}

// Len returns the current number of elements.
func (s *FIFOSet[T]) Len() int {
	return s.ll.Len()
}

// Values returns elements in FIFO order (oldest first).
func (s *FIFOSet[T]) Values() []T {
	out := make([]T, 0, s.ll.Len())
	for e := s.ll.Front(); e != nil; e = e.Next() {
		out = append(out, e.Value.(T))
	}
	return out
}

func (s *FIFOSet[T]) String() string {
	// Using s.Values() is safe, but ensure it doesn't cause a data race
	// if your FIFOSet is used concurrently (might need a lock here).
	elements := s.Values()
	size := len(elements)

	var sb strings.Builder
	// Follow the "package.Type" pattern you've established
	fmt.Fprintf(&sb, "support.FIFOSet cap=%d len=%d [", s.capacity, size)

	for i, v := range elements {
		if i > 0 {
			sb.WriteString(" ") // Use space instead of comma for consistency
		}
		fmt.Fprintf(&sb, "%v", v)
	}
	sb.WriteString("]")
	return sb.String()
}

// // String returns a string representation of the FIFOSet, including its capacity
// // and its elements listed in FIFO (oldest to newest) order.
// func (s *FIFOSet[T]) String() string {
// 	var sb strings.Builder
// 	sb.WriteString(fmt.Sprintf("FIFOSet (Capacity: %d, Size: %d) [", s.capacity, s.ll.Len()))
//
// 	elements := s.Values()
// 	for i, v := range elements {
// 		sb.WriteString(fmt.Sprintf("%v", v))
// 		if i < len(elements)-1 {
// 			sb.WriteString(", ")
// 		}
// 	}
// 	sb.WriteString("]")
// 	return sb.String()
// }
