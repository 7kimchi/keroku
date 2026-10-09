package commands

import (
	"strconv"
	"sync"
	"testing"
)

// Many moderators on many targets at once: every command answers exactly once.
func TestDispatchTargetFlood(t *testing.T) {
	tr := &tracker{order: map[string][]string{}}
	h := newHarness(t, 100000, tr.cmd("act"))
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Go(func() {
			for i := range 50 {
				send(h, "act", "1000000000000"+strconv.Itoa(10000+g*100+i%7), i)
			}
		})
	}
	wg.Wait()
	h.drain()
	total := 0
	for _, v := range tr.order {
		total += len(v)
	}
	if total != 1000 {
		t.Fatalf("handled %d of 1000", total)
	}
}
