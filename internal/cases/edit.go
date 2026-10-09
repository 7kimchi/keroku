package cases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/store"
)

// Edit is one reason change.
type Edit struct {
	EditorID  int64
	OldReason string
	NewReason string
	CreatedAt time.Time
}

// EditReason changes a case's reason and records the change. The row is locked so two
// concurrent edits apply one after the other and both land in the history.
func EditReason(ctx context.Context, tx pgx.Tx, guildID, number, editorID int64, reason string, interactionID int64) (Case, error) {
	c, err := scan(tx.QueryRow(ctx, `SELECT `+columns+` FROM "cases" WHERE "guildId" = $1 AND "number" = $2 FOR UPDATE`, guildID, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return Case{}, ErrNotFound
	}
	if err != nil {
		return Case{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO "caseEdits" ("guildId", "caseId", "editorId", "oldReason", "newReason", "interactionId")
		VALUES ($1, $2, $3, $4, $5, nullif($6::bigint, 0))`, guildID, c.ID, editorID, c.Reason, reason, interactionID)
	if store.IsUniqueViolation(err) {
		return Case{}, ErrDuplicate
	}
	if err != nil {
		return Case{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE "cases" SET "reason" = $1 WHERE "id" = $2 AND "guildId" = $3`, reason, c.ID, guildID); err != nil {
		return Case{}, err
	}
	c.Reason = reason
	return c, nil
}

// Edits lists a case's reason changes, oldest first.
func Edits(ctx context.Context, q store.Querier, guildID, caseID int64) ([]Edit, error) {
	rows, err := q.Query(ctx, `SELECT "editorId", "oldReason", "newReason", "createdAt" FROM "caseEdits"
		WHERE "guildId" = $1 AND "caseId" = $2 ORDER BY "id"`, guildID, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Edit
	for rows.Next() {
		var e Edit
		if err := rows.Scan(&e.EditorID, &e.OldReason, &e.NewReason, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
