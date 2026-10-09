package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// UnbanCommand is /unban.
type UnbanCommand struct{ S *Service }

// Definition describes /unban.
func (UnbanCommand) Definition() *discordgo.ApplicationCommand {
	return definition("unban", "Unban a user", perms.BanMembers, userOption("User to unban"), reasonOption(false))
}

// Handle runs /unban.
func (c UnbanCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	return simple{s: c.S, kind: cases.Unban, reasonRequired: false}.handle(ctx, r)
}
