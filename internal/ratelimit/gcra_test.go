package ratelimit

import (
	"testing"
	"time"
)

func TestNewRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct {
		i    time.Duration
		b, k int
	}{{0, 1, 1}, {-time.Second, 1, 1}, {time.Second, 0, 1}, {time.Second, 1, 0}} {
		if _, err := New(tc.i, tc.b, tc.k, nil); err == nil {
			t.Fatalf("%+v accepted", tc)
		}
	}
	if _, err := New(time.Second, 1, 1, nil); err != nil {
		t.Fatal(err)
	}
}

func TestBurstThenSteadyRate(t *testing.T) {
	l, clk := newTest(t, time.Second, 3, 100)
	for i := range 3 {
		if ok, _ := l.Allow("u"); !ok {
			t.Fatalf("burst request %d refused", i)
		}
	}
	ok, wait := l.Allow("u")
	if ok || wait != time.Second {
		t.Fatalf("4th: ok %v wait %v", ok, wait)
	}
	clk.Advance(999 * time.Millisecond)
	if ok, wait := l.Allow("u"); ok || wait != time.Millisecond {
		t.Fatalf("early: ok %v wait %v", ok, wait)
	}
	clk.Advance(time.Millisecond)
	if ok, _ := l.Allow("u"); !ok {
		t.Fatal("token not refilled on time")
	}
	if ok, _ := l.Allow("u"); ok {
		t.Fatal("only one token should refill per interval")
	}
}

func TestRefusalsDoNotCost(t *testing.T) {
	l, clk := newTest(t, time.Second, 1, 100)
	l.Allow("u")
	for range 1000 {
		l.Allow("u")
	}
	clk.Advance(time.Second)
	if ok, _ := l.Allow("u"); !ok {
		t.Fatal("spamming while limited pushed the window out")
	}
}

func TestIdleRestoresFullBurst(t *testing.T) {
	l, clk := newTest(t, time.Second, 5, 100)
	for range 5 {
		l.Allow("u")
	}
	clk.Advance(time.Hour)
	for i := range 5 {
		if ok, _ := l.Allow("u"); !ok {
			t.Fatalf("request %d after idle refused", i)
		}
	}
	if ok, _ := l.Allow("u"); ok {
		t.Fatal("burst exceeded after idle")
	}
}

func TestKeysAreIndependent(t *testing.T) {
	l, _ := newTest(t, time.Minute, 1, 100)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("a")
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("b limited by a")
	}
	if ok, _ := l.Allow(""); !ok {
		t.Fatal("empty key")
	}
}

func TestSweepDropsIdleKeys(t *testing.T) {
	l, clk := newTest(t, time.Second, 2, 100)
	l.Allow("a")
	l.Allow("b")
	clk.Advance(2 * time.Second)
	if n := l.Sweep(100); n != 2 || l.Len() != 0 {
		t.Fatalf("swept %d, len %d", n, l.Len())
	}
}
