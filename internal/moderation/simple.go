package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
)

// simple is a command that only needs a user and a reason.
type simple struct {
	s              *Service
	kind           cases.Kind
	reasonRequired bool
}

func (c simple) handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	a, err := fromRequest(r, c.kind, c.reasonRequired)
	if err != nil {
		return nil, err
	}
	return c.s.run(ctx, a)
}
