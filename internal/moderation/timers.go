package moderation

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
)

// renewMargin re-applies a long timeout an hour before Discord's 28 day timeout lapses.
const renewMargin = time.Hour

// timers keeps the pending unban and timeout renewal in step with the action, in the same
// transaction as the case.
func (s *Service) timers(ctx context.Context, tx pgx.Tx, a Action) error {
	now := s.now()
	switch a.Kind {
	case cases.Ban:
		if a.Duration == 0 {
			_, err := store.CancelTimer(ctx, tx, a.GuildID, store.TimerUnban, a.TargetID)
			return err
		}
		end := now.Add(a.Duration)
		return store.ScheduleTimer(ctx, tx, a.GuildID, store.TimerUnban, a.TargetID, end, end)
	case cases.Unban:
		_, err := store.CancelTimer(ctx, tx, a.GuildID, store.TimerUnban, a.TargetID)
		return err
	case cases.Timeout:
		if a.Duration <= DiscordTimeout {
			_, err := store.CancelTimer(ctx, tx, a.GuildID, store.TimerTimeoutRenew, a.TargetID)
			return err
		}
		end, due := now.Add(a.Duration), now.Add(DiscordTimeout-renewMargin)
		return store.ScheduleTimer(ctx, tx, a.GuildID, store.TimerTimeoutRenew, a.TargetID, end, due)
	case cases.Untimeout:
		_, err := store.CancelTimer(ctx, tx, a.GuildID, store.TimerTimeoutRenew, a.TargetID)
		return err
	}
	return nil
}
