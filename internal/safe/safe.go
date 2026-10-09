// Package safe runs functions and goroutines so a panic drops one unit of work, not the process.
package safe

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
)

// Guard recovers panics, logs them with a stack and counts them.
type Guard struct {
	log   *slog.Logger
	count func(where string)
}

// NewGuard builds a guard. count may be nil.
func NewGuard(log *slog.Logger, count func(where string)) *Guard {
	if count == nil {
		count = func(string) {}
	}
	return &Guard{log: log, count: count}
}

// Run calls fn and reports whether it panicked.
func (g *Guard) Run(where string, fn func()) (panicked bool) {
	defer func() {
		if v := recover(); v != nil {
			panicked = true
			g.count(where)
			g.log.Error("panic recovered", "where", where, "panic", fmt.Sprint(v), "stack", string(debug.Stack()))
		}
	}()
	fn()
	return false
}

// Go runs fn on a new goroutine tracked by wg, with panic recovery.
func (g *Guard) Go(wg *sync.WaitGroup, where string, fn func()) {
	wg.Go(func() { g.Run(where, fn) })
}
