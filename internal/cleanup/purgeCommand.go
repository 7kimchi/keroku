package cleanup

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
)

// PurgeCommand is /purge.
type PurgeCommand struct{ S *Service }

// Definition describes /purge.
func (PurgeCommand) Definition() *discordgo.ApplicationCommand {
	return definition("purge", "Delete recent messages in this channel", perms.ManageMessages,
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: "count",
			Description: "How many messages, up to 500", Required: true, MinValue: ptr(1), MaxValue: MaxPurge},
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionUser, Name: "user",
			Description: "Only delete this user's messages"}, reasonOption())
}

func ptr(f float64) *float64 { return &f }

// Handle runs /purge.
func (c PurgeCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	count, ok, err := r.Int("count", 1, MaxPurge)
	if err != nil || !ok {
		return nil, orFail(err, "Purge failed", "A count is required.")
	}
	author, _, err := r.User("user")
	if err != nil {
		return nil, err
	}
	why, err := reason(r)
	if err != nil {
		return nil, err
	}
	if err := needBot(r, "Purge failed", perms.ManageMessages|perms.ReadHistory, "Manage Messages and Read Message History"); err != nil {
		return nil, err
	}
	if fresh, err := c.S.claim(ctx, r); err != nil || !fresh {
		return replay(), err
	}
	res, err := c.S.Purge(ctx, r.ChannelID, int(count), author, audit(r, why))
	if err != nil && res.Deleted == 0 {
		return nil, explain("Purge failed", "Manage Messages", err)
	}
	gid, _ := strconv.ParseInt(r.GuildID, 10, 64)
	c.S.modlog.Case(gid, event("Messages purged", r, [2]string{"Channel", embeds.Channel(r.ChannelID)},
		[2]string{"Deleted", strconv.Itoa(res.Deleted)}))
	b := embeds.New("Messages deleted").Field("Deleted", strconv.Itoa(res.Deleted), true)
	if res.Stopped {
		b.Description("Messages older than 14 days cannot be bulk deleted and were kept.")
	}
	if err != nil {
		b.Description("Discord stopped responding partway. Run the command again for the rest.")
	}
	return b.Build(), nil
}

func orFail(err error, title, detail string) error {
	if err != nil {
		return err
	}
	return commands.Fail(title, detail)
}
