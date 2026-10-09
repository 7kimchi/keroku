package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// WarnCommand is /warn.
type WarnCommand struct{ S *Service }

// Definition describes /warn.
func (WarnCommand) Definition() *discordgo.ApplicationCommand {
	return definition("warn", "Warn a member and record it", perms.ModerateMembers, userOption("Member to warn"), reasonOption(true))
}

// Handle runs /warn.
func (c WarnCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	return simple{s: c.S, kind: cases.Warn, reasonRequired: true}.handle(ctx, r)
}
