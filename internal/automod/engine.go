// Package automod deletes spam, repeated messages and unapproved links, and keeps Discord's
// native AutoMod rules for mention spam and invite links in sync with the guild's settings.
package automod

import (
	"context"
	"hash/maphash"
	"log/slog"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
)

// Bounds on memory. A million hostile accounts still fit in maxUsers entries.
const (
	maxUsers  = 200_000
	maxGuilds = 50_000
	stateTTL  = 10 * time.Minute
	rolesTTL  = 5 * time.Minute
)

// Modlog receives automod notices.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Deps are the engine's collaborators.
type Deps struct {
	Settings   *store.Cached[store.AutomodSettings]
	Client     discord.Client
	Moderation *moderation.Service
	Modlog     Modlog
	Pool       *workers.Pool
	Metrics    *metrics.Metrics
	Log        *slog.Logger
	BotID      int64
	Now        func() time.Time
}

// Engine checks messages against the custom rules.
type Engine struct {
	d       Deps
	state   *cache.Cache[string, *userState]
	guilds  *cache.Cache[string, *discordgo.Guild]
	seed    maphash.Seed
	pattern *regexp.Regexp
}

// New builds an engine.
func New(d Deps) (*Engine, error) {
	if d.Now == nil {
		d.Now = time.Now
	}
	st, err := cache.New[string, *userState](maxUsers, stateTTL, d.Now)
	if err != nil {
		return nil, err
	}
	gs, err := cache.New[string, *discordgo.Guild](maxGuilds, rolesTTL, d.Now)
	if err != nil {
		return nil, err
	}
	return &Engine{d: d, state: st, guilds: gs, seed: maphash.MakeSeed(), pattern: linkPattern()}, nil
}

// Message hands a new message to the guild's lane. It never blocks the gateway reader.
func (e *Engine) Message(m *discordgo.Message) {
	if !e.d.Pool.Submit(m.GuildID, func(ctx context.Context) { e.handle(ctx, m) }) {
		e.d.Metrics.EventsDropped.WithLabelValues("automodQueueFull").Inc()
	}
}

// Sweep drops expired per user state.
func (e *Engine) Sweep() {
	e.state.Sweep(2000)
	e.guilds.Sweep(1000)
}
