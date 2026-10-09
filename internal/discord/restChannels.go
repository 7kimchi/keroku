package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Channel fetches a channel with its permission overwrites.
func (r *REST) Channel(ctx context.Context, channelID string) (c *discordgo.Channel, err error) {
	err = r.call(ctx, "channel", true, func(o ...discordgo.RequestOption) (e error) {
		c, e = r.s.Channel(channelID, o...)
		return
	})
	return c, err
}

// Send posts one embed with every mention disabled. It is never retried on server errors,
// so a flaky response cannot double post.
func (r *REST) Send(ctx context.Context, channelID string, embed *discordgo.MessageEmbed) error {
	if embed == nil {
		return &Error{Op: "send", Kind: BadRequest}
	}
	msg := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{embed}, AllowedMentions: NoMentions()}
	return r.call(ctx, "send", false, func(o ...discordgo.RequestOption) error {
		_, e := r.s.ChannelMessageSendComplex(channelID, msg, o...)
		return e
	})
}

// DM opens a DM channel and sends one embed.
func (r *REST) DM(ctx context.Context, userID string, embed *discordgo.MessageEmbed) error {
	if embed == nil {
		return &Error{Op: "dm", Kind: BadRequest}
	}
	var ch *discordgo.Channel
	err := r.call(ctx, "dmChannel", true, func(o ...discordgo.RequestOption) (e error) {
		ch, e = r.s.UserChannelCreate(userID, o...)
		return
	})
	if err != nil {
		return err
	}
	return r.Send(ctx, ch.ID, embed)
}

// Messages lists up to limit messages before beforeID, newest first.
func (r *REST) Messages(ctx context.Context, channelID string, limit int, beforeID string) (m []*discordgo.Message, err error) {
	err = r.call(ctx, "messages", true, func(o ...discordgo.RequestOption) (e error) {
		m, e = r.s.ChannelMessages(channelID, limit, beforeID, "", "", o...)
		return
	})
	return m, err
}

// BulkDelete removes 2 to 100 messages under 14 days old.
func (r *REST) BulkDelete(ctx context.Context, channelID string, ids []string, reason string) error {
	body := map[string][]string{"messages": ids}
	return r.call(ctx, "bulkDelete", false, func(o ...discordgo.RequestOption) error {
		ep := discordgo.EndpointChannelMessagesBulkDelete(channelID)
		_, e := r.s.RequestWithBucketID("POST", ep, body, ep, with(o, reason)...)
		return e
	})
}

// DeleteMessage removes one message.
func (r *REST) DeleteMessage(ctx context.Context, channelID, messageID, reason string) error {
	return r.call(ctx, "deleteMessage", true, func(o ...discordgo.RequestOption) error {
		return r.s.ChannelMessageDelete(channelID, messageID, with(o, reason)...)
	})
}

// SetSlowmode sets the per user message interval in seconds.
func (r *REST) SetSlowmode(ctx context.Context, channelID string, seconds int, reason string) error {
	return r.call(ctx, "slowmode", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.ChannelEdit(channelID, &discordgo.ChannelEdit{RateLimitPerUser: &seconds}, with(o, reason)...)
		return e
	})
}
