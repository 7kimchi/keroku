package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
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
	gid, tid := validate.FormatSnowflake(a.GuildID), validate.FormatSnowflake(a.TargetID)
	g, err := s.client.Guild(ctx, gid)
	if err != nil {
		return state{}, err
	}
	member, err := s.client.Member(ctx, gid, tid)
	if err != nil && !discord.Is(err, discord.NotFound) {
		return state{}, err
	}
	bot, err := s.client.Member(ctx, gid, s.botID)
	if err != nil {
		return state{}, err
	}
	req := perms.Request{
		Guild: g, Need: need(a.Kind), Automated: a.Automated,
		InvokerID: validate.FormatSnowflake(a.ModeratorID), InvokerRoles: a.InvokerRoles, InvokerPerms: a.InvokerPerms,
		BotID: s.botID, BotRoles: bot.Roles, BotPerms: a.BotPerms, TargetID: tid,
	}
	if a.Automated || req.BotPerms == 0 {
		req.BotPerms = perms.Base(g, s.botID, bot.Roles)
	}
	if member != nil {
		req.TargetRoles, req.TargetMember = member.Roles, true
	}
	if d := perms.Check(req); d != nil {
		return state{}, refuse(a, d.Message())
	}
	return state{guild: g, member: member}, s.stateCheck(ctx, a, g, member)
}

func (s *Service) stateCheck(ctx context.Context, a Action, g *discordgo.Guild, m *discordgo.Member) error {
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
		banned, err := s.client.IsBanned(ctx, g.ID, validate.FormatSnowflake(a.TargetID))
		if err != nil {
			return err
		}
		if a.Kind == cases.Ban && banned {
			return refuse(a, "Already banned.")
		}
		if a.Kind == cases.Unban && !banned {
			return refuse(a, NotBanned)
		}
	}
	return nil
}

func refuse(a Action, detail string) error {
	return commands.Fail(verb(a.Kind)+" failed", detail)
}
