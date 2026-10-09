package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

func TestCached(t *testing.T) {
	loads := 0
	fail := false
	c, err := store.NewCached(func(_ context.Context, g int64) (int64, error) {
		loads++
		if fail {
			return 0, errors.New("db down")
		}
		return g * 10, nil
	}, 10, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.Get(t.Context(), 2); v != 20 {
		t.Fatal(v)
	}
	if v, _ := c.Get(t.Context(), 2); v != 20 || loads != 1 {
		t.Fatalf("not cached: %d loads", loads)
	}
	c.Invalidate(2)
	fail = true
	if _, err := c.Get(t.Context(), 2); err == nil {
		t.Fatal("load error swallowed")
	}
	fail = false
	if v, _ := c.Get(t.Context(), 2); v != 20 || loads != 3 {
		t.Fatalf("error was cached: %d loads", loads)
	}
	c.Sweep()
	if _, err := store.NewCached(func(context.Context, int64) (int, error) { return 0, nil }, 0, time.Hour); err == nil {
		t.Fatal("zero size accepted")
	}
}
