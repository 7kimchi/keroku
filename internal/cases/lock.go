package cases

import (
	"context"
	"hash/fnv"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/store"
)

// LockTarget serializes actions on one member of one guild until the transaction ends.
// Different targets hash to different keys. A collision only adds waiting.
func LockTarget(ctx context.Context, tx pgx.Tx, guildID, targetID int64) error {
	h := fnv.New64a()
	_, _ = h.Write([]byte("target:" + strconv.FormatInt(guildID, 10) + ":" + strconv.FormatInt(targetID, 10)))
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(h.Sum64())) // #nosec G115 -- bit pattern reuse is intended
	return err
}

// Claim records that an interaction is being handled. It reports false if another
// transaction already claimed it. Inside a transaction, a rollback releases the claim.
func Claim(ctx context.Context, q store.Querier, interactionID, guildID int64) (bool, error) {
	tag, err := q.Exec(ctx, `INSERT INTO "interactionClaims" ("interactionId", "guildId") VALUES ($1, $2)
		ON CONFLICT ("interactionId") DO NOTHING`, interactionID, guildID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}
