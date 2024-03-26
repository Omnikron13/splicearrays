package tree

import "testing"

func TestMinimalSlab_Add(t *testing.T) {
	s := MinimalSlab[int]{}
	start, length := s.Add(1, 2, 3)
	if start != 0 {
		t.Error("Expected start 0, got", start)
	}
	if length != 3 {
		t.Error("Expected length 3, got", length)
	}
}

func TestMinimalSlab_Get(t *testing.T) {
	s := MinimalSlab[int]{}
	s.Add(1, 2, 3)
	if *s.Get(0) != 1 {
		t.Error("Expected 1, got", s.Get(0))
	}
	if *s.Get(1) != 2 {
		t.Error("Expected 2, got", s.Get(1))
	}
	if *s.Get(2) != 3 {
		t.Error("Expected 3, got", s.Get(2))
	}
}

func TestMinimalSlab_Len(t *testing.T) {
	s := MinimalSlab[int]{}
	s.Add(1, 2, 3)
	if s.Len() != 3 {
		t.Error("Expected 3, got", s.Len())
	}
}

func TestMinimalSlab_SliceIter(t *testing.T) {
	s := MinimalSlab[int]{}
	s.Add(1, 2, 3)
	i := 1
	for x := range s.SliceIter(0, 3) {
		if *x != i {
			t.Error("Expected 1, got", x)
		}
		i++
	}
}
