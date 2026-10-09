package ratelimit

import (
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/clock"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func newTest(t *testing.T, interval time.Duration, burst, keys int) (*Limiter, *clock.Manual) {
	t.Helper()
	clk := clock.NewManual(time.Unix(1_700_000_000, 0))
	l, err := New(interval, burst, keys, clk.Now)
	if err != nil {
		t.Fatal(err)
	}
	return l, clk
}
