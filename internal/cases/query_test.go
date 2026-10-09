package cases

import (
	"errors"
	"testing"
	"time"
)

func TestHistoryAndCounts(t *testing.T) {
	pool := db(t)
	for _, k := range []Kind{Warn, Note, Warn, Timeout, Warn} {
		_, _ = insert(t, pool, newCase(1, 2, k))
	}
	_, _ = insert(t, pool, newCase(1, 3, Warn))
	h, err := History(t.Context(), pool, 1, 2, 10, 0)
	if err != nil || len(h) != 5 || h[0].Number != 5 || h[4].Number != 1 {
		t.Fatalf("history %v %v", h, err)
	}
	page, _ := History(t.Context(), pool, 1, 2, 2, 2)
	if len(page) != 2 || page[0].Number != 3 {
		t.Fatalf("page %v", page)
	}
	if n, err := CountByKind(t.Context(), pool, 1, 2, Warn); err != nil || n != 3 {
		t.Fatalf("count %d %v", n, err)
	}
	if n, err := CountUpTo(t.Context(), pool, 1, 2, Warn, 3); err != nil || n != 2 {
		t.Fatalf("count up to 3: %d %v", n, err)
	}
	if n, _ := CountUpTo(t.Context(), pool, 2, 2, Warn, 100); n != 0 {
		t.Fatal("count up to across guilds")
	}
	w, _ := ListByKind(t.Context(), pool, 1, 2, Warn, 2)
	if len(w) != 2 || w[0].Number != 5 {
		t.Fatalf("warnings %v", w)
	}
	if empty, _ := History(t.Context(), pool, 1, 99, 10, 0); len(empty) != 0 {
		t.Fatal("history for unknown target")
	}
}

// Every read takes the guild id. A case from guild 1 must be invisible from guild 2.
func TestGuildIsolation(t *testing.T) {
	pool := db(t)
	n := newCase(1, 2, Warn)
	n.IdempotencyKey = "secret"
	c, _ := insert(t, pool, n)
	if _, err := Get(t.Context(), pool, 2, c.Number); !errors.Is(err, ErrNotFound) {
		t.Fatal("case read across guilds")
	}
	if h, _ := History(t.Context(), pool, 2, 2, 10, 0); len(h) != 0 {
		t.Fatal("history across guilds")
	}
	if n, _ := CountByKind(t.Context(), pool, 2, 2, Warn); n != 0 {
		t.Fatal("count across guilds")
	}
	if w, _ := ListByKind(t.Context(), pool, 2, 2, Warn, 10); len(w) != 0 {
		t.Fatal("warnings across guilds")
	}
	if _, ok, _ := Recent(t.Context(), pool, 2, 2, Warn, 0, time.Time{}); ok {
		t.Fatal("recent across guilds")
	}
	if _, ok, _ := ByKey(t.Context(), pool, 2, "secret"); ok {
		t.Fatal("key across guilds")
	}
	if _, err := editOnce(t, pool, 2, c.Number, "hijack", 0); !errors.Is(err, ErrNotFound) {
		t.Fatal("reason edited across guilds")
	}
	if got, _ := Get(t.Context(), pool, 1, c.Number); got.Reason != "r" {
		t.Fatal("cross guild edit changed the case")
	}
	if e, _ := Edits(t.Context(), pool, 2, c.ID); len(e) != 0 {
		t.Fatal("edit history across guilds")
	}
}
