package cache

import (
	"testing"
	"time"

	"github.com/keroku/keroku/internal/clock"
)

func newTest(t *testing.T, size int, ttl time.Duration) (*Cache[string, int], *clock.Manual) {
	t.Helper()
	clk := clock.NewManual(time.Unix(1_700_000_000, 0))
	c, err := New[string, int](size, ttl, clk.Now)
	if err != nil {
		t.Fatal(err)
	}
	return c, clk
}

func TestNewRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct {
		size int
		ttl  time.Duration
	}{{0, time.Second}, {-1, time.Second}, {1, 0}, {1, -time.Second}} {
		if _, err := New[string, int](tc.size, tc.ttl, nil); err == nil {
			t.Fatalf("size %d ttl %v: want error", tc.size, tc.ttl)
		}
	}
}

func TestNewDefaultsToWallClock(t *testing.T) {
	c, err := New[string, int](4, time.Hour, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Set("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("got %v %v", v, ok)
	}
}

func TestSetGet(t *testing.T) {
	c, _ := newTest(t, 10, time.Minute)
	if _, ok := c.Get("missing"); ok {
		t.Fatal("empty cache returned a value")
	}
	c.Set("a", 1)
	c.Set("a", 2)
	if v, ok := c.Get("a"); !ok || v != 2 {
		t.Fatalf("got %v %v", v, ok)
	}
	if c.Len() != 1 {
		t.Fatalf("len %d", c.Len())
	}
}

func TestZeroValueKeysAndUnicode(t *testing.T) {
	c, _ := newTest(t, 10, time.Minute)
	c.Set("", 7)
	c.Set("\U000000E9\U00004E16\U0001F600", 8)
	if v, _ := c.Get(""); v != 7 {
		t.Fatalf("empty key got %d", v)
	}
	if v, _ := c.Get("\U000000E9\U00004E16\U0001F600"); v != 8 {
		t.Fatalf("unicode key got %d", v)
	}
}

func TestDelete(t *testing.T) {
	c, _ := newTest(t, 10, time.Minute)
	c.Set("a", 1)
	if !c.Delete("a") {
		t.Fatal("delete of present key reported false")
	}
	if c.Delete("a") {
		t.Fatal("delete of absent key reported true")
	}
	if _, ok := c.Get("a"); ok {
		t.Fatal("deleted key still readable")
	}
}
