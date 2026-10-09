package cases

import (
	"testing"
	"time"
)

func TestRecentAndByKey(t *testing.T) {
	pool := db(t)
	n := newCase(1, 2, Ban)
	n.IdempotencyKey = "k"
	c, _ := insert(t, pool, n)
	if got, ok, err := Recent(t.Context(), pool, 1, 2, Ban, 0, time.Now().Add(-10*time.Second)); !ok || err != nil || got.ID != c.ID {
		t.Fatalf("recent: %v %v", ok, err)
	}
	if _, ok, _ := Recent(t.Context(), pool, 1, 2, Kick, 0, time.Now().Add(-time.Minute)); ok {
		t.Fatal("wrong kind matched")
	}
	if _, ok, _ := Recent(t.Context(), pool, 1, 2, Ban, 0, time.Now().Add(time.Minute)); ok {
		t.Fatal("old case matched")
	}
	if _, ok, _ := Recent(t.Context(), pool, 1, 2, Ban, 8, time.Now().Add(-time.Minute)); ok {
		t.Fatal("other moderator matched")
	}
	if _, ok, _ := Recent(t.Context(), pool, 1, 2, Ban, 7, time.Now().Add(-time.Minute)); !ok {
		t.Fatal("same moderator missed")
	}
	if got, ok, err := ByKey(t.Context(), pool, 1, "k"); !ok || err != nil || got.ID != c.ID {
		t.Fatal("by key")
	}
	if _, ok, _ := ByKey(t.Context(), pool, 1, "other"); ok {
		t.Fatal("unknown key matched")
	}
}
