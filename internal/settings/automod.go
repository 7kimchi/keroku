package settings

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/automod"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func automodGroup() *discordgo.ApplicationCommandOption {
	on := func(desc string) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionBoolean, Name: "enabled", Description: desc, Required: true}
	}
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommandGroup, Name: "automod",
		Description: "Automatic message moderation", Options: []*discordgo.ApplicationCommandOption{
			sub("spam", "Delete messages sent too fast", on("Turn the rule on or off"),
				intOpt("messages", "Messages that trigger it, 2 to 50", 2, 50), intOpt("seconds", "Window in seconds, 1 to 60", 1, 60)),
			sub("duplicates", "Delete the same text sent again and again", on("Turn the rule on or off"),
				intOpt("count", "Repeats that trigger it, 2 to 20", 2, 20), intOpt("seconds", "Window in seconds, 1 to 600", 1, 600)),
			sub("links", "Delete links outside the allowlist", on("Turn the rule on or off"),
				&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "allow",
					Description: "Allowed domains, comma separated, like youtube.com,github.com", MaxLength: 1500}),
			sub("mentions", "Block messages with too many mentions, using Discord AutoMod",
				intOpt("limit", "Mentions per message, 1 to 50. 0 turns it off", 0, 50)),
			sub("invites", "Block invite links, using Discord AutoMod", on("Turn the rule on or off")),
			sub("timeout", "Also time out members who trigger a rule",
				&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "duration",
					Description: "Like 10m or 1h, up to 28d. off turns it off", Required: true, MaxLength: 32}),
		}}
}

// automod applies one automod subcommand to the stored settings and syncs native rules.
func (c ConfigCommand) automod(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	cfg, err := c.D.Store.Automod(ctx, gid)
	if err != nil {
		return nil, err
	}
	if err := applyAutomod(r, &cfg); err != nil {
		return nil, err
	}
	if err := c.D.Store.SetAutomod(ctx, gid, cfg); err != nil {
		return nil, err
	}
	c.D.Automod.Invalidate(gid)
	if strings.HasSuffix(r.Sub, "mentions") || strings.HasSuffix(r.Sub, "invites") || strings.HasSuffix(r.Sub, "timeout") {
		if err := automod.SyncNative(ctx, c.D.Client, r.GuildID, c.D.BotID, cfg); err != nil {
			return nil, nativeFailure(err)
		}
	}
	return embeds.New("Automod updated").Description(automodSummary(cfg)).Build(), nil
}

func applyAutomod(r *commands.Request, cfg *store.AutomodSettings) error {
	enabled, _, err := r.Bool("enabled")
	if err != nil {
		return err
	}
	switch r.Sub {
	case "automod spam":
		cfg.SpamEnabled = enabled
		return windowed(r, "messages", &cfg.SpamMessages, &cfg.SpamWindow, 50, 60)
	case "automod duplicates":
		cfg.DuplicateEnabled = enabled
		return windowed(r, "count", &cfg.DuplicateCount, &cfg.DuplicateWindow, 20, 600)
	case "automod links":
		cfg.LinksEnabled = enabled
		return allowlist(r, cfg)
	case "automod mentions":
		n, _, err := r.Int("limit", 0, 50)
		cfg.MentionLimit = int(n)
		return err
	case "automod invites":
		cfg.InvitesBlocked = enabled
	case "automod timeout":
		return timeoutSetting(r, cfg)
	}
	return nil
}
