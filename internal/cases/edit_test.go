package cases

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func editOnce(t *testing.T, pool *pgxpool.Pool, guild, number int64, reason string, interaction int64) (Case, error) {
	t.Helper()
	var c Case
	err := pgx.BeginFunc(context.Background(), pool, func(tx pgx.Tx) error {
		var err error
		c, err = EditReason(context.Background(), tx, guild, number, 8, reason, interaction)
		return err
	})
	return c, err
}

func TestEditReasonKeepsHistory(t *testing.T) {
	pool := db(t)
	c, _ := insert(t, pool, newCase(1, 2, Ban))
	if got, err := editOnce(t, pool, 1, c.Number, "second", 0); err != nil || got.Reason != "second" {
		t.Fatalf("%v %v", got, err)
	}
	_, _ = editOnce(t, pool, 1, c.Number, "third", 0)
	edits, err := Edits(t.Context(), pool, 1, c.ID)
	if err != nil || len(edits) != 2 || edits[0].OldReason != "r" || edits[1].NewReason != "third" || edits[0].EditorID != 8 {
		t.Fatalf("edits %+v %v", edits, err)
	}
	if _, err := editOnce(t, pool, 1, 999, "x", 0); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing case edited")
	}
}

func TestEditReplayIsRejected(t *testing.T) {
	pool := db(t)
	c, _ := insert(t, pool, newCase(1, 2, Ban))
	if _, err := editOnce(t, pool, 1, c.Number, "a", 55); err != nil {
		t.Fatal(err)
	}
	if _, err := editOnce(t, pool, 1, c.Number, "b", 55); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("replayed edit: %v", err)
	}
	if got, _ := Get(t.Context(), pool, 1, c.Number); got.Reason != "a" {
		t.Fatalf("reason %q", got.Reason)
	}
}

func TestConcurrentEditsAreAllRecorded(t *testing.T) {
	pool := db(t)
	c, _ := insert(t, pool, newCase(1, 2, Ban))
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			if _, err := editOnce(t, pool, 1, c.Number, "edit "+strconv.Itoa(i), 0); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	edits, _ := Edits(t.Context(), pool, 1, c.ID)
	if len(edits) != 20 {
		t.Fatalf("%d edits recorded", len(edits))
	}
	// Each edit's old reason must be the previous edit's new reason: no lost updates.
	for i := 1; i < len(edits); i++ {
		if edits[i].OldReason != edits[i-1].NewReason {
			t.Fatalf("edit %d saw %q, previous wrote %q", i, edits[i].OldReason, edits[i-1].NewReason)
		}
	}
}
