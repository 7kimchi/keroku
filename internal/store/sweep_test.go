package store_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestClaimSkipsLockedAndFuture(t *testing.T) {
	pool := dbtest.New(t)
	ctx := t.Context()
	now := time.Now()
	_ = store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 2, now.Add(-time.Minute), now.Add(-time.Minute))
	_ = store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 3, now.Add(-time.Second), now.Add(-time.Second))
	_ = store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 4, now.Add(time.Hour), now.Add(time.Hour))
	first, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Rollback(context.Background()) }()
	a, ok, err := store.ClaimDue(ctx, first, now)
	if !ok || err != nil || a.TargetID != 2 {
		t.Fatalf("first claim %+v %v %v", a, ok, err)
	}
	_ = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		b, ok, _ := store.ClaimDue(ctx, tx, now)
		if !ok || b.TargetID != 3 {
			t.Fatalf("second claim got %+v", b)
		}
		// A third sweeper finds nothing: both due rows are held and the last is in the future.
		_ = pgx.BeginFunc(ctx, pool, func(third pgx.Tx) error {
			if c, ok, _ := store.ClaimDue(ctx, third, now); ok {
				t.Fatalf("claimed a held or future timer: %+v", c)
			}
			return nil
		})
		return store.FinishJob(ctx, tx, b.ID, "done", "")
	})
	if err := store.RescheduleJob(ctx, first, a.ID, now.Add(time.Hour), 1, strings.Repeat("e", 900)); err != nil {
		t.Fatal(err)
	}
	if err := first.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var status, lastErr string
	var attempts int
	_ = pool.QueryRow(ctx, `SELECT "status", "attempts", "lastError" FROM "tempActions" WHERE "targetId" = 2`).Scan(&status, &attempts, &lastErr)
	if status != "pending" || attempts != 1 || len(lastErr) != 500 {
		t.Fatalf("rescheduled row %s %d %d", status, attempts, len(lastErr))
	}
	_ = pool.QueryRow(ctx, `SELECT "status" FROM "tempActions" WHERE "targetId" = 3`).Scan(&status)
	if status != "done" {
		t.Fatalf("finished row %s", status)
	}
}
