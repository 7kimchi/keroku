package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// GuildSettings are a guild's channel choices. Zero means unset.
type GuildSettings struct {
	ModlogChannelID int64
	LogChannelID    int64
}

// GuildSettings loads a guild's settings. A guild with no row gets zero values.
func (s *Store) GuildSettings(ctx context.Context, guildID int64) (GuildSettings, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	var g GuildSettings
	err := s.pool.QueryRow(ctx, `SELECT coalesce("modlogChannelId", 0), coalesce("logChannelId", 0)
		FROM "guildSettings" WHERE "guildId" = $1`, guildID).Scan(&g.ModlogChannelID, &g.LogChannelID)
	if errors.Is(err, pgx.ErrNoRows) {
		return GuildSettings{}, nil
	}
	return g, err
}

// SetModlogChannel stores the modlog channel. 0 clears it.
func (s *Store) SetModlogChannel(ctx context.Context, guildID, channelID int64) error {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	_, err := s.pool.Exec(ctx, `INSERT INTO "guildSettings" ("guildId", "modlogChannelId") VALUES ($1, nullif($2::bigint, 0))
		ON CONFLICT ("guildId") DO UPDATE SET "modlogChannelId" = excluded."modlogChannelId", "updatedAt" = now()`,
		guildID, channelID)
	return err
}

// SetLogChannel stores the message and member log channel. 0 clears it.
func (s *Store) SetLogChannel(ctx context.Context, guildID, channelID int64) error {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	_, err := s.pool.Exec(ctx, `INSERT INTO "guildSettings" ("guildId", "logChannelId") VALUES ($1, nullif($2::bigint, 0))
		ON CONFLICT ("guildId") DO UPDATE SET "logChannelId" = excluded."logChannelId", "updatedAt" = now()`,
		guildID, channelID)
	return err
}
