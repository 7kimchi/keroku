// Package info answers the read only lookup commands for servers, users, roles and the bot.
package info

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

// Deps are the info commands' collaborators.
type Deps struct {
	Client  discord.Info
	Started time.Time // process start, for uptime
	Version string
	Now     func() time.Time
}

func (d Deps) now() time.Time {
	if d.Now == nil {
		return time.Now()
	}
	return d.Now()
}

// Anyone who can see the channel can run these. They show what Discord clients already show.
func definition(name, desc string, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: name, Description: desc, Options: opts,
		DefaultMemberPermissions: commands.Perm(perms.ViewChannel), Contexts: commands.GuildOnly()}
}

// unavailable turns a Discord failure into copy. The cause goes to the log with the ref.
func unavailable(title string, err error) error {
	if discord.Is(err, discord.NotFound) {
		return commands.Fail(title, "Not found.")
	}
	return commands.FailWith(title, "Discord did not respond. Try again.", err)
}

// created reads the creation time out of a snowflake id.
func created(id string) string {
	n, err := validate.Snowflake(id)
	if err != nil {
		return "Unknown"
	}
	return embeds.Relative(validate.SnowflakeTime(n))
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func itoa(n int) string { return strconv.Itoa(n) }

// name escapes a user supplied name for a title, with a fallback for empty ones.
func name(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return embeds.Escape(s)
}
