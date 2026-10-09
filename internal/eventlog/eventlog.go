// Package eventlog posts message edits and deletes and member joins and leaves to the
// guild's log channel. Recent message text is kept in a bounded cache so a delete can show
// what was removed.
package eventlog

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
)

// Memory bounds for the message cache.
const (
	maxMessages = 50_000
	maxText     = 500
	messageTTL  = time.Hour
)

// Poster queues embeds for the log channel.
type Poster interface {
	Log(guildID int64, e *discordgo.MessageEmbed) bool
}

// Settings resolves whether a guild has a log channel.
type Settings interface {
	Get(ctx context.Context, guildID int64) (store.GuildSettings, error)
}

// cached is what is remembered about one message.
type cached struct {
	authorID  string
	channelID string
	text      string
}

// Logger turns gateway events into log channel posts.
type Logger struct {
	poster   Poster
	settings Settings
	pool     *workers.Pool
	metrics  *metrics.Metrics
	messages *cache.Cache[string, cached]
	now      func() time.Time
}

// New builds a logger. now may be nil.
func New(p Poster, s Settings, pool *workers.Pool, m *metrics.Metrics, now func() time.Time) (*Logger, error) {
	if now == nil {
		now = time.Now
	}
	c, err := cache.New[string, cached](maxMessages, messageTTL, now)
	if err != nil {
		return nil, err
	}
	return &Logger{poster: p, settings: s, pool: pool, metrics: m, messages: c, now: now}, nil
}

// submit hands work to the guild's lane without blocking the gateway reader.
func (l *Logger) submit(guildID string, fn func(ctx context.Context)) {
	if !l.pool.Submit(guildID, fn) {
		l.metrics.EventsDropped.WithLabelValues("logQueueFull").Inc()
	}
}

// Sweep drops expired cached messages.
func (l *Logger) Sweep() { l.messages.Sweep(2000) }
