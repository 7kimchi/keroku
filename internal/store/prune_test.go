package store_test

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestPrune(t *testing.T) {
	pool := dbtest.New(t)
	s := store.New(pool)
	ctx := t.Context()
	_, _ = pool.Exec(ctx, `INSERT INTO "interactionClaims" ("interactionId", "guildId", "createdAt")
		VALUES (1, 1, now() - interval '2 days'), (2, 1, now())`)
	_, _ = pool.Exec(ctx, `INSERT INTO "tempActions" ("guildId", "kind", "targetId", "endsAt", "dueAt", "status", "updatedAt") VALUES
		(1, 'unban', 1, now(), now(), 'done', now() - interval '40 days'),
		(1, 'unban', 2, now(), now(), 'pending', now() - interval '40 days'),
		(1, 'unban', 3, now(), now(), 'failed', now())`)
	_, _ = pool.Exec(ctx, `INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason", "createdAt")
		VALUES (1, 1, 'ban', 2, 3, 'old', now() - interval '5 years')`)
	n, err := s.Prune(ctx, time.Now())
	if err != nil || n != 2 {
		t.Fatalf("pruned %d %v", n, err)
	}
	var claims, timers, cases int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM "interactionClaims"`).Scan(&claims)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM "tempActions"`).Scan(&timers)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM "cases"`).Scan(&cases)
	if claims != 1 || timers != 2 || cases != 1 {
		t.Fatalf("claims %d timers %d cases %d", claims, timers, cases)
	}
}
