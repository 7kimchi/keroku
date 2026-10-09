package store

import (
	"context"
	"time"
)

// Retention for bookkeeping rows. Cases and settings are never pruned.
const (
	ClaimRetention = 24 * time.Hour
	TimerRetention = 30 * 24 * time.Hour
	pruneBatch     = 5000
)

// Prune deletes interaction claims and finished timers past retention, in bounded batches
// so one run never holds long locks. It returns how many rows went.
func (s *Store) Prune(ctx context.Context, now time.Time) (int64, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	var total int64
	for _, q := range []struct {
		sql    string
		before time.Time
	}{
		{`DELETE FROM "interactionClaims" WHERE "interactionId" IN (
			SELECT "interactionId" FROM "interactionClaims" WHERE "createdAt" < $1 LIMIT $2)`, now.Add(-ClaimRetention)},
		{`DELETE FROM "tempActions" WHERE "id" IN (
			SELECT "id" FROM "tempActions" WHERE "status" <> 'pending' AND "updatedAt" < $1 LIMIT $2)`, now.Add(-TimerRetention)},
	} {
		tag, err := s.pool.Exec(ctx, q.sql, q.before, pruneBatch)
		if err != nil {
			return total, err
		}
		total += tag.RowsAffected()
	}
	return total, nil
}
