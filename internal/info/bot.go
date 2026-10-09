package info

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
)

// BotInfoCommand is /botinfo.
type BotInfoCommand struct{ D Deps }

// Definition describes /botinfo.
func (BotInfoCommand) Definition() *discordgo.ApplicationCommand {
	return definition("botinfo", "Show Keroku's version and uptime")
}

// Handle runs /botinfo. The server count call doubles as the latency probe.
func (c BotInfoCommand) Handle(ctx context.Context, _ *commands.Request) (*discordgo.MessageEmbed, error) {
	start := c.D.now()
	n, err := c.D.Client.ServerCount(ctx)
	if err != nil {
		return nil, unavailable("Bot info failed", err)
	}
	now := c.D.now()
	version := c.D.Version
	if version == "" {
		version = "dev"
	}
	latency := max(now.Sub(start).Milliseconds(), 0)
	return embeds.New("Keroku").
		Field("Version", embeds.Escape(version), true).
		Field("Uptime", embeds.Duration(now.Sub(c.D.Started)), true).
		Field("Servers", itoa(n), true).
		Field("Latency", strconv.FormatInt(latency, 10)+" ms", true).
		Build(), nil
}
