package sweeper

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// Three sweepers race over the same expired timers. Each timer must be handled once.
func TestSeveralInstancesProcessEachTimerOnce(t *testing.T) {
	e := setup(t)
	const n = 300
	past := e.clk.Now().Add(-time.Minute)
	for i := range n {
		target := int64(200000000000000000 + i)
		_ = e.fake.Ban(t.Context(), gs, itoa(target), 0, "")
		if err := store.ScheduleTimer(t.Context(), e.store.Pool(), guild, store.TimerUnban, target, past, past); err != nil {
			t.Fatal(err)
		}
	}
	instances := []*Sweeper{e.sw, New(e.sw.d), New(e.sw.d)}
	var wg sync.WaitGroup
	for _, sw := range instances {
		for range 4 {
			wg.Go(func() {
				for {
					worked, err := sw.Once(t.Context())
					if err != nil {
						t.Error(err)
						return
					}
					if !worked {
						return
					}
				}
			})
		}
	}
	wg.Wait()
	var done, cases int
	_ = e.store.Pool().QueryRow(t.Context(), `SELECT count(*) FROM "tempActions" WHERE "status" = 'done'`).Scan(&done)
	_ = e.store.Pool().QueryRow(t.Context(), `SELECT count(*) FROM "cases" WHERE "kind" = 'unban'`).Scan(&cases)
	if done != n || cases != n || e.fake.Calls("unban") != n {
		t.Fatalf("done %d cases %d unban calls %d, want %d each", done, cases, e.fake.Calls("unban"), n)
	}
}
