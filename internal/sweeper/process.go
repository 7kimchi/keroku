package sweeper

import (
	"context"
	"strconv"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// process does one timer's work. Every branch is safe to repeat after a crash.
func (s *Sweeper) process(ctx context.Context, job store.Job) outcome {
	switch job.Kind {
	case store.TimerUnban:
		return s.unban(ctx, job)
	case store.TimerTimeoutRenew:
		return s.renew(ctx, job)
	case store.TimerChannelUnlock:
		if _, err := s.d.Cleanup.UnlockFromTimer(ctx, job.GuildID, job.TargetID, "Timed lock ended."); err != nil {
			return s.retry(job, err)
		}
		return outcome{status: "done"}
	case store.TimerLockdownEnd:
		if _, err := s.d.Cleanup.EndLockdownFromTimer(ctx, job.GuildID, "Timed lockdown ended."); err != nil {
			return s.retry(job, err)
		}
		return outcome{status: "done"}
	}
	return outcome{status: "failed", err: commands.Fail("Timer failed", "Unknown timer kind.")}
}

// unban lifts a temporary ban as a case. The timer id is the idempotency key, so a retry
// after a crash finds the earlier case instead of making a second one.
func (s *Sweeper) unban(ctx context.Context, job store.Job) outcome {
	_, err := s.d.Moderation.Execute(ctx, moderation.Action{
		Kind: cases.Unban, GuildID: job.GuildID, TargetID: job.TargetID, ModeratorID: s.d.BotID,
		Reason: "Temporary ban ended.", IdempotencyKey: "timer:" + strconv.FormatInt(job.ID, 10),
		Automated: true, FromTimer: true,
	})
	var refused *commands.UserError
	if err != nil && asUserErr(err, &refused) && refused.Detail == moderation.NotBanned {
		// Someone lifted the ban outside Keroku. Nothing left to do.
		return outcome{status: "done"}
	}
	if err != nil {
		return s.retry(job, err)
	}
	return outcome{status: "done"}
}

// renew re-applies a timeout longer than Discord's 28 day cap until it really ends.
func (s *Sweeper) renew(ctx context.Context, job store.Job) outcome {
	now := s.d.Now()
	if !job.EndsAt.After(now) {
		return outcome{status: "done"}
	}
	until := minTime(job.EndsAt, now.Add(moderation.DiscordTimeout))
	gid, tid := validate.FormatSnowflake(job.GuildID), validate.FormatSnowflake(job.TargetID)
	err := s.d.Client.Timeout(ctx, gid, tid, &until, "Timeout renewed until it ends.")
	switch {
	case discord.Is(err, discord.NotFound):
		// Not in the server right now. Look again later; the timer ends on its own.
		return outcome{next: minTime(job.EndsAt, now.Add(6*time.Hour)), attempts: job.Attempts}
	case err != nil:
		return s.retry(job, err)
	case until.Equal(job.EndsAt):
		return outcome{status: "done"}
	}
	return outcome{next: until.Add(-time.Hour), attempts: 0}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
