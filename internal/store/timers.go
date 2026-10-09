package store

import (
	"context"
	"time"
)

// Timer kinds, matching the tempActions check constraint.
const (
	TimerUnban         = "unban"
	TimerTimeoutRenew  = "timeoutRenew"
	TimerChannelUnlock = "channelUnlock"
	TimerLockdownEnd   = "lockdownEnd"
)

// ScheduleTimer sets the pending timer for (guild, kind, target), replacing any earlier one.
// dueAt is when the sweeper should next look at it, endsAt when the action really ends.
func ScheduleTimer(ctx context.Context, q Querier, guildID int64, kind string, targetID int64, endsAt, dueAt time.Time) error {
	_, err := q.Exec(ctx, `INSERT INTO "tempActions" ("guildId", "kind", "targetId", "endsAt", "dueAt")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ("guildId", "kind", "targetId") WHERE "status" = 'pending'
		DO UPDATE SET "endsAt" = excluded."endsAt", "dueAt" = excluded."dueAt", "attempts" = 0,
			"lastError" = NULL, "updatedAt" = now()`, guildID, kind, targetID, endsAt, dueAt)
	return err
}

// CancelTimer marks the pending timer for (guild, kind, target) cancelled. It reports
// whether one was pending.
func CancelTimer(ctx context.Context, q Querier, guildID int64, kind string, targetID int64) (bool, error) {
	tag, err := q.Exec(ctx, `UPDATE "tempActions" SET "status" = 'cancelled', "updatedAt" = now()
		WHERE "guildId" = $1 AND "kind" = $2 AND "targetId" = $3 AND "status" = 'pending'`, guildID, kind, targetID)
	return tag.RowsAffected() > 0, err
}

// PendingTimer returns when the pending timer ends, if there is one.
func PendingTimer(ctx context.Context, q Querier, guildID int64, kind string, targetID int64) (time.Time, bool, error) {
	var endsAt time.Time
	err := q.QueryRow(ctx, `SELECT "endsAt" FROM "tempActions" WHERE "guildId" = $1 AND "kind" = $2
		AND "targetId" = $3 AND "status" = 'pending'`, guildID, kind, targetID).Scan(&endsAt)
	if isNoRows(err) {
		return time.Time{}, false, nil
	}
	return endsAt, err == nil, err
}
