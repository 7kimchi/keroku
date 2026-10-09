// Package perms computes Discord permissions and runs the moderation hierarchy checks.
package perms

import "github.com/bwmarrin/discordgo"

// Permission bits the bot cares about, re-exported so callers do not need discordgo.
const (
	Administrator   int64 = discordgo.PermissionAdministrator
	ViewChannel     int64 = discordgo.PermissionViewChannel
	SendMessages    int64 = discordgo.PermissionSendMessages
	EmbedLinks      int64 = discordgo.PermissionEmbedLinks
	ReadHistory     int64 = discordgo.PermissionReadMessageHistory
	ManageMessages  int64 = discordgo.PermissionManageMessages
	ManageChannels  int64 = discordgo.PermissionManageChannels
	ManageRoles     int64 = discordgo.PermissionManageRoles
	ManageGuild     int64 = discordgo.PermissionManageGuild
	KickMembers     int64 = discordgo.PermissionKickMembers
	BanMembers      int64 = discordgo.PermissionBanMembers
	ModerateMembers int64 = discordgo.PermissionModerateMembers
	SendInThreads   int64 = discordgo.PermissionSendMessagesInThreads
	PublicThreads   int64 = discordgo.PermissionCreatePublicThreads
	PrivateThreads  int64 = discordgo.PermissionCreatePrivateThreads
	AddReactions    int64 = discordgo.PermissionAddReactions
)

// LockBits are the permissions removed from @everyone by a channel lock or lockdown.
const LockBits = SendMessages | SendInThreads | PublicThreads | PrivateThreads | AddReactions

// Has reports whether granted covers every bit in need. Administrator covers everything.
func Has(granted, need int64) bool {
	if granted&Administrator != 0 {
		return true
	}
	return granted&need == need
}

// Name returns the display name of a single permission bit.
func Name(bit int64) string {
	switch bit {
	case BanMembers:
		return "Ban Members"
	case KickMembers:
		return "Kick Members"
	case ModerateMembers:
		return "Timeout Members"
	case ManageMessages:
		return "Manage Messages"
	case ManageChannels:
		return "Manage Channels"
	case ManageRoles:
		return "Manage Roles"
	case ManageGuild:
		return "Manage Server"
	case ReadHistory:
		return "Read Message History"
	case SendMessages:
		return "Send Messages"
	case EmbedLinks:
		return "Embed Links"
	case ViewChannel:
		return "View Channel"
	case Administrator:
		return "Administrator"
	}
	return "Unknown"
}
