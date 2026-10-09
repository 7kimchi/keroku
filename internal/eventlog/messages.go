package eventlog

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// MessageCreate remembers the text of messages in guilds that log.
func (l *Logger) MessageCreate(m *discordgo.Message) {
	l.submit(m.GuildID, func(ctx context.Context) {
		if _, ok := l.logging(ctx, m.GuildID); ok {
			l.remember(m)
		}
	})
}

// MessageUpdate logs real edits. Discord also sends updates when link previews load;
// those have no edit time and are skipped.
func (l *Logger) MessageUpdate(m *discordgo.MessageUpdate) {
	msg := m.Message
	l.submit(msg.GuildID, func(ctx context.Context) {
		gid, ok := l.logging(ctx, msg.GuildID)
		if !ok || msg.EditedTimestamp == nil || (msg.Author != nil && msg.Author.Bot) {
			return
		}
		before, known := l.messages.Get(msg.ID)
		after := text(msg.Content)
		if known && before.text == after {
			return
		}
		author := ""
		if msg.Author != nil {
			author = msg.Author.ID
		} else if known {
			author = before.authorID
		}
		l.remember(msg)
		l.poster.Log(gid, embeds.New("Message edited").
			Field("Author", userOrUnknown(author), true).Field("Channel", embeds.Channel(msg.ChannelID), true).
			Field("Before", beforeText(before, known), false).Field("After", show(after), false).
			Field("Message", jump(msg.GuildID, msg.ChannelID, msg.ID), false).Timestamp(l.now()).Build())
	})
}

// MessageDelete logs a deleted message with whatever was remembered about it.
func (l *Logger) MessageDelete(m *discordgo.MessageDelete) {
	msg := m.Message
	l.submit(msg.GuildID, func(ctx context.Context) {
		gid, ok := l.logging(ctx, msg.GuildID)
		if !ok {
			return
		}
		before, known := l.messages.Get(msg.ID)
		l.messages.Delete(msg.ID)
		l.poster.Log(gid, embeds.New("Message deleted").
			Field("Author", userOrUnknown(before.authorID), true).Field("Channel", embeds.Channel(msg.ChannelID), true).
			Field("Content", beforeText(before, known), false).Timestamp(l.now()).Build())
	})
}

func (l *Logger) remember(m *discordgo.Message) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	l.messages.Set(m.ID, cached{authorID: m.Author.ID, channelID: m.ChannelID, text: text(m.Content)})
}

// logging reports whether the guild has a log channel, so nothing is cached or posted otherwise.
func (l *Logger) logging(ctx context.Context, guildID string) (int64, bool) {
	gid, err := validate.Snowflake(guildID)
	if err != nil {
		return 0, false
	}
	s, err := l.settings.Get(ctx, gid)
	return gid, err == nil && s.LogChannelID != 0
}
