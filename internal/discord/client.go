// Package discord hides the Discord REST API behind small interfaces with a real and a fake
// implementation, so every handler can be tested without the network.
package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Members covers guild, member and ban calls.
type Members interface {
	Guild(ctx context.Context, guildID string) (*discordgo.Guild, error)
	Member(ctx context.Context, guildID, userID string) (*discordgo.Member, error)
	Ban(ctx context.Context, guildID, userID string, deleteSeconds int, reason string) error
	Unban(ctx context.Context, guildID, userID, reason string) error
	IsBanned(ctx context.Context, guildID, userID string) (bool, error)
	Kick(ctx context.Context, guildID, userID, reason string) error
	Timeout(ctx context.Context, guildID, userID string, until *time.Time, reason string) error
}

// Channels covers messages, channel settings and role permission edits.
type Channels interface {
	Channel(ctx context.Context, channelID string) (*discordgo.Channel, error)
	Send(ctx context.Context, channelID string, embed *discordgo.MessageEmbed) error
	DM(ctx context.Context, userID string, embed *discordgo.MessageEmbed) error
	Messages(ctx context.Context, channelID string, limit int, beforeID string) ([]*discordgo.Message, error)
	BulkDelete(ctx context.Context, channelID string, messageIDs []string, reason string) error
	DeleteMessage(ctx context.Context, channelID, messageID, reason string) error
	SetSlowmode(ctx context.Context, channelID string, seconds int, reason string) error
	SetRoleOverwrite(ctx context.Context, channelID, roleID string, allow, deny int64, reason string) error
	DeleteOverwrite(ctx context.Context, channelID, targetID, reason string) error
	SetRolePermissions(ctx context.Context, guildID, roleID string, permissions int64, reason string) error
}

// Interactions covers replying to slash commands.
type Interactions interface {
	Respond(ctx context.Context, i *discordgo.Interaction, resp *discordgo.InteractionResponse) error
	EditResponse(ctx context.Context, i *discordgo.Interaction, embed *discordgo.MessageEmbed) error
}

// AutoMod covers Discord's native AutoMod rules.
type AutoMod interface {
	AutoModRules(ctx context.Context, guildID string) ([]*discordgo.AutoModerationRule, error)
	CreateAutoModRule(ctx context.Context, guildID string, rule *discordgo.AutoModerationRule, reason string) error
	EditAutoModRule(ctx context.Context, guildID, ruleID string, rule *discordgo.AutoModerationRule, reason string) error
	DeleteAutoModRule(ctx context.Context, guildID, ruleID, reason string) error
}

// Commands covers slash command registration and the bot's own identity.
type Commands interface {
	OverwriteCommands(ctx context.Context, appID string, cmds []*discordgo.ApplicationCommand) error
	Identity(ctx context.Context) (appID, botUserID string, err error)
}

// Info covers the read only calls behind the info commands.
type Info interface {
	GuildCounts(ctx context.Context, guildID string) (*discordgo.Guild, error)
	ServerCount(ctx context.Context) (int, error)
}

// Client is everything the bot calls.
type Client interface {
	Members
	Channels
	Interactions
	AutoMod
	Commands
	Info
}

// NoMentions blocks every ping. It is attached to every message the bot sends.
func NoMentions() *discordgo.MessageAllowedMentions {
	return &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}
}
