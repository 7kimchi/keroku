package moderation

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/validate"
)

const compensateTimeout = 10 * time.Second

// apply performs the Discord side of an action. Warnings and notes have none.
func (s *Service) apply(ctx context.Context, a Action) error {
	gid, tid := validate.FormatSnowflake(a.GuildID), validate.FormatSnowflake(a.TargetID)
	reason := auditReason(a)
	switch a.Kind {
	case cases.Ban:
		return s.client.Ban(ctx, gid, tid, a.DeleteSeconds, reason)
	case cases.Unban:
		return s.client.Unban(ctx, gid, tid, reason)
	case cases.Kick:
		return s.client.Kick(ctx, gid, tid, reason)
	case cases.Timeout:
		until := s.now().Add(min(a.Duration, DiscordTimeout))
		return s.client.Timeout(ctx, gid, tid, &until, reason)
	case cases.Untimeout:
		return s.client.Timeout(ctx, gid, tid, nil, reason)
	}
	return nil
}

// auditReason names the moderator in Discord's audit log, since the bot is the actor there.
func auditReason(a Action) string {
	r := a.Reason
	if r == "" {
		r = "No reason given."
	}
	return "Moderator " + validate.FormatSnowflake(a.ModeratorID) + ": " + r
}

// compensate undoes an applied action whose case could not be saved.
func (s *Service) compensate(ctx context.Context, a Action, cause error) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), compensateTimeout)
	defer cancel()
	gid, tid := validate.FormatSnowflake(a.GuildID), validate.FormatSnowflake(a.TargetID)
	const why = "Reverted: the case could not be saved."
	var err error
	switch a.Kind {
	case cases.Ban:
		err = s.client.Unban(ctx, gid, tid, why)
	case cases.Unban:
		err = s.client.Ban(ctx, gid, tid, 0, why)
	case cases.Timeout:
		err = s.client.Timeout(ctx, gid, tid, nil, why)
	case cases.Kick, cases.Untimeout:
		s.log.Error("action applied without a case", "kind", string(a.Kind), "guildId", a.GuildID, "targetId", a.TargetID, "err", cause)
		return incomplete(a, "Applied in Discord, but the case could not be saved.", cause)
	default:
		return failWith(a, "The case could not be saved.", cause)
	}
	if err != nil {
		s.log.Error("compensation failed", "kind", string(a.Kind), "guildId", a.GuildID, "targetId", a.TargetID, "err", err, "cause", cause)
		return incomplete(a, "Applied in Discord, but the case could not be saved and reverting failed.", cause)
	}
	return failWith(a, "The case could not be saved, so the action was reverted.", cause)
}

// discordFailure explains a failed Discord call in plain words.
func discordFailure(a Action, err error) error {
	switch discord.KindOf(err) {
	case discord.Forbidden:
		return refuse(a, "Discord refused the request. Check Keroku's role position and permissions.")
	case discord.NotFound:
		return refuse(a, "Member or server not found.")
	case discord.RateLimited, discord.Unavailable, discord.Timeout:
		return failWith(a, "Discord did not respond in time. Nothing was changed.", err)
	}
	return failWith(a, "Nothing was changed.", err)
}
