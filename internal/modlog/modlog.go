// Package modlog queues posts to a guild's modlog and log channels so a flood of events
// never waits on Discord inside a command handler.
package modlog

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
	"github.com/7kimchi/keroku/internal/workers"
)

// Settings resolves a guild's channels.
type Settings interface {
	Get(ctx context.Context, guildID int64) (store.GuildSettings, error)
}

// Poster sends embeds to configured channels through a bounded queue.
type Poster struct {
	client   discord.Channels
	settings Settings
	pool     *workers.Pool
	metrics  *metrics.Metrics
	log      *slog.Logger
}

// New starts the queue with lanes workers and queue slots per lane.
func New(client discord.Channels, s Settings, lanes, queue int, m *metrics.Metrics, log *slog.Logger, g *safe.Guard) (*Poster, error) {
	pool, err := workers.New("modlog", lanes, queue, g)
	if err != nil {
		return nil, err
	}
	return &Poster{client: client, settings: s, pool: pool, metrics: m, log: log}, nil
}

// Case queues an embed for the modlog channel. It reports false if the queue is full.
func (p *Poster) Case(guildID int64, e *discordgo.MessageEmbed) bool {
	return p.enqueue(guildID, e, func(g store.GuildSettings) int64 { return g.ModlogChannelID })
}

// Log queues an embed for the message and member log channel.
func (p *Poster) Log(guildID int64, e *discordgo.MessageEmbed) bool {
	return p.enqueue(guildID, e, func(g store.GuildSettings) int64 { return g.LogChannelID })
}

func (p *Poster) enqueue(guildID int64, e *discordgo.MessageEmbed, pick func(store.GuildSettings) int64) bool {
	ok := p.pool.Submit(validate.FormatSnowflake(guildID), func(ctx context.Context) {
		p.send(ctx, guildID, e, pick)
	})
	if !ok {
		p.metrics.Modlog.WithLabelValues("dropped").Inc()
	}
	return ok
}

func (p *Poster) send(ctx context.Context, guildID int64, e *discordgo.MessageEmbed, pick func(store.GuildSettings) int64) {
	g, err := p.settings.Get(ctx, guildID)
	if err != nil {
		p.metrics.Modlog.WithLabelValues("settingsError").Inc()
		p.log.Warn("modlog settings lookup failed", "guildId", guildID, "err", err)
		return
	}
	channel := pick(g)
	if channel == 0 {
		p.metrics.Modlog.WithLabelValues("unset").Inc()
		return
	}
	if err := p.client.Send(ctx, validate.FormatSnowflake(channel), e); err != nil {
		// A deleted channel or lost permission is the guild's to fix. Log it and move on.
		p.metrics.Modlog.WithLabelValues(discord.KindOf(err).String()).Inc()
		p.log.Warn("modlog post failed", "guildId", guildID, "channelId", channel, "err", err)
		return
	}
	p.metrics.Modlog.WithLabelValues("ok").Inc()
}

// Depth reports queued posts.
func (p *Poster) Depth() int { return p.pool.Depth() }

// Close stops intake and drains queued posts until ctx ends.
func (p *Poster) Close(ctx context.Context) error { return p.pool.Close(ctx) }
