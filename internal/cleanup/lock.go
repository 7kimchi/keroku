package cleanup

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// Lock denies @everyone the lock permissions in a channel and remembers what to restore.
// A non zero until schedules the unlock.
func (s *Service) Lock(ctx context.Context, guildID, channelID int64, reason string, until time.Time) error {
	gid, cid := validate.FormatSnowflake(guildID), validate.FormatSnowflake(channelID)
	return s.store.InTx(ctx, func(tx pgx.Tx) error {
		ch, err := s.client.Channel(ctx, cid)
		if err != nil {
			return err
		}
		if ch.GuildID != gid {
			return commands.Fail("Lock failed", "Channel is not in this server.")
		}
		var allow, deny int64
		had := false
		for _, o := range ch.PermissionOverwrites {
			if o.ID == gid {
				allow, deny, had = o.Allow, o.Deny, true
			}
		}
		ok, err := store.AddChannelLock(ctx, tx, guildID, channelID, store.ChannelLock{
			HadOverwrite: had, PreviousAllow: allow & perms.LockBits, PreviousDeny: deny & perms.LockBits})
		if err != nil {
			return err
		}
		if !ok {
			return commands.Fail("Lock failed", "Channel is already locked.")
		}
		if !until.IsZero() {
			if err := store.ScheduleTimer(ctx, tx, guildID, store.TimerChannelUnlock, channelID, until, until); err != nil {
				return err
			}
		}
		return s.client.SetRoleOverwrite(ctx, cid, gid, allow&^perms.LockBits, deny|perms.LockBits, reason)
	})
}

// Unlock restores the lock permissions to what they were. It reports false if the channel
// was not locked by Keroku. Other overwrite bits changed meanwhile are kept.
func (s *Service) Unlock(ctx context.Context, guildID, channelID int64, reason string) (bool, error) {
	gid, cid := validate.FormatSnowflake(guildID), validate.FormatSnowflake(channelID)
	found := false
	err := s.store.InTx(ctx, func(tx pgx.Tx) error {
		l, ok, err := store.TakeChannelLock(ctx, tx, guildID, channelID)
		if err != nil || !ok {
			return err
		}
		found = true
		if _, err := store.CancelTimer(ctx, tx, guildID, store.TimerChannelUnlock, channelID); err != nil {
			return err
		}
		ch, err := s.client.Channel(ctx, cid)
		if err != nil {
			return err
		}
		var allow, deny int64
		for _, o := range ch.PermissionOverwrites {
			if o.ID == gid {
				allow, deny = o.Allow, o.Deny
			}
		}
		allow = allow&^perms.LockBits | l.PreviousAllow
		deny = deny&^perms.LockBits | l.PreviousDeny
		if !l.HadOverwrite && allow == 0 && deny == 0 {
			return s.client.DeleteOverwrite(ctx, cid, gid, reason)
		}
		return s.client.SetRoleOverwrite(ctx, cid, gid, allow, deny, reason)
	})
	return found, err
}
