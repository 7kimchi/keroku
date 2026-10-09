package cases

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
)

func TestParallelInsertsHaveNoGapsOrDuplicates(t *testing.T) {
	pool := db(t)
	const workers, each = 40, 25
	var wg sync.WaitGroup
	numbers := make(chan int64, workers*each)
	for w := range workers {
		wg.Go(func() {
			for range each {
				c, err := insert(t, pool, newCase(1, int64(w+1), Warn))
				if err != nil {
					t.Error(err)
					return
				}
				numbers <- c.Number
			}
		})
	}
	wg.Wait()
	close(numbers)
	var got []int64
	for n := range numbers {
		got = append(got, n)
	}
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if len(got) != workers*each {
		t.Fatalf("%d cases", len(got))
	}
	for i, n := range got {
		if n != int64(i+1) {
			t.Fatalf("position %d has number %d", i, n)
		}
	}
}

func TestReplayedInteractionCreatesExactlyOneCase(t *testing.T) {
	pool := db(t)
	var ok, dup atomic.Int64
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			n := newCase(1, 2, Ban)
			n.InteractionID = 777
			_, err := insert(t, pool, n)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, ErrDuplicate):
				dup.Add(1)
			default:
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if ok.Load() != 1 || dup.Load() != 49 {
		t.Fatalf("ok %d dup %d", ok.Load(), dup.Load())
	}
	if c, _ := insert(t, pool, newCase(1, 3, Ban)); c.Number != 2 {
		t.Fatalf("next number %d, want 2", c.Number)
	}
}

func TestManyGuildsInParallel(t *testing.T) {
	pool := db(t)
	var wg sync.WaitGroup
	for g := range 30 {
		for range 10 {
			wg.Go(func() {
				if _, err := insert(t, pool, newCase(int64(g+1), 9, Note)); err != nil {
					t.Error(err)
				}
			})
		}
	}
	wg.Wait()
	for g := int64(1); g <= 30; g++ {
		if _, err := Get(t.Context(), pool, g, 10); err != nil {
			t.Fatalf("guild %d missing case 10: %v", g, err)
		}
		if _, err := Get(t.Context(), pool, g, 11); !errors.Is(err, ErrNotFound) {
			t.Fatalf("guild %d has case 11", g)
		}
	}
}
