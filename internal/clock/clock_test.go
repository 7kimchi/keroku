package clock

import (
	"sync"
	"testing"
	"time"
)

func TestManualStartsAtStart(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := NewManual(start).Now(); !got.Equal(start) {
		t.Fatalf("got %v want %v", got, start)
	}
}

func TestManualAdvance(t *testing.T) {
	start := time.Unix(1000, 0)
	m := NewManual(start)
	m.Advance(0)
	m.Advance(1500 * time.Millisecond)
	if got := m.Now(); !got.Equal(start.Add(1500 * time.Millisecond)) {
		t.Fatalf("got %v", got)
	}
}

func TestManualConcurrentAdvance(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewManual(start)
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			m.Advance(time.Millisecond)
			_ = m.Now()
		})
	}
	wg.Wait()
	if got := m.Now(); !got.Equal(start.Add(time.Second)) {
		t.Fatalf("lost updates: got %v", got)
	}
}
