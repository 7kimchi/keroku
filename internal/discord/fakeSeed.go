package discord

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func itoa(n int) string { return strconv.Itoa(n) }

// AddGuild registers a guild. Its @everyone role gets the guild's id.
func (f *Fake) AddGuild(id, ownerID string, everyonePerms int64, roles ...*discordgo.Role) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := append([]*discordgo.Role{{ID: id, Name: "@everyone", Permissions: everyonePerms}}, roles...)
	f.guilds[id] = &discordgo.Guild{ID: id, OwnerID: ownerID, Roles: all}
	f.members[id] = map[string]*discordgo.Member{}
	f.bans[id] = map[string]bool{}
}

// AddMember puts a user in a guild with the given roles.
func (f *Fake) AddMember(guildID, userID string, roles ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members[guildID][userID] = &discordgo.Member{
		GuildID: guildID, User: &discordgo.User{ID: userID, Username: "user" + userID}, Roles: roles,
	}
}

// RemoveMember takes a user out of a guild, as if they left.
func (f *Fake) RemoveMember(guildID, userID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.members[guildID], userID)
}

// SetMemberRoles replaces a member's roles, as if a moderator edited them.
func (f *Fake) SetMemberRoles(guildID, userID string, roles ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m := f.members[guildID][userID]; m != nil {
		m.Roles = roles
	}
}

// AddChannel registers a text channel.
func (f *Fake) AddChannel(guildID, channelID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.channels[channelID] = &discordgo.Channel{ID: channelID, GuildID: guildID, Type: discordgo.ChannelTypeGuildText}
}

// DeleteChannel removes a channel, as if an admin deleted it.
func (f *Fake) DeleteChannel(channelID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.channels, channelID)
}

// AddMessage stores a message in a channel. Messages must be added oldest first.
func (f *Fake) AddMessage(channelID, messageID, authorID string, at time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages[channelID] = append(f.messages[channelID], &discordgo.Message{
		ID: messageID, ChannelID: channelID, Author: &discordgo.User{ID: authorID}, Timestamp: at,
	})
}

// BlockDMs makes DMs to the user fail like a user with DMs closed.
func (f *Fake) BlockDMs(userID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blocked[userID] = true
}

// ExpireInteraction makes every reply to the interaction fail as expired.
func (f *Fake) ExpireInteraction(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expired[id] = true
}
