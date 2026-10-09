package cleanup

import (
	"errors"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
)

// explain turns a Discord failure into copy the moderator can act on.
func explain(title, permission string, err error) error {
	var userErr *commands.UserError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &userErr):
		return err
	case discord.Is(err, discord.Forbidden):
		return commands.Fail(title, "Keroku is missing permission: "+permission+".")
	case discord.Is(err, discord.NotFound):
		return commands.Fail(title, "Channel not found.")
	case discord.Is(err, discord.RateLimited), discord.Is(err, discord.Unavailable), discord.Is(err, discord.Timeout):
		return commands.FailWith(title, "Discord did not respond in time. Try again.", err)
	}
	return err
}
