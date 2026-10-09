package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// MaxEscalationSteps caps how many thresholds a guild can configure.
const MaxEscalationSteps = 10

// ErrTooManySteps means the guild already has MaxEscalationSteps thresholds.
var ErrTooManySteps = errors.New("store: too many escalation steps")

// EscalationStep fires Action when a member reaches WarnCount warnings.
type EscalationStep struct {
	WarnCount int
	Action    string // timeout, kick or ban
	Duration  time.Duration
}

// EscalationSteps lists a guild's thresholds by warn count.
func (s *Store) EscalationSteps(ctx context.Context, guildID int64) ([]EscalationStep, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	rows, err := s.pool.Query(ctx, `SELECT "warnCount", "action", coalesce("durationSeconds", 0)
		FROM "escalationSteps" WHERE "guildId" = $1 ORDER BY "warnCount"`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EscalationStep
	for rows.Next() {
		var st EscalationStep
		var seconds int64
		if err := rows.Scan(&st.WarnCount, &st.Action, &seconds); err != nil {
			return nil, err
		}
		st.Duration = time.Duration(seconds) * time.Second
		out = append(out, st)
	}
	return out, rows.Err()
}

// SetEscalationStep adds or replaces the step for st.WarnCount, keeping the per guild cap
// even when two moderators add steps at once.
func (s *Store) SetEscalationStep(ctx context.Context, guildID int64, st EscalationStep) error {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('escalation:' || $1::bigint::text, 0))`, guildID); err != nil {
			return err
		}
		var others int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM "escalationSteps" WHERE "guildId" = $1 AND "warnCount" <> $2`,
			guildID, st.WarnCount).Scan(&others); err != nil {
			return err
		}
		if others >= MaxEscalationSteps {
			return ErrTooManySteps
		}
		_, err := tx.Exec(ctx, `INSERT INTO "escalationSteps" ("guildId", "warnCount", "action", "durationSeconds")
			VALUES ($1, $2, $3, nullif($4::bigint, 0)) ON CONFLICT ("guildId", "warnCount")
			DO UPDATE SET "action" = excluded."action", "durationSeconds" = excluded."durationSeconds"`,
			guildID, st.WarnCount, st.Action, int64(st.Duration/time.Second))
		return err
	})
}

// RemoveEscalationStep deletes a threshold. It reports whether one existed.
func (s *Store) RemoveEscalationStep(ctx context.Context, guildID int64, warnCount int) (bool, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	tag, err := s.pool.Exec(ctx, `DELETE FROM "escalationSteps" WHERE "guildId" = $1 AND "warnCount" = $2`, guildID, warnCount)
	return tag.RowsAffected() == 1, err
}
