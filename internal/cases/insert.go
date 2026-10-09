package cases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/store"
)

// ErrDuplicate means the interaction or idempotency key already produced a case.
var ErrDuplicate = errors.New("cases: duplicate")

const columns = `"id", "guildId", "number", "kind", "targetId", "moderatorId", "reason",
	coalesce("durationSeconds", 0), coalesce("interactionId", 0), coalesce("idempotencyKey", ''), "details", "createdAt"`

// Insert assigns the next number for the guild and stores the case. Call it inside a
// transaction: the counter row stays locked until commit, so numbers never gap or repeat.
func Insert(ctx context.Context, tx pgx.Tx, n New) (Case, error) {
	if n.GuildID <= 0 || n.TargetID <= 0 || n.ModeratorID <= 0 || !n.Kind.Valid() || n.Duration < 0 || !n.Details.Valid() {
		return Case{}, errors.New("cases: invalid case")
	}
	var number int64
	err := tx.QueryRow(ctx, `INSERT INTO "caseCounters" ("guildId", "lastNumber") VALUES ($1, 1)
		ON CONFLICT ("guildId") DO UPDATE SET "lastNumber" = "caseCounters"."lastNumber" + 1
		RETURNING "lastNumber"`, n.GuildID).Scan(&number)
	if err != nil {
		return Case{}, err
	}
	row := tx.QueryRow(ctx, `INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason",
		"durationSeconds", "interactionId", "idempotencyKey", "details")
		VALUES ($1, $2, $3, $4, $5, $6, nullif($7::bigint, 0), nullif($8::bigint, 0), nullif($9::text, ''), $10::jsonb)
		RETURNING `+columns, n.GuildID, number, string(n.Kind), n.TargetID, n.ModeratorID, n.Reason,
		int64(n.Duration/time.Second), n.InteractionID, n.IdempotencyKey, n.Details)
	c, err := scan(row)
	if store.IsUniqueViolation(err) {
		return Case{}, ErrDuplicate
	}
	return c, err
}

func scan(row pgx.Row) (Case, error) {
	var c Case
	var kind string
	var seconds int64
	err := row.Scan(&c.ID, &c.GuildID, &c.Number, &kind, &c.TargetID, &c.ModeratorID, &c.Reason,
		&seconds, &c.InteractionID, &c.IdempotencyKey, &c.Details, &c.CreatedAt)
	c.Kind, c.Duration = Kind(kind), time.Duration(seconds)*time.Second
	return c, err
}

func scanAll(rows pgx.Rows) ([]Case, error) {
	defer rows.Close()
	var out []Case
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
