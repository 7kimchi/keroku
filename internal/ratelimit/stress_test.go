package ratelimit

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrentSameKeyNeverOverAdmits(t *testing.T) {
	l, _ := newTest(t, time.Minute, 10, 100)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for range 5000 {
		wg.Go(func() {
			if ok, _ := l.Allow("hot"); ok {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 10 {
		t.Fatalf("admitted %d, want exactly 10", allowed.Load())
	}
}

func TestThousandsOfUsersEachGetTheirBurst(t *testing.T) {
	l, _ := newTest(t, time.Minute, 3, 10_000)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for u := range 5000 {
		wg.Go(func() {
			for range 5 {
				if ok, _ := l.Allow("user" + strconv.Itoa(u)); ok {
					allowed.Add(1)
				}
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 15000 {
		t.Fatalf("admitted %d, want 15000", allowed.Load())
	}
}

func TestMillionHostileKeysStayBounded(t *testing.T) {
	l, _ := newTest(t, time.Second, 1, 5000)
	for i := range 1_000_000 {
		l.Allow(strconv.Itoa(i))
	}
	if l.Len() > 5000 {
		t.Fatalf("tracking %d keys", l.Len())
	}
}

func TestSpreadOverTimeHoldsRate(t *testing.T) {
	// 50 users send one request every 100ms for five simulated minutes against 1 per second.
	l, clk := newTest(t, time.Second, 1, 1000)
	allowed := 0
	for range 3000 {
		for u := range 50 {
			if ok, _ := l.Allow(strconv.Itoa(u)); ok {
				allowed++
			}
		}
		clk.Advance(100 * time.Millisecond)
	}
	if want := 50 * 300; allowed != want {
		t.Fatalf("admitted %d over 5 minutes, want %d", allowed, want)
	}
}

func FuzzAllow(f *testing.F) {
	f.Add(uint16(3), []byte{1, 0, 2, 255})
	f.Fuzz(func(t *testing.T, burst uint16, steps []byte) {
		b := int(burst%20) + 1
		l, clk := newTest(t, time.Second, b, 10)
		admittedInWindow := 0
		windowStart := clk.Now()
		for _, s := range steps {
			clk.Advance(time.Duration(s) * 10 * time.Millisecond)
			if clk.Now().Sub(windowStart) >= time.Duration(b)*time.Second {
				windowStart, admittedInWindow = clk.Now(), 0
			}
			if ok, wait := l.Allow("k"); ok {
				admittedInWindow++
			} else if wait <= 0 || wait > time.Duration(b)*time.Second {
				t.Fatalf("bad wait %v", wait)
			}
			if admittedInWindow > 2*b {
				t.Fatalf("admitted %d in one window with burst %d", admittedInWindow, b)
			}
		}
	})
}
