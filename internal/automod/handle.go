package automod

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func (e *Engine) handle(ctx context.Context, m *discordgo.Message) {
	gid, err := validate.Snowflake(m.GuildID)
	if err != nil || m.Author == nil || m.Author.Bot || m.Member == nil {
		return
	}
	cfg, err := e.d.Settings.Get(ctx, gid)
	if err != nil {
		e.d.Log.Warn("automod settings lookup failed", "guildId", gid, "err", err)
		return
	}
	if !cfg.SpamEnabled && !cfg.DuplicateEnabled && !cfg.LinksEnabled {
		return
	}
	if exempt, err := e.exempt(ctx, m); err != nil || exempt {
		return
	}
	rule := e.evaluate(m, cfg)
	if rule == "" {
		return
	}
	e.d.Metrics.Automod.WithLabelValues(rule).Inc()
	e.act(ctx, gid, m, cfg, rule)
}

// evaluate updates the member's state and returns the first rule that fired, or "".
func (e *Engine) evaluate(m *discordgo.Message, cfg store.AutomodSettings) string {
	now := e.d.Now()
	rule := ""
	e.state.Update(m.GuildID+":"+m.Author.ID, func(st *userState, found bool) *userState {
		if !found || st == nil {
			st = &userState{}
		}
		// Every rule sees every message so its window stays accurate.
		hits := []struct {
			name string
			hit  bool
		}{
			{"spam", spam(st, cfg, now)},
			{"duplicate", duplicate(st, cfg, m.Content, e.seed, now)},
			{"links", links(cfg, m.Content, e.pattern)},
		}
		for _, h := range hits {
			if h.hit && rule == "" {
				rule = h.name
			}
		}
		return st
	})
	return rule
}

// exempt skips staff: members who can manage messages, administrators and the owner.
func (e *Engine) exempt(ctx context.Context, m *discordgo.Message) (bool, error) {
	g, ok := e.guilds.Get(m.GuildID)
	if !ok {
		var err error
		if g, err = e.d.Client.Guild(ctx, m.GuildID); err != nil {
			return false, err
		}
		e.guilds.Set(m.GuildID, g)
	}
	p := perms.Base(g, m.Author.ID, m.Member.Roles)
	return perms.Has(p, perms.ManageMessages), nil
}
