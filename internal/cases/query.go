package cases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/store"
)

// ErrNotFound means no case matched in this guild.
var ErrNotFound = errors.New("cases: not found")

// Get loads case number from guildID.
func Get(ctx context.Context, q store.Querier, guildID, number int64) (Case, error) {
	c, err := scan(q.QueryRow(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "number" = $2`, guildID, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return Case{}, ErrNotFound
	}
	return c, err
}

// History lists a target's cases in a guild, newest first.
func History(ctx context.Context, q store.Querier, guildID, targetID int64, limit, offset int) ([]Case, error) {
	rows, err := q.Query(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "targetId" = $2
		ORDER BY "number" DESC LIMIT $3 OFFSET $4`, guildID, targetID, limit, offset)
	if err != nil {
		return nil, err
	}
	return scanAll(rows)
}

// CountByKind counts a target's cases of one kind in a guild.
func CountByKind(ctx context.Context, q store.Querier, guildID, targetID int64, kind Kind) (int, error) {
	var n int
	err := q.QueryRow(ctx, `SELECT count(*) FROM "cases" WHERE "guildId" = $1 AND "targetId" = $2 AND "kind" = $3`,
		guildID, targetID, string(kind)).Scan(&n)
	return n, err
}

// CountUpTo counts a target's cases of one kind numbered at most number. Counting up to a
// specific case keeps the answer stable when newer cases land concurrently.
func CountUpTo(ctx context.Context, q store.Querier, guildID, targetID int64, kind Kind, number int64) (int, error) {
	var n int
	err := q.QueryRow(ctx, `SELECT count(*) FROM "cases" WHERE "guildId" = $1 AND "targetId" = $2 AND "kind" = $3
		AND "number" <= $4`, guildID, targetID, string(kind), number).Scan(&n)
	return n, err
}

// ListByKind lists a target's cases of one kind, newest first.
func ListByKind(ctx context.Context, q store.Querier, guildID, targetID int64, kind Kind, limit int) ([]Case, error) {
	rows, err := q.Query(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "targetId" = $2 AND "kind" = $3
		ORDER BY "number" DESC LIMIT $4`, guildID, targetID, string(kind), limit)
	if err != nil {
		return nil, err
	}
	return scanAll(rows)
}

// Recent finds a case of kind against target created after since, by moderatorID or by
// anyone when moderatorID is 0. Used to catch double submits.
func Recent(ctx context.Context, q store.Querier, guildID, targetID int64, kind Kind, moderatorID int64, since time.Time) (Case, bool, error) {
	c, err := scan(q.QueryRow(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "targetId" = $2
		AND "kind" = $3 AND "createdAt" > $4 AND ($5::bigint = 0 OR "moderatorId" = $5)
		ORDER BY "number" DESC LIMIT 1`, guildID, targetID, string(kind), since, moderatorID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Case{}, false, nil
	}
	return c, err == nil, err
}

// ByKey finds the case created by an automatic action's idempotency key.
func ByKey(ctx context.Context, q store.Querier, guildID int64, key string) (Case, bool, error) {
	c, err := scan(q.QueryRow(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "idempotencyKey" = $2`, guildID, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return Case{}, false, nil
	}
	return c, err == nil, err
}
