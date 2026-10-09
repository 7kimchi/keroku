package cleanup

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

// MaxSlowmode is Discord's limit.
const MaxSlowmode = 6 * time.Hour

// SlowmodeCommand is /slowmode.
type SlowmodeCommand struct{ S *Service }

// Definition describes /slowmode.
func (SlowmodeCommand) Definition() *discordgo.ApplicationCommand {
	return definition("slowmode", "Set how often each member can send messages", perms.ManageChannels,
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "interval",
			Description: "Like 10s or 5m, up to 6h. Use off to turn it off", Required: true, MaxLength: 32},
		channelOption("Channel to change. Defaults to this one"), reasonOption())
}

// Handle runs /slowmode.
func (c SlowmodeCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	raw, _, err := r.Text("interval", 32)
	if err != nil {
		return nil, err
	}
	var d time.Duration
	if !strings.EqualFold(raw, "off") {
		if d, err = validate.Duration(raw, 0, MaxSlowmode); err != nil {
			return nil, commands.Fail("Slowmode failed", "Interval must be like 10s or 5m, at most 6h, or off.")
		}
	}
	channel, err := target(r)
	if err != nil {
		return nil, err
	}
	if err := c.S.checkChannel(ctx, r, channel, "Slowmode failed", perms.ManageChannels); err != nil {
		return nil, err
	}
	why, err := reason(r)
	if err != nil {
		return nil, err
	}
	seconds := int(d / time.Second)
	if err := c.S.client.SetSlowmode(ctx, channel, seconds, audit(r, why)); err != nil {
		return nil, explain("Slowmode failed", "Manage Channels", err)
	}
	value := "Off"
	if seconds > 0 {
		value = embeds.Duration(d)
	}
	gid, _ := strconv.ParseInt(r.GuildID, 10, 64)
	c.S.modlog.Case(gid, event("Slowmode changed", r, [2]string{"Channel", embeds.Channel(channel)}, [2]string{"Interval", value}))
	return embeds.New("Slowmode changed").Field("Channel", embeds.Channel(channel), true).Field("Interval", value, true).Build(), nil
}
