package store_test

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestTimers(t *testing.T) {
	pool := dbtest.New(t)
	ctx := t.Context()
	end := time.Now().Add(time.Hour).Truncate(time.Microsecond)
	if err := store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 2, end, end); err != nil {
		t.Fatal(err)
	}
	later := end.Add(time.Hour)
	if err := store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 2, later, later); err != nil {
		t.Fatal("replace failed:", err)
	}
	got, ok, err := store.PendingTimer(ctx, pool, 1, store.TimerUnban, 2)
	if err != nil || !ok || !got.Equal(later) {
		t.Fatalf("pending %v %v %v", got, ok, err)
	}
	var rows int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM "tempActions"`).Scan(&rows)
	if rows != 1 {
		t.Fatalf("%d rows, replace should not add one", rows)
	}
	if _, ok, _ := store.PendingTimer(ctx, pool, 9, store.TimerUnban, 2); ok {
		t.Fatal("timer visible from another guild")
	}
	if ok, _ := store.CancelTimer(ctx, pool, 9, store.TimerUnban, 2); ok {
		t.Fatal("cancelled from another guild")
	}
	if ok, err := store.CancelTimer(ctx, pool, 1, store.TimerUnban, 2); !ok || err != nil {
		t.Fatal("cancel")
	}
	if _, ok, _ := store.PendingTimer(ctx, pool, 1, store.TimerUnban, 2); ok {
		t.Fatal("cancelled timer still pending")
	}
	if err := store.ScheduleTimer(ctx, pool, 1, store.TimerUnban, 2, end, end); err != nil {
		t.Fatal("new timer after cancel:", err)
	}
	if err := store.ScheduleTimer(ctx, pool, 1, "explode", 2, end, end); err == nil {
		t.Fatal("bad kind accepted")
	}
}
