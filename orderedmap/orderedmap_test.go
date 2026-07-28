package orderedmap

import (
	"sync"
	"testing"
)

func TestMapOrderAndUpdate(t *testing.T) {
	m := New[string, int]()
	m.Set("first", 1)
	m.Set("second", 2)
	m.Set("first", 3)

	if got := m.Keys(); len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("Keys() = %#v", got)
	}
	if key, value, ok := m.At(0); !ok || key != "first" || value != 3 {
		t.Fatalf("At(0) = %q, %d, %v", key, value, ok)
	}
	if !m.Delete("first") || m.Delete("missing") {
		t.Fatal("Delete returned an unexpected result")
	}
}

func TestZeroValueAndConcurrentSet(t *testing.T) {
	var m Map[int, int]
	var wait sync.WaitGroup
	for i := 0; i < 100; i++ {
		wait.Add(1)
		go func(value int) {
			defer wait.Done()
			m.Set(value, value)
		}(i)
	}
	wait.Wait()
	if m.Len() != 100 {
		t.Fatalf("Len() = %d, want 100", m.Len())
	}
}
