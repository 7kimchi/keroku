package info

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
)

// UserInfoCommand is /userinfo. It reads what Discord sends with the interaction, so it
// makes no API calls.
type UserInfoCommand struct{ D Deps }

// Definition describes /userinfo.
func (UserInfoCommand) Definition() *discordgo.ApplicationCommand {
	return definition("userinfo", "Show a user's account and membership details",
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionUser, Name: "user",
			Description: "User to look up. Defaults to you"})
}

// Handle runs /userinfo.
func (c UserInfoCommand) Handle(_ context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	id, ok, err := r.User("user")
	if err != nil {
		return nil, err
	}
	if !ok {
		return userEmbed(r.Member.User, r.Member, r.UserID, c.D.now()), nil
	}
	u, found := r.ResolvedUser(id)
	if !found {
		return nil, commands.Fail("User info failed", "User not found.")
	}
	return userEmbed(u, r.ResolvedMember(id), id, c.D.now()), nil
}

// userEmbed renders a user and, when they are in the server, their membership. m may be nil.
func userEmbed(u *discordgo.User, m *discordgo.Member, id string, now time.Time) *discordgo.MessageEmbed {
	b := embeds.New(name(u.Username, "User")).
		Field("User", embeds.User(id), true).
		Field("Created", created(id), true)
	if u.GlobalName != "" {
		b.Field("Display name", embeds.Escape(u.GlobalName), true)
	}
	if u.Bot {
		b.Field("Bot", "Yes", true)
	}
	if m == nil {
		return b.Field("Member", "Not in this server", false).Build()
	}
	if m.Nick != "" {
		b.Field("Nickname", embeds.Escape(m.Nick), true)
	}
	if !m.JoinedAt.IsZero() {
		b.Field("Joined", embeds.Relative(m.JoinedAt), true)
	}
	if t := m.CommunicationDisabledUntil; t != nil && t.After(now) {
		b.Field("Timed out until", embeds.Relative(*t), true)
	}
	return b.Field("Roles ("+itoa(len(m.Roles))+")", mentionList("@&", m.Roles, 1024), false).Build()
}
