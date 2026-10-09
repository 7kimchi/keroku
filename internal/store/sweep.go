package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Job is a due timer claimed by one sweeper.
type Job struct {
	ID       int64
	GuildID  int64
	Kind     string
	TargetID int64
	EndsAt   time.Time
	Attempts int
}

// ClaimDue locks the oldest due timer that no other transaction holds. Other sweepers
// skip it until this transaction ends, so a row is never worked on twice at once.
func ClaimDue(ctx context.Context, tx pgx.Tx, now time.Time) (Job, bool, error) {
	var j Job
	err := tx.QueryRow(ctx, `SELECT "id", "guildId", "kind", "targetId", "endsAt", "attempts" FROM "tempActions"
		WHERE "status" = 'pending' AND "dueAt" <= $1 ORDER BY "dueAt" LIMIT 1 FOR UPDATE SKIP LOCKED`, now).
		Scan(&j.ID, &j.GuildID, &j.Kind, &j.TargetID, &j.EndsAt, &j.Attempts)
	if isNoRows(err) {
		return Job{}, false, nil
	}
	return j, err == nil, err
}

// FinishJob marks a claimed timer done or failed.
func FinishJob(ctx context.Context, tx pgx.Tx, id int64, status, lastError string) error {
	_, err := tx.Exec(ctx, `UPDATE "tempActions" SET "status" = $2, "lastError" = nullif($3::text, ''),
		"updatedAt" = now() WHERE "id" = $1`, id, status, truncate(lastError, 500))
	return err
}

// RescheduleJob keeps a claimed timer pending and moves its next run.
func RescheduleJob(ctx context.Context, tx pgx.Tx, id int64, dueAt time.Time, attempts int, lastError string) error {
	_, err := tx.Exec(ctx, `UPDATE "tempActions" SET "dueAt" = $2, "attempts" = $3, "lastError" = nullif($4::text, ''),
		"updatedAt" = now() WHERE "id" = $1`, id, dueAt, attempts, truncate(lastError, 500))
	return err
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}
