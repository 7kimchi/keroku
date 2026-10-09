package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Channel returns a copy of the channel and its overwrites.
func (f *Fake) Channel(ctx context.Context, channelID string) (*discordgo.Channel, error) {
	if err := f.enter(ctx, "channel"); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channels[channelID]
	if !ok {
		return nil, notFound("channel", 10003)
	}
	cp := *c
	cp.PermissionOverwrites = make([]*discordgo.PermissionOverwrite, len(c.PermissionOverwrites))
	for i, o := range c.PermissionOverwrites {
		oc := *o
		cp.PermissionOverwrites[i] = &oc
	}
	return &cp, nil
}

// Send records an embed posted to a channel.
func (f *Fake) Send(ctx context.Context, channelID string, embed *discordgo.MessageEmbed) error {
	if err := f.enter(ctx, "send"); err != nil {
		return err
	}
	if embed == nil {
		return &Error{Op: "send", Kind: BadRequest}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.channels[channelID]; !ok {
		return notFound("send", 10003)
	}
	f.sent = append(f.sent, Sent{To: channelID, Embed: embed})
	return nil
}

// DM records an embed sent to a user unless their DMs are closed.
func (f *Fake) DM(ctx context.Context, userID string, embed *discordgo.MessageEmbed) error {
	if err := f.enter(ctx, "dm"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.blocked[userID] {
		return &Error{Op: "dm", Kind: Forbidden, Status: 403, Code: 50007}
	}
	f.sent = append(f.sent, Sent{To: "dm:" + userID, Embed: embed})
	return nil
}

// SetSlowmode stores the channel's slowmode.
func (f *Fake) SetSlowmode(ctx context.Context, channelID string, seconds int, _ string) error {
	if err := f.enter(ctx, "slowmode"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channels[channelID]
	if !ok {
		return notFound("slowmode", 10003)
	}
	if seconds < 0 || seconds > 21600 {
		return &Error{Op: "slowmode", Kind: BadRequest, Status: 400}
	}
	c.RateLimitPerUser = seconds
	return nil
}

// SetRolePermissions replaces a role's permissions.
func (f *Fake) SetRolePermissions(ctx context.Context, guildID, roleID string, permissions int64, _ string) error {
	if err := f.enter(ctx, "rolePermissions"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if g, ok := f.guilds[guildID]; ok {
		for _, r := range g.Roles {
			if r.ID == roleID {
				r.Permissions = permissions
				return nil
			}
		}
	}
	return notFound("rolePermissions", 10011)
}
