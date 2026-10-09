package cache

import (
	"strconv"
	"testing"
	"time"
)

func TestSizeNeverExceedsMax(t *testing.T) {
	for _, size := range []int{1, 2, 15, 16, 17, 100, 1000} {
		c, _ := newTest(t, size, time.Hour)
		for i := range size * 10 {
			c.Set(strconv.Itoa(i), i)
			if c.Len() > size {
				t.Fatalf("size %d: len %d after %d inserts", size, c.Len(), i+1)
			}
		}
	}
}

func TestEvictsLeastRecentlyUsed(t *testing.T) {
	c, _ := newTest(t, 1, time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	if _, ok := c.Get("a"); ok {
		t.Fatal("oldest entry survived over capacity")
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Fatal("newest entry evicted")
	}
}

func TestReadKeepsEntryWarm(t *testing.T) {
	// 32 entries over 16 shards gives two slots per shard.
	c, _ := newTest(t, 32, time.Hour)
	shard := c.index("k0")
	var keys []string
	for i := 0; len(keys) < 3; i++ {
		k := "k" + strconv.Itoa(i)
		if c.index(k) == shard {
			keys = append(keys, k)
		}
	}
	c.Set(keys[0], 0)
	c.Get(keys[0])
	c.Set(keys[1], 1)
	c.Get(keys[0])
	c.Set(keys[2], 2)
	if _, ok := c.Get(keys[0]); !ok {
		t.Fatal("recently read entry evicted")
	}
	if _, ok := c.Get(keys[1]); ok {
		t.Fatal("cold entry survived")
	}
}

func TestSweepDropsExpired(t *testing.T) {
	c, clk := newTest(t, 1000, time.Minute)
	for i := range 500 {
		c.Set(strconv.Itoa(i), i)
	}
	if n := c.Sweep(1000); n != 0 {
		t.Fatalf("swept %d live entries", n)
	}
	clk.Advance(time.Minute)
	if n := c.Sweep(1000); n != 500 {
		t.Fatalf("swept %d, want 500", n)
	}
	if c.Len() != 0 {
		t.Fatalf("len %d", c.Len())
	}
}

func TestSweepRespectsLimit(t *testing.T) {
	c, clk := newTest(t, 1600, time.Minute)
	for i := range 1600 {
		c.Set(strconv.Itoa(i), i)
	}
	clk.Advance(time.Minute)
	if n := c.Sweep(1); n > 16 {
		t.Fatalf("swept %d with limit 1 per shard", n)
	}
	if n := c.Sweep(0); n != 0 {
		t.Fatalf("limit 0 swept %d", n)
	}
}

func TestMillionHostileKeysStayBounded(t *testing.T) {
	c, _ := newTest(t, 10_000, time.Hour)
	for i := range 1_000_000 {
		c.Set(strconv.Itoa(i), i)
	}
	if c.Len() > 10_000 {
		t.Fatalf("len %d over bound", c.Len())
	}
}
