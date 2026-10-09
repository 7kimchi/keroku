package modlog

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestFloodIsBoundedAndNeverBlocks(t *testing.T) {
	p, f, _, m := setup(t, 10)
	f.SetDelay("send", 50*time.Millisecond)
	start := time.Now()
	accepted := 0
	for i := range 10_000 {
		if p.Case(1, embed(i)) {
			accepted++
		}
	}
	if time.Since(start) > time.Second {
		t.Fatalf("enqueue blocked for %v", time.Since(start))
	}
	if accepted > 22 || testutil.ToFloat64(m.Modlog.WithLabelValues("dropped")) < 9_978 {
		t.Fatalf("accepted %d of 10000 into a queue of 2x10", accepted)
	}
	if p.Depth() > 20 {
		t.Fatalf("depth %d", p.Depth())
	}
}
