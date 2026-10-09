package store

import (
	"context"
	"time"
)

// RaidSettings configures raid protection for one guild.
type RaidSettings struct {
	Enabled       bool
	JoinLimit     int
	Window        time.Duration
	MinAccountAge time.Duration
	Action        string // kick, ban or lockdown
	LockdownFor   time.Duration
}

// DefaultRaid matches the column defaults: off.
func DefaultRaid() RaidSettings {
	return RaidSettings{JoinLimit: 10, Window: 10 * time.Second, Action: "kick", LockdownFor: 15 * time.Minute}
}

// Raid loads a guild's raid settings, or the defaults.
func (s *Store) Raid(ctx context.Context, guildID int64) (RaidSettings, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	r := DefaultRaid()
	var win, age, lock int64
	err := s.pool.QueryRow(ctx, `SELECT "enabled", "joinLimit", "windowSeconds", "minAccountAgeSeconds", "action",
		"lockdownSeconds" FROM "raidSettings" WHERE "guildId" = $1`, guildID).
		Scan(&r.Enabled, &r.JoinLimit, &win, &age, &r.Action, &lock)
	if isNoRows(err) {
		return DefaultRaid(), nil
	}
	if err != nil {
		return RaidSettings{}, err
	}
	r.Window, r.MinAccountAge = time.Duration(win)*time.Second, time.Duration(age)*time.Second
	r.LockdownFor = time.Duration(lock) * time.Second
	return r, nil
}

// SetRaid stores every raid setting for a guild. The database checks the ranges.
func (s *Store) SetRaid(ctx context.Context, guildID int64, r RaidSettings) error {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	_, err := s.pool.Exec(ctx, `INSERT INTO "raidSettings" ("guildId", "enabled", "joinLimit", "windowSeconds",
		"minAccountAgeSeconds", "action", "lockdownSeconds") VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("guildId") DO UPDATE SET "enabled" = excluded."enabled", "joinLimit" = excluded."joinLimit",
		"windowSeconds" = excluded."windowSeconds", "minAccountAgeSeconds" = excluded."minAccountAgeSeconds",
		"action" = excluded."action", "lockdownSeconds" = excluded."lockdownSeconds", "updatedAt" = now()`,
		guildID, r.Enabled, r.JoinLimit, int64(r.Window/time.Second), int64(r.MinAccountAge/time.Second),
		r.Action, int64(r.LockdownFor/time.Second))
	return err
}
