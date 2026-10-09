package cases

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestDetailsRoundTrip(t *testing.T) {
	pool := db(t)
	n := newCase(1, 2, Ban)
	n.Details = Details{Source: FromRaid, DeleteSeconds: 3600}
	c, err := insert(t, pool, n)
	if err != nil || c.Details != n.Details {
		t.Fatalf("%+v %v", c.Details, err)
	}
	got, err := Get(context.Background(), pool, 1, c.Number)
	if err != nil || got.Details != n.Details {
		t.Fatalf("%+v %v", got.Details, err)
	}
	var raw string
	_ = pool.QueryRow(t.Context(), `SELECT "details"::text FROM "cases" WHERE "id" = $1`, c.ID).Scan(&raw)
	if raw != `{"source": "raid", "deleteSeconds": 3600}` {
		t.Fatalf("stored %s", raw)
	}
}

func TestDetailsEmptyIsObject(t *testing.T) {
	pool := db(t)
	c, _ := insert(t, pool, newCase(1, 2, Warn))
	var typ string
	_ = pool.QueryRow(t.Context(), `SELECT jsonb_typeof("details") FROM "cases" WHERE "id" = $1`, c.ID).Scan(&typ)
	if typ != "object" || c.Details != (Details{}) {
		t.Fatalf("%s %+v", typ, c.Details)
	}
}

func TestInvalidDetailsNeverReachSQL(t *testing.T) {
	pool := db(t)
	n := newCase(1, 2, Warn)
	n.Details = Details{Source: "forged"}
	if _, err := insert(t, pool, n); err == nil || strings.Contains(err.Error(), "SQLSTATE") {
		t.Fatalf("got %v", err)
	}
	var count int
	_ = pool.QueryRow(t.Context(), `SELECT count(*) FROM "caseCounters"`).Scan(&count)
	if count != 0 {
		t.Fatal("counter moved for a rejected case")
	}
}

// The database refuses anything but a small object, even from a raw query.
func TestDetailsColumnConstraints(t *testing.T) {
	pool := db(t)
	for _, v := range []string{`[]`, `"x"`, `1`, `null`, `{"rule":"` + strings.Repeat("a", 1100) + `"}`} {
		_, err := pool.Exec(t.Context(), `INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason", "details")
			VALUES (1, 1, 'ban', 2, 3, 'x', $1::jsonb)`, v)
		if err == nil {
			t.Fatalf("%.20s accepted", v)
		}
	}
}

// Details are written once. Only the reason may change after insert.
func TestDetailsImmutable(t *testing.T) {
	pool := db(t)
	c, _ := insert(t, pool, newCase(1, 2, Ban))
	for _, sql := range []string{`UPDATE "cases" SET "details" = '{"source":"raid"}'`,
		`UPDATE "cases" SET "details" = "details" || '{"x":1}'`, `UPDATE "cases" SET "details" = '{}' , "kind" = 'kick'`} {
		if _, err := pool.Exec(t.Context(), sql); err == nil {
			t.Fatalf("%s allowed", sql)
		}
	}
	if _, err := pool.Exec(t.Context(), `UPDATE "cases" SET "reason" = 'y' WHERE "id" = $1`, c.ID); err != nil {
		t.Fatalf("reason edit blocked: %v", err)
	}
}

// Parallel inserts with details keep numbers gapless and details attached to the right case.
func TestDetailsConcurrentInserts(t *testing.T) {
	pool := db(t)
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			n := newCase(1, int64(100+i), Timeout)
			n.Duration, n.Details = 60e9, Details{Source: FromTimer, TimerID: int64(i + 1)}
			if _, err := insert(t, pool, n); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	var bad, maxN int
	_ = pool.QueryRow(t.Context(), `SELECT count(*) FILTER (WHERE ("details"->>'timerId')::bigint <> "targetId" - 99),
		max("number") FROM "cases" WHERE "guildId" = 1`).Scan(&bad, &maxN)
	if bad != 0 || maxN != 200 {
		t.Fatalf("mismatched %d max %d", bad, maxN)
	}
}
