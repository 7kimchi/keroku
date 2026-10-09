package cleanup

import (
	"context"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
)

// LockdownCommand is /lockdown with start and end subcommands.
type LockdownCommand struct{ S *Service }

// Definition describes /lockdown.
func (LockdownCommand) Definition() *discordgo.ApplicationCommand {
	sub := func(name, desc string, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommand, Name: name, Description: desc, Options: opts}
	}
	return definition("lockdown", "Stop everyone from sending messages server wide", perms.ManageGuild,
		sub("start", "Start a lockdown", durationOption("End automatically after this long, like 1h"), reasonOption()),
		sub("end", "End the lockdown", reasonOption()))
}

// Handle runs /lockdown.
func (c LockdownCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	why, err := reason(r)
	if err != nil {
		return nil, err
	}
	gid, _ := strconv.ParseInt(r.GuildID, 10, 64)
	if r.Sub == "end" {
		found, err := c.S.EndLockdown(ctx, gid, audit(r, why))
		if err != nil {
			return nil, explain("Lockdown end failed", "Manage Roles", err)
		}
		if !found {
			return nil, commands.Fail("Lockdown end failed", "No lockdown is active.")
		}
		c.S.modlog.Case(gid, event("Lockdown ended", r))
		return embeds.New("Lockdown ended").Build(), nil
	}
	d, _, err := r.Duration("duration", time.Minute, MaxLockFor)
	if err != nil {
		return nil, err
	}
	var until time.Time
	if d > 0 {
		until = c.S.now().Add(d)
	}
	if err := c.S.Lockdown(ctx, gid, audit(r, why), until); err != nil {
		return nil, explain("Lockdown failed", "Manage Roles", err)
	}
	var fields [][2]string
	if d > 0 {
		fields = append(fields, [2]string{"Ends", embeds.Relative(until)})
	}
	c.S.modlog.Case(gid, event("Lockdown started", r, fields...))
	return embeds.New("Lockdown started").Description("Members cannot send messages until the lockdown ends.").Build(), nil
}
