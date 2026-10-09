package info

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// ServerInfoCommand is /serverinfo.
type ServerInfoCommand struct{ D Deps }

// Definition describes /serverinfo.
func (ServerInfoCommand) Definition() *discordgo.ApplicationCommand {
	return definition("serverinfo", "Show this server's details")
}

// Handle runs /serverinfo.
func (c ServerInfoCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	g, err := c.D.Client.GuildCounts(ctx, r.GuildID)
	if err != nil {
		return nil, unavailable("Server info failed", err)
	}
	return serverEmbed(g, r.GuildID), nil
}

// serverEmbed renders the guild. id is the guild the command ran in, not the one in the reply.
func serverEmbed(g *discordgo.Guild, id string) *discordgo.MessageEmbed {
	owner := "Unknown"
	if _, err := validate.Snowflake(g.OwnerID); err == nil {
		owner = embeds.User(g.OwnerID)
	}
	return embeds.New(name(g.Name, "Server")).
		Field("Owner", owner, true).
		Field("Created", created(id), true).
		Field("Members", itoa(max(g.ApproximateMemberCount, 0)), true).
		Field("Online", itoa(max(g.ApproximatePresenceCount, 0)), true).
		Field("Roles", itoa(len(g.Roles)), true).
		Field("Emojis", itoa(len(g.Emojis)), true).
		Field("Boost level", itoa(max(int(g.PremiumTier), 0)), true).
		Field("Boosts", itoa(max(g.PremiumSubscriptionCount, 0)), true).
		Field("Verification", verification(g.VerificationLevel), true).
		Footer("ID " + id).
		Build()
}

func verification(v discordgo.VerificationLevel) string {
	switch v {
	case discordgo.VerificationLevelNone:
		return "None"
	case discordgo.VerificationLevelLow:
		return "Low"
	case discordgo.VerificationLevelMedium:
		return "Medium"
	case discordgo.VerificationLevelHigh:
		return "High"
	case discordgo.VerificationLevelVeryHigh:
		return "Highest"
	}
	return "Unknown"
}
