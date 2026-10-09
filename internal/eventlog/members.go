package eventlog

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// MemberAdd logs a join with the account's age.
func (l *Logger) MemberAdd(m *discordgo.GuildMemberAdd) {
	userID, guildID := m.User.ID, m.GuildID
	l.submit(guildID, func(ctx context.Context) {
		gid, ok := l.logging(ctx, guildID)
		if !ok {
			return
		}
		uid, err := validate.Snowflake(userID)
		if err != nil {
			return
		}
		l.poster.Log(gid, embeds.New("Member joined").Field("Member", embeds.User(userID), true).
			Field("Account created", embeds.Relative(validate.SnowflakeTime(uid)), true).Timestamp(l.now()).Build())
	})
}

// MemberRemove logs a leave. Kicks and bans also arrive here; their cases are in the modlog.
func (l *Logger) MemberRemove(m *discordgo.GuildMemberRemove) {
	userID, guildID := m.User.ID, m.GuildID
	l.submit(guildID, func(ctx context.Context) {
		gid, ok := l.logging(ctx, guildID)
		if !ok {
			return
		}
		l.poster.Log(gid, embeds.New("Member left").Field("Member", embeds.User(userID), true).Timestamp(l.now()).Build())
	})
}
