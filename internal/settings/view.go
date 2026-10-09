package settings

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

func (c ConfigCommand) view(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	g, err := c.D.Settings.Get(ctx, gid)
	if err != nil {
		return nil, err
	}
	steps, err := c.D.Store.EscalationSteps(ctx, gid)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(steps))
	for _, s := range steps {
		lines = append(lines, stepText(s))
	}
	esc := strings.Join(lines, "\n")
	if esc == "" {
		esc = "Off"
	}
	am, err := c.D.Store.Automod(ctx, gid)
	if err != nil {
		return nil, err
	}
	rd, err := c.D.Store.Raid(ctx, gid)
	if err != nil {
		return nil, err
	}
	return embeds.New("Settings").
		Field("Modlog channel", channelText(g.ModlogChannelID), true).
		Field("Log channel", channelText(g.LogChannelID), true).
		Field("Warn escalation", esc, false).
		Field("Automod", automodSummary(am), false).
		Field("Raid protection", raidSummary(rd), false).Build(), nil
}

func channelText(id int64) string {
	if id == 0 {
		return "Off"
	}
	return embeds.Channel(validate.FormatSnowflake(id))
}
