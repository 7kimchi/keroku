package settings

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func raidGroup() *discordgo.ApplicationCommandOption {
	str := func(name, desc string) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: name, Description: desc, MaxLength: 32}
	}
	action := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "action",
		Description: "What happens to a raid", Choices: []*discordgo.ApplicationCommandOptionChoice{
			{Name: "Kick joiners", Value: "kick"}, {Name: "Ban joiners", Value: "ban"}, {Name: "Lock the server down", Value: "lockdown"}}}
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommandGroup, Name: "raid",
		Description: "Join burst and account age protection", Options: []*discordgo.ApplicationCommandOption{
			sub("set", "Change raid protection. Options left empty keep their value",
				&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionBoolean, Name: "enabled",
					Description: "Turn raid protection on or off"},
				intOpt("joins", "Joins that count as a raid, 2 to 500", 2, 500),
				intOpt("seconds", "Window for those joins in seconds, 1 to 600", 1, 600),
				str("minage", "Remove accounts younger than this, like 7d. off turns it off"),
				action, str("lockdown", "How long a raid lockdown lasts, like 15m, 1m to 7d")),
		}}
}

func (c ConfigCommand) raid(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	cfg, err := c.D.Store.Raid(ctx, gid)
	if err != nil {
		return nil, err
	}
	if err := applyRaid(r, &cfg); err != nil {
		return nil, err
	}
	if err := c.D.Store.SetRaid(ctx, gid, cfg); err != nil {
		return nil, err
	}
	c.D.Raid.Invalidate(gid)
	return embeds.New("Raid protection updated").Description(raidSummary(cfg)).Build(), nil
}

func applyRaid(r *commands.Request, cfg *store.RaidSettings) error {
	if on, ok, err := r.Bool("enabled"); err != nil {
		return err
	} else if ok {
		cfg.Enabled = on
	}
	var seconds int64
	if err := firstErr(intInto(r, "joins", 2, 500, &cfg.JoinLimit), int64Into(r, "seconds", 1, 600, &seconds)); err != nil {
		return err
	}
	if seconds > 0 {
		cfg.Window = time.Duration(seconds) * time.Second
	}
	if raw, ok, err := r.Text("minage", 32); err != nil {
		return err
	} else if ok {
		if strings.EqualFold(raw, "off") {
			cfg.MinAccountAge = 0
		} else if cfg.MinAccountAge, err = validate.Duration(raw, time.Minute, 365*24*time.Hour); err != nil {
			return commands.Fail("Config failed", "Minimum age must be like 7d, up to 365d, or off.")
		}
	}
	if a, ok, err := r.Text("action", 16); err != nil {
		return err
	} else if ok {
		if a != "kick" && a != "ban" && a != "lockdown" {
			return commands.Fail("Config failed", "Action must be kick, ban or lockdown.")
		}
		cfg.Action = a
	}
	d, ok, err := r.Duration("lockdown", time.Minute, 7*24*time.Hour)
	if err != nil {
		return err
	}
	if ok {
		cfg.LockdownFor = d
	}
	return nil
}
