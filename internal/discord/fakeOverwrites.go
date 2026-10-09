package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// SetRoleOverwrite replaces a role overwrite on a channel.
func (f *Fake) SetRoleOverwrite(ctx context.Context, channelID, roleID string, allow, deny int64, _ string) error {
	if err := f.enter(ctx, "setOverwrite"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channels[channelID]
	if !ok {
		return notFound("setOverwrite", 10003)
	}
	for _, o := range c.PermissionOverwrites {
		if o.ID == roleID {
			o.Allow, o.Deny = allow, deny
			return nil
		}
	}
	c.PermissionOverwrites = append(c.PermissionOverwrites, &discordgo.PermissionOverwrite{
		ID: roleID, Type: discordgo.PermissionOverwriteTypeRole, Allow: allow, Deny: deny,
	})
	return nil
}

// DeleteOverwrite removes an overwrite. Missing overwrites are fine, like Discord.
func (f *Fake) DeleteOverwrite(ctx context.Context, channelID, targetID, _ string) error {
	if err := f.enter(ctx, "deleteOverwrite"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channels[channelID]
	if !ok {
		return notFound("deleteOverwrite", 10003)
	}
	kept := c.PermissionOverwrites[:0]
	for _, o := range c.PermissionOverwrites {
		if o.ID != targetID {
			kept = append(kept, o)
		}
	}
	c.PermissionOverwrites = kept
	return nil
}
