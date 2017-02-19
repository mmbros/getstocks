package main

import "testing"

func TestSessions(t *testing.T) {
	var s Sessions

	L := s.Length()
	if L != 0 {
		t.Errorf("nil Sessions - Length == %d, expecting %d", L, 0)
	}

	i := s.Item(0)
	if i != nil {
		t.Errorf("nil Sessions - Item(0) == %v, expecting %v", i, nil)
	}
}
