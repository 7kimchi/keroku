// Package settings serves /config: channels, warn escalation, automod and raid settings.
package settings

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

// Deps are what /config needs.
type Deps struct {
	Store    *store.Store
	Settings *store.SettingsCache
	Client   discord.Client
	BotID    string
}

// ConfigCommand is /config.
type ConfigCommand struct{ D Deps }

// Definition describes /config.
func (ConfigCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: "config", Description: "View or change Keroku's settings for this server",
		DefaultMemberPermissions: commands.Perm(perms.ManageGuild), Contexts: commands.GuildOnly(),
		Options: []*discordgo.ApplicationCommandOption{
			sub("view", "Show every setting"),
			sub("modlog", "Set the channel for case posts. Leave empty to turn them off", channelOption()),
			sub("logs", "Set the channel for message and member logs. Leave empty to turn them off", channelOption()),
			escalationGroup(),
		}}
}

// Handle routes /config subcommands.
func (c ConfigCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	switch r.Sub {
	case "view":
		return c.view(ctx, r)
	case "modlog":
		return c.setChannel(ctx, r, "Modlog channel", c.D.Settings.SetModlog)
	case "logs":
		return c.setChannel(ctx, r, "Log channel", c.D.Settings.SetLog)
	case "escalation add":
		return c.escalationAdd(ctx, r)
	case "escalation remove":
		return c.escalationRemove(ctx, r)
	}
	return nil, commands.Fail("Config failed", "Unknown subcommand.")
}

func sub(name, desc string, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommand, Name: name,
		Description: desc, Options: opts}
}

func channelOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel",
		Description: "Text channel", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews}}
}
