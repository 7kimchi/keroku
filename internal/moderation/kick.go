package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// KickCommand is /kick.
type KickCommand struct{ S *Service }

// Definition describes /kick.
func (KickCommand) Definition() *discordgo.ApplicationCommand {
	return definition("kick", "Remove a member from the server", perms.KickMembers, userOption("Member to kick"), reasonOption(false))
}

// Handle runs /kick.
func (c KickCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	return simple{s: c.S, kind: cases.Kick, reasonRequired: false}.handle(ctx, r)
}
