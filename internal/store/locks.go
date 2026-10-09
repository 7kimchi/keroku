package store

import "context"

// ChannelLock remembers a channel's @everyone overwrite bits from before a lock.
type ChannelLock struct {
	HadOverwrite  bool
	PreviousAllow int64
	PreviousDeny  int64
}

// AddChannelLock records a lock. It reports false if the channel is already locked.
func AddChannelLock(ctx context.Context, q Querier, guildID, channelID int64, l ChannelLock) (bool, error) {
	tag, err := q.Exec(ctx, `INSERT INTO "channelLocks" ("guildId", "channelId", "hadOverwrite", "previousAllow", "previousDeny")
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("channelId") DO NOTHING`,
		guildID, channelID, l.HadOverwrite, l.PreviousAllow, l.PreviousDeny)
	return tag.RowsAffected() == 1, err
}

// TakeChannelLock removes and returns a lock. It reports false if there was none.
func TakeChannelLock(ctx context.Context, q Querier, guildID, channelID int64) (ChannelLock, bool, error) {
	var l ChannelLock
	err := q.QueryRow(ctx, `DELETE FROM "channelLocks" WHERE "guildId" = $1 AND "channelId" = $2
		RETURNING "hadOverwrite", "previousAllow", "previousDeny"`, guildID, channelID).
		Scan(&l.HadOverwrite, &l.PreviousAllow, &l.PreviousDeny)
	if isNoRows(err) {
		return ChannelLock{}, false, nil
	}
	return l, err == nil, err
}

// AddLockdown records a server lockdown with the @everyone permissions from before it.
func AddLockdown(ctx context.Context, q Querier, guildID, previous int64) (bool, error) {
	tag, err := q.Exec(ctx, `INSERT INTO "lockdowns" ("guildId", "previousPermissions") VALUES ($1, $2)
		ON CONFLICT ("guildId") DO NOTHING`, guildID, previous)
	return tag.RowsAffected() == 1, err
}

// TakeLockdown removes and returns a lockdown's previous permissions.
func TakeLockdown(ctx context.Context, q Querier, guildID int64) (int64, bool, error) {
	var previous int64
	err := q.QueryRow(ctx, `DELETE FROM "lockdowns" WHERE "guildId" = $1 RETURNING "previousPermissions"`, guildID).Scan(&previous)
	if isNoRows(err) {
		return 0, false, nil
	}
	return previous, err == nil, err
}
