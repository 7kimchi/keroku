package settings

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

// postBits are what Keroku needs in a channel it posts embeds to.
const postBits = perms.ViewChannel | perms.SendMessages | perms.EmbedLinks

func (c ConfigCommand) setChannel(ctx context.Context, r *commands.Request, label string,
	set func(ctx context.Context, guildID, channelID int64) error) (*discordgo.MessageEmbed, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	channel, ok, err := r.Channel("channel")
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := set(ctx, gid, 0); err != nil {
			return nil, err
		}
		return embeds.New(label + " cleared").Build(), nil
	}
	if err := c.checkPostable(ctx, r.GuildID, channel); err != nil {
		return nil, err
	}
	cid, _ := validate.Snowflake(channel)
	if err := set(ctx, gid, cid); err != nil {
		return nil, err
	}
	return embeds.New(label + " set").Description(embeds.Channel(channel)).Build(), nil
}

// checkPostable confirms the channel belongs to this guild, is a text channel and that
// Keroku can post embeds there.
func (c ConfigCommand) checkPostable(ctx context.Context, guildID, channelID string) error {
	ch, err := c.D.Client.Channel(ctx, channelID)
	if discord.Is(err, discord.NotFound) || discord.Is(err, discord.Forbidden) {
		return commands.Fail("Config failed", "Keroku cannot see that channel.")
	}
	if err != nil {
		return err
	}
	if ch.GuildID != guildID || (ch.Type != discordgo.ChannelTypeGuildText && ch.Type != discordgo.ChannelTypeGuildNews) {
		return commands.Fail("Config failed", "Pick a text channel in this server.")
	}
	g, err := c.D.Client.Guild(ctx, guildID)
	if err != nil {
		return err
	}
	bot, err := c.D.Client.Member(ctx, guildID, c.D.BotID)
	if err != nil {
		return err
	}
	if !perms.Has(perms.InChannel(g, ch, c.D.BotID, bot.Roles), postBits) {
		return commands.Fail("Config failed", "Keroku needs View Channel, Send Messages and Embed Links in that channel.")
	}
	return nil
}
