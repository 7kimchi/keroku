// Package clock provides a manual clock for tests and time dependent code.
package clock

import (
	"sync"
	"time"
)

// Manual is a clock that only moves when told to.
type Manual struct {
	mu  sync.Mutex
	now time.Time
}

// NewManual starts a manual clock at start.
func NewManual(start time.Time) *Manual {
	return &Manual{now: start}
}

// Now returns the current manual time.
func (m *Manual) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

// Advance moves the clock forward by d.
func (m *Manual) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = m.now.Add(d)
}
