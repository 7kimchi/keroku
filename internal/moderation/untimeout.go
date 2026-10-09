package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// UntimeoutCommand is /untimeout.
type UntimeoutCommand struct{ S *Service }

// Definition describes /untimeout.
func (UntimeoutCommand) Definition() *discordgo.ApplicationCommand {
	return definition("untimeout", "Remove a member's timeout", perms.ModerateMembers, userOption("Member to release"), reasonOption(false))
}

// Handle runs /untimeout.
func (c UntimeoutCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	return simple{s: c.S, kind: cases.Untimeout, reasonRequired: false}.handle(ctx, r)
}
