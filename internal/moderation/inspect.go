package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

type state struct {
	guild  *discordgo.Guild
	member *discordgo.Member // nil when the target is not in the guild
}

// inspect fetches the guild, the target and the bot fresh over REST, runs the hierarchy
// check and confirms the target is in a state the action applies to.
func (s *Service) inspect(ctx context.Context, a Action) (state, error) {
	f, err := s.fetch(ctx, a)
	if err != nil {
		return state{}, err
	}
	g, member, tid := f.guild, f.member, validate.FormatSnowflake(a.TargetID)
	req := perms.Request{
		Guild: g, Need: need(a.Kind), Automated: a.Automated,
		InvokerID: validate.FormatSnowflake(a.ModeratorID), InvokerRoles: a.InvokerRoles, InvokerPerms: a.InvokerPerms,
		BotID: s.botID, BotRoles: f.bot.Roles, BotPerms: a.BotPerms, TargetID: tid,
	}
	if a.Automated || req.BotPerms == 0 {
		req.BotPerms = perms.Base(g, s.botID, f.bot.Roles)
	}
	if member != nil {
		req.TargetRoles, req.TargetMember = member.Roles, true
	}
	if d := perms.Check(req); d != nil {
		return state{}, refuse(a, d.Message())
	}
	return state{guild: g, member: member}, s.stateCheck(a, g, f)
}

func (s *Service) stateCheck(a Action, g *discordgo.Guild, f fetched) error {
	m := f.member
	switch a.Kind {
	case cases.Kick, cases.Timeout, cases.Untimeout:
		if m == nil {
			return refuse(a, "Not a member of this server.")
		}
	}
	switch a.Kind {
	case cases.Timeout:
		if perms.Base(g, validate.FormatSnowflake(a.TargetID), m.Roles)&perms.Administrator != 0 {
			return refuse(a, "Target has Administrator, which timeouts do not affect.")
		}
	case cases.Untimeout:
		if m.CommunicationDisabledUntil == nil || !m.CommunicationDisabledUntil.After(s.now()) {
			return refuse(a, "Not timed out.")
		}
	case cases.Ban, cases.Unban:
		if f.banErr != nil {
			return f.banErr
		}
		if a.Kind == cases.Ban && f.banned {
			return refuse(a, "Already banned.")
		}
		if a.Kind == cases.Unban && !f.banned {
			return refuse(a, NotBanned)
		}
	}
	return nil
}

func refuse(a Action, detail string) error {
	return commands.Fail(verb(a.Kind)+" failed", detail)
}
