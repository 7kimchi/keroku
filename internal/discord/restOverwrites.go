package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// SetRoleOverwrite replaces a role's overwrite on a channel.
func (r *REST) SetRoleOverwrite(ctx context.Context, channelID, roleID string, allow, deny int64, reason string) error {
	return r.call(ctx, "setOverwrite", true, func(o ...discordgo.RequestOption) error {
		return r.s.ChannelPermissionSet(channelID, roleID, discordgo.PermissionOverwriteTypeRole, allow, deny, with(o, reason)...)
	})
}

// DeleteOverwrite removes an overwrite from a channel.
func (r *REST) DeleteOverwrite(ctx context.Context, channelID, targetID, reason string) error {
	return r.call(ctx, "deleteOverwrite", true, func(o ...discordgo.RequestOption) error {
		return r.s.ChannelPermissionDelete(channelID, targetID, with(o, reason)...)
	})
}

// SetRolePermissions replaces a role's guild level permissions.
func (r *REST) SetRolePermissions(ctx context.Context, guildID, roleID string, permissions int64, reason string) error {
	return r.call(ctx, "rolePermissions", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.GuildRoleEdit(guildID, roleID, &discordgo.RoleParams{Permissions: &permissions}, with(o, reason)...)
		return e
	})
}
