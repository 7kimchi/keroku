package moderation

import (
	"context"
	"errors"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
)

// failure turns an error into what the moderator sees. If Discord already applied the
// action but the case was not saved, the action is reverted when Discord allows it.
func (s *Service) failure(ctx context.Context, a Action, applied bool, err error) error {
	var userErr *commands.UserError
	if errors.As(err, &userErr) {
		return err
	}
	if !applied {
		return discordFailure(a, err)
	}
	return s.compensate(ctx, a, err)
}

// discordFailure explains a failed Discord call in plain words.
func discordFailure(a Action, err error) error {
	switch discord.KindOf(err) {
	case discord.Forbidden:
		return refuse(a, "Discord refused the request. Check Keroku's role position and permissions.")
	case discord.NotFound:
		return refuse(a, "Member or server not found.")
	case discord.RateLimited, discord.Unavailable, discord.Timeout:
		return failWith(a, "Discord did not respond in time. Nothing was changed.", err)
	}
	return failWith(a, "Nothing was changed.", err)
}
