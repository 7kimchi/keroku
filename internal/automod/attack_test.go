package automod

import (
	"testing"
	"time"
)

// One million distinct accounts post once each. Memory must stay bounded.
func TestMillionAccountsStayBounded(t *testing.T) {
	v := setup(t, cfg())
	now := v.clk.Now()
	for i := range 1_000_000 {
		st := &userState{}
		spam(st, cfg(), now)
		v.e.state.Set(gs+":"+itoa(int64(i)), st)
	}
	if n := v.e.state.Len(); n > maxUsers {
		t.Fatalf("%d entries", n)
	}
}

func TestStateExpires(t *testing.T) {
	v := setup(t, cfg())
	v.post(spammer, "x")
	v.drain()
	v.clk.Advance(stateTTL + time.Second)
	v.e.Sweep()
	if v.e.state.Len() != 0 {
		t.Fatal("idle state kept")
	}
}

func TestFullQueueDropsInsteadOfBlocking(t *testing.T) {
	v := setup(t, cfg())
	v.drain()
	start := time.Now()
	for range 1000 {
		v.post(spammer, "x")
	}
	if time.Since(start) > time.Second {
		t.Fatal("posting into a closed queue blocked")
	}
}
