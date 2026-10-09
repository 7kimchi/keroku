package perms

import "github.com/bwmarrin/discordgo"

// Base returns guild level permissions for a member: @everyone plus every role they hold.
func Base(g *discordgo.Guild, userID string, memberRoles []string) int64 {
	if g.OwnerID == userID {
		return allBits
	}
	held := make(map[string]bool, len(memberRoles))
	for _, id := range memberRoles {
		held[id] = true
	}
	var p int64
	for _, r := range g.Roles {
		if r.ID == g.ID || held[r.ID] {
			p |= r.Permissions
		}
	}
	if p&Administrator != 0 {
		return allBits
	}
	return p
}

// InChannel applies channel overwrites in Discord's order: @everyone, roles, then the member.
func InChannel(g *discordgo.Guild, ch *discordgo.Channel, userID string, memberRoles []string) int64 {
	p := Base(g, userID, memberRoles)
	if p&Administrator != 0 {
		return allBits
	}
	held := make(map[string]bool, len(memberRoles))
	for _, id := range memberRoles {
		held[id] = true
	}
	var roleAllow, roleDeny int64
	var member *discordgo.PermissionOverwrite
	for _, o := range ch.PermissionOverwrites {
		switch {
		case o.Type == discordgo.PermissionOverwriteTypeRole && o.ID == g.ID:
			p = (p &^ o.Deny) | o.Allow
		case o.Type == discordgo.PermissionOverwriteTypeRole && held[o.ID]:
			roleAllow |= o.Allow
			roleDeny |= o.Deny
		case o.Type == discordgo.PermissionOverwriteTypeMember && o.ID == userID:
			member = o
		}
	}
	p = (p &^ roleDeny) | roleAllow
	if member != nil {
		p = (p &^ member.Deny) | member.Allow
	}
	return p
}

// allBits grants every permission, used for owners and administrators.
const allBits int64 = 1<<53 - 1
