package cache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestAddIsExclusive(t *testing.T) {
	c, _ := newTest(t, 100, time.Hour)
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for range 2000 {
		wg.Go(func() {
			if c.Add("only", 1) {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("%d callers won Add, want 1", wins)
	}
}

func TestUpdateHasNoLostWrites(t *testing.T) {
	c, _ := newTest(t, 100, time.Hour)
	var wg sync.WaitGroup
	for range 5000 {
		wg.Go(func() {
			c.Update("n", func(old int, _ bool) int { return old + 1 })
		})
	}
	wg.Wait()
	if v, _ := c.Get("n"); v != 5000 {
		t.Fatalf("got %d, want 5000", v)
	}
}

func TestUpdateSeesFoundFlag(t *testing.T) {
	c, _ := newTest(t, 10, time.Hour)
	if got := c.Update("a", func(old int, found bool) int {
		if found {
			t.Fatal("found on empty cache")
		}
		return 1
	}); got != 1 {
		t.Fatalf("got %d", got)
	}
	c.Update("a", func(old int, found bool) int {
		if !found || old != 1 {
			t.Fatalf("found %v old %d", found, old)
		}
		return 2
	})
}

func TestParallelMixedLoad(t *testing.T) {
	c, _ := newTest(t, 500, time.Hour)
	var wg sync.WaitGroup
	for w := range 64 {
		wg.Go(func() {
			for i := range 2000 {
				k := strconv.Itoa((w*7919 + i) % 3000)
				switch i % 4 {
				case 0:
					c.Set(k, i)
				case 1:
					c.Get(k)
				case 2:
					c.Add(k, i)
				default:
					c.Delete(k)
				}
			}
			c.Sweep(10)
		})
	}
	wg.Wait()
	if c.Len() > 500 {
		t.Fatalf("len %d over bound", c.Len())
	}
}
