package automod

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// noticeEvery limits modlog notices to one per member per minute during a flood.
const noticeEvery = time.Minute

// act deletes the message and, when configured, times the member out through the normal
// moderation path, which records a case and posts it.
func (e *Engine) act(ctx context.Context, gid int64, m *discordgo.Message, cfg store.AutomodSettings, rule string) {
	reason := "Automod: " + rule + "."
	if err := e.d.Client.DeleteMessage(ctx, m.ChannelID, m.ID, reason); err != nil {
		e.d.Log.Warn("automod delete failed", "guildId", gid, "rule", rule, "err", err)
	}
	if cfg.Timeout > 0 {
		uid, err := validate.Snowflake(m.Author.ID)
		if err != nil {
			return
		}
		_, err = e.d.Moderation.Execute(ctx, moderation.Action{
			Kind: cases.Timeout, GuildID: gid, TargetID: uid, ModeratorID: e.d.BotID, Reason: reason,
			Duration: cfg.Timeout, IdempotencyKey: "automod:" + m.ID, Automated: true,
		})
		if err == nil {
			return
		}
		e.d.Log.Warn("automod timeout failed", "guildId", gid, "rule", rule, "err", err)
	}
	if e.shouldNotice(m) {
		e.d.Modlog.Case(gid, embeds.New("Message removed").Field("Member", embeds.User(m.Author.ID), true).
			Field("Rule", rule, true).Field("Channel", embeds.Channel(m.ChannelID), true).Timestamp(e.d.Now()).Build())
	}
}

func (e *Engine) shouldNotice(m *discordgo.Message) bool {
	now := e.d.Now()
	notice := false
	e.state.Update(m.GuildID+":"+m.Author.ID, func(st *userState, found bool) *userState {
		if !found || st == nil {
			st = &userState{}
		}
		if now.Sub(st.noticeAt) >= noticeEvery {
			st.noticeAt, notice = now, true
		}
		return st
	})
	return notice
}
