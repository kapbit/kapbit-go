package support

import (
	"testing"
)

func TestFIFOSetBasic(t *testing.T) {
	s := NewFIFOSet[int](3)

	s.Add(1)
	s.Add(2)
	s.Add(3)

	if s.Len() != 3 {
		t.Fatalf("expected length 3, got %d", s.Len())
	}

	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		t.Fatalf("expected elements 1,2,3 to exist")
	}

	values := s.Values()
	expected := []int{1, 2, 3}
	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}

func TestFIFOSetDrop(t *testing.T) {
	s := NewFIFOSet[int](3)

	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Add(4)

	if s.Contains(1) {
		t.Fatalf("expected 1 to be dropped")
	}
	if !s.Contains(4) {
		t.Fatalf("expected 4 to exist")
	}

	values := s.Values()
	expected := []int{2, 3, 4}
	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}

func TestFIFOSetDuplicateInsert(t *testing.T) {
	s := NewFIFOSet[int](3)

	s.Add(1)
	s.Add(2)
	s.Add(2) // duplicate, should not reinsert

	if s.Len() != 2 {
		t.Fatalf("expected length 2, got %d", s.Len())
	}

	values := s.Values()
	expected := []int{1, 2}
	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}

func TestFIFOSetRemove(t *testing.T) {
	s := NewFIFOSet[int](3)

	s.Add(1)
	s.Add(2)
	s.Add(3)

	s.Remove(2)

	if s.Contains(2) {
		t.Fatalf("expected 2 to be removed")
	}

	values := s.Values()
	expected := []int{1, 3}

	if len(values) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(values))
	}

	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, values)
		}
	}
}

func TestFIFOSetDropAfterRemove(t *testing.T) {
	s := NewFIFOSet[int](3)

	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Remove(1) // set contains [2,3]
	s.Add(4)    // set now [2,3,4]
	s.Add(5)    // drop 2 → [3,4,5]

	values := s.Values()
	expected := []int{3, 4, 5}

	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}

func TestFIFOSetOrderAfterRepeatedEvictions(t *testing.T) {
	s := NewFIFOSet[int](3)

	for i := 1; i <= 10; i++ {
		s.Add(i)
	}

	// Only the last 3 should remain
	values := s.Values()
	expected := []int{8, 9, 10}

	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}
