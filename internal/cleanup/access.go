package cleanup

import (
	"context"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

// checkChannel confirms the invoker holds need in a channel other than the one the command
// ran in. Discord only computes permissions for the command's own channel, so overwrites
// on the target channel are applied here from fresh data.
func (s *Service) checkChannel(ctx context.Context, r *commands.Request, channelID, title string, need int64) error {
	if channelID == r.ChannelID {
		return nil
	}
	ch, err := s.client.Channel(ctx, channelID)
	if err != nil {
		return explain(title, "View Channel", err)
	}
	if ch.GuildID != r.GuildID {
		return commands.Fail(title, "Channel is not in this server.")
	}
	g, err := s.client.Guild(ctx, r.GuildID)
	if err != nil {
		return err
	}
	if !perms.Has(perms.InChannel(g, ch, r.UserID, r.Member.Roles), need) {
		return commands.Fail(title, "You need "+perms.Name(need)+" in that channel.")
	}
	return nil
}
