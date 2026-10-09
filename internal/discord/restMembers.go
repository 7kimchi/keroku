package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Guild fetches the guild with its roles.
func (r *REST) Guild(ctx context.Context, guildID string) (g *discordgo.Guild, err error) {
	err = r.call(ctx, "guild", true, func(o ...discordgo.RequestOption) (e error) {
		g, e = r.s.Guild(guildID, o...)
		return
	})
	return g, err
}

// Member fetches a member fresh. A non member returns a NotFound error.
func (r *REST) Member(ctx context.Context, guildID, userID string) (m *discordgo.Member, err error) {
	err = r.call(ctx, "member", true, func(o ...discordgo.RequestOption) (e error) {
		m, e = r.s.GuildMember(guildID, userID, o...)
		return
	})
	return m, err
}

// Ban bans a user and deletes their messages from the last deleteSeconds.
func (r *REST) Ban(ctx context.Context, guildID, userID string, deleteSeconds int, reason string) error {
	body := map[string]int{"delete_message_seconds": deleteSeconds}
	return r.call(ctx, "ban", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.RequestWithBucketID("PUT", discordgo.EndpointGuildBan(guildID, userID), body,
			discordgo.EndpointGuildBan(guildID, ""), with(o, reason)...)
		return e
	})
}

// Unban lifts a ban.
func (r *REST) Unban(ctx context.Context, guildID, userID, reason string) error {
	return r.call(ctx, "unban", true, func(o ...discordgo.RequestOption) error {
		return r.s.GuildBanDelete(guildID, userID, with(o, reason)...)
	})
}

// IsBanned reports whether the user is banned.
func (r *REST) IsBanned(ctx context.Context, guildID, userID string) (bool, error) {
	err := r.call(ctx, "getBan", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.GuildBan(guildID, userID, o...)
		return e
	})
	if Is(err, NotFound) {
		return false, nil
	}
	return err == nil, err
}

// Kick removes a member.
func (r *REST) Kick(ctx context.Context, guildID, userID, reason string) error {
	return r.call(ctx, "kick", true, func(o ...discordgo.RequestOption) error {
		return r.s.GuildMemberDeleteWithReason(guildID, userID, "", with(o, reason)...)
	})
}

// Timeout sets or clears (nil until) a member's timeout.
func (r *REST) Timeout(ctx context.Context, guildID, userID string, until *time.Time, reason string) error {
	return r.call(ctx, "timeout", true, func(o ...discordgo.RequestOption) error {
		return r.s.GuildMemberTimeout(guildID, userID, until, with(o, reason)...)
	})
}
