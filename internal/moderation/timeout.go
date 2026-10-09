package moderation

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// TimeoutCommand is /timeout. Lengths past Discord's 28 days are renewed automatically.
type TimeoutCommand struct{ S *Service }

// Definition describes /timeout.
func (TimeoutCommand) Definition() *discordgo.ApplicationCommand {
	return definition("timeout", "Stop a member from talking for a while", perms.ModerateMembers,
		userOption("Member to time out"), durationOption("Length like 10m, 2h or 7d, up to 365d", true), reasonOption(false))
}

// Handle runs /timeout.
func (c TimeoutCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	a, err := fromRequest(r, cases.Timeout, false)
	if err != nil {
		return nil, err
	}
	d, ok, err := r.Duration("duration", time.Minute, MaxTimeout)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, commands.Fail("Timeout failed", "A duration is required.")
	}
	a.Duration = d
	return c.S.run(ctx, a)
}
