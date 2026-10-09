package cleanup

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// Lockdown removes the lock permissions from @everyone server wide. Channel overwrites
// that allow sending still apply. A non zero until schedules the end.
func (s *Service) Lockdown(ctx context.Context, guildID int64, reason string, until time.Time) error {
	gid := validate.FormatSnowflake(guildID)
	return s.store.InTx(ctx, func(tx pgx.Tx) error {
		everyone, err := s.everyone(ctx, gid)
		if err != nil {
			return err
		}
		ok, err := store.AddLockdown(ctx, tx, guildID, everyone.Permissions&perms.LockBits)
		if err != nil {
			return err
		}
		if !ok {
			return commands.Fail("Lockdown failed", "Server is already in lockdown.")
		}
		if !until.IsZero() {
			if err := store.ScheduleTimer(ctx, tx, guildID, store.TimerLockdownEnd, guildID, until, until); err != nil {
				return err
			}
		}
		return s.client.SetRolePermissions(ctx, gid, gid, everyone.Permissions&^perms.LockBits, reason)
	})
}

// EndLockdown gives @everyone back the lock permissions it had. It reports false if no
// lockdown was active.
func (s *Service) EndLockdown(ctx context.Context, guildID int64, reason string) (bool, error) {
	gid := validate.FormatSnowflake(guildID)
	found := false
	err := s.store.InTx(ctx, func(tx pgx.Tx) error {
		previous, ok, err := store.TakeLockdown(ctx, tx, guildID)
		if err != nil || !ok {
			return err
		}
		found = true
		if _, err := store.CancelTimer(ctx, tx, guildID, store.TimerLockdownEnd, guildID); err != nil {
			return err
		}
		everyone, err := s.everyone(ctx, gid)
		if err != nil {
			return err
		}
		return s.client.SetRolePermissions(ctx, gid, gid, everyone.Permissions&^perms.LockBits|previous, reason)
	})
	return found, err
}

// everyone fetches the @everyone role fresh.
func (s *Service) everyone(ctx context.Context, gid string) (*discordgo.Role, error) {
	g, err := s.client.Guild(ctx, gid)
	if err != nil {
		return nil, err
	}
	for _, r := range g.Roles {
		if r.ID == gid {
			return r, nil
		}
	}
	return nil, commands.Fail("Lockdown failed", "The @everyone role is missing.")
}
