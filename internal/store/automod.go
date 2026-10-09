package store

import (
	"context"
	"time"
)

// AutomodSettings configures the automod rules for one guild.
type AutomodSettings struct {
	SpamEnabled      bool
	SpamMessages     int
	SpamWindow       time.Duration
	DuplicateEnabled bool
	DuplicateCount   int
	DuplicateWindow  time.Duration
	LinksEnabled     bool
	AllowedDomains   []string
	Timeout          time.Duration // 0 means delete only
	MentionLimit     int           // native rule, 0 means off
	InvitesBlocked   bool          // native rule
}

// DefaultAutomod matches the column defaults: every rule off.
func DefaultAutomod() AutomodSettings {
	return AutomodSettings{SpamMessages: 6, SpamWindow: 5 * time.Second, DuplicateCount: 3, DuplicateWindow: 30 * time.Second}
}

// Automod loads a guild's automod settings, or the defaults.
func (s *Store) Automod(ctx context.Context, guildID int64) (AutomodSettings, error) {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	a := DefaultAutomod()
	var spamWin, dupWin, timeout int64
	err := s.pool.QueryRow(ctx, `SELECT "spamEnabled", "spamMessages", "spamWindowSeconds", "duplicateEnabled",
		"duplicateCount", "duplicateWindowSeconds", "linksEnabled", "allowedDomains", "timeoutSeconds",
		"mentionLimit", "invitesBlocked" FROM "automodSettings" WHERE "guildId" = $1`, guildID).
		Scan(&a.SpamEnabled, &a.SpamMessages, &spamWin, &a.DuplicateEnabled, &a.DuplicateCount, &dupWin,
			&a.LinksEnabled, &a.AllowedDomains, &timeout, &a.MentionLimit, &a.InvitesBlocked)
	if isNoRows(err) {
		return DefaultAutomod(), nil
	}
	if err != nil {
		return AutomodSettings{}, err
	}
	a.SpamWindow, a.DuplicateWindow = time.Duration(spamWin)*time.Second, time.Duration(dupWin)*time.Second
	a.Timeout = time.Duration(timeout) * time.Second
	return a, nil
}

// SetAutomod stores every automod setting for a guild. The database checks the ranges.
func (s *Store) SetAutomod(ctx context.Context, guildID int64, a AutomodSettings) error {
	ctx, cancel := Bounded(ctx)
	defer cancel()
	if a.AllowedDomains == nil {
		a.AllowedDomains = []string{}
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO "automodSettings" ("guildId", "spamEnabled", "spamMessages",
		"spamWindowSeconds", "duplicateEnabled", "duplicateCount", "duplicateWindowSeconds", "linksEnabled",
		"allowedDomains", "timeoutSeconds", "mentionLimit", "invitesBlocked")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT ("guildId") DO UPDATE SET "spamEnabled" = excluded."spamEnabled",
		"spamMessages" = excluded."spamMessages", "spamWindowSeconds" = excluded."spamWindowSeconds",
		"duplicateEnabled" = excluded."duplicateEnabled", "duplicateCount" = excluded."duplicateCount",
		"duplicateWindowSeconds" = excluded."duplicateWindowSeconds", "linksEnabled" = excluded."linksEnabled",
		"allowedDomains" = excluded."allowedDomains", "timeoutSeconds" = excluded."timeoutSeconds",
		"mentionLimit" = excluded."mentionLimit", "invitesBlocked" = excluded."invitesBlocked", "updatedAt" = now()`,
		guildID, a.SpamEnabled, a.SpamMessages, int64(a.SpamWindow/time.Second), a.DuplicateEnabled, a.DuplicateCount,
		int64(a.DuplicateWindow/time.Second), a.LinksEnabled, a.AllowedDomains, int64(a.Timeout/time.Second),
		a.MentionLimit, a.InvitesBlocked)
	return err
}
