package moderation

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// BanCommand is /ban.
type BanCommand struct{ S *Service }

// Definition describes /ban.
func (BanCommand) Definition() *discordgo.ApplicationCommand {
	del := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: "delete",
		Description: "Delete the member's recent messages", Choices: []*discordgo.ApplicationCommandOptionChoice{
			{Name: "None", Value: 0}, {Name: "Last hour", Value: 3600}, {Name: "Last 6 hours", Value: 21600},
			{Name: "Last day", Value: 86400}, {Name: "Last 3 days", Value: 259200}, {Name: "Last 7 days", Value: 604800},
		}}
	return definition("ban", "Ban a user, optionally for a limited time", perms.BanMembers,
		userOption("User to ban"), reasonOption(false),
		durationOption("Ban length like 7d or 12h. Leave empty for a permanent ban", false), del)
}

// Handle runs /ban.
func (c BanCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	a, err := fromRequest(r, cases.Ban, false)
	if err != nil {
		return nil, err
	}
	if a.Duration, _, err = r.Duration("duration", time.Minute, MaxBan); err != nil {
		return nil, err
	}
	del, _, err := r.Int("delete", 0, MaxDeleteSeconds)
	if err != nil {
		return nil, err
	}
	a.DeleteSeconds = int(del)
	return c.S.run(ctx, a)
}
