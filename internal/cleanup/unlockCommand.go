package cleanup

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
)

// UnlockCommand is /unlock.
type UnlockCommand struct{ S *Service }

// Definition describes /unlock.
func (UnlockCommand) Definition() *discordgo.ApplicationCommand {
	return definition("unlock", "Undo a channel lock", perms.ManageChannels,
		channelOption("Channel to unlock. Defaults to this one"), reasonOption())
}

// Handle runs /unlock.
func (c UnlockCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	channel, err := target(r)
	if err != nil {
		return nil, err
	}
	why, err := reason(r)
	if err != nil {
		return nil, err
	}
	gid, _ := strconv.ParseInt(r.GuildID, 10, 64)
	cid, _ := strconv.ParseInt(channel, 10, 64)
	found, err := c.S.Unlock(ctx, gid, cid, audit(r, why))
	if err != nil {
		return nil, explain("Unlock failed", "Manage Roles", err)
	}
	if !found {
		return nil, commands.Fail("Unlock failed", "Channel is not locked.")
	}
	c.S.modlog.Case(gid, event("Channel unlocked", r, [2]string{"Channel", embeds.Channel(channel)}))
	return embeds.New("Channel unlocked").Field("Channel", embeds.Channel(channel), true).Build(), nil
}
