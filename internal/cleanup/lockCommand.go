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

// LockCommand is /lock.
type LockCommand struct{ S *Service }

// Definition describes /lock.
func (LockCommand) Definition() *discordgo.ApplicationCommand {
	return definition("lock", "Stop members from sending messages in a channel", perms.ManageChannels,
		channelOption("Channel to lock. Defaults to this one"),
		durationOption("Unlock automatically after this long, like 30m"), reasonOption())
}

// Handle runs /lock.
func (c LockCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	channel, err := target(r)
	if err != nil {
		return nil, err
	}
	if err := c.S.checkChannel(ctx, r, channel, "Lock failed", perms.ManageChannels); err != nil {
		return nil, err
	}
	d, _, err := r.Duration("duration", time.Minute, MaxLockFor)
	if err != nil {
		return nil, err
	}
	why, err := reason(r)
	if err != nil {
		return nil, err
	}
	var until time.Time
	if d > 0 {
		until = c.S.now().Add(d)
	}
	gid, _ := strconv.ParseInt(r.GuildID, 10, 64)
	cid, _ := strconv.ParseInt(channel, 10, 64)
	if err := c.S.Lock(ctx, gid, cid, audit(r, why), until); err != nil {
		return nil, explain("Lock failed", "Manage Roles", err)
	}
	fields := [][2]string{{"Channel", embeds.Channel(channel)}}
	if d > 0 {
		fields = append(fields, [2]string{"Unlocks", embeds.Relative(until)})
	}
	c.S.modlog.Case(gid, event("Channel locked", r, fields...))
	b := embeds.New("Channel locked")
	for _, f := range fields {
		b.Field(f[0], f[1], true)
	}
	return b.Build(), nil
}
