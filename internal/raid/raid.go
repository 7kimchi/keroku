// Package raid watches joins. A burst of joins inside the window is a raid: Keroku alerts
// the modlog and kicks or bans the joiners, or locks the server down. Accounts younger than
// the minimum age are removed as they join.
package raid

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
)

// maxGuilds bounds the per guild join windows held in memory.
const maxGuilds = 100_000

// Modlog receives raid alerts.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Deps are the detector's collaborators.
type Deps struct {
	Settings   *store.Cached[store.RaidSettings]
	Moderation *moderation.Service
	Cleanup    *cleanup.Service
	Modlog     Modlog
	Pool       *workers.Pool
	Metrics    *metrics.Metrics
	Log        *slog.Logger
	BotID      int64
	Now        func() time.Time
}

// Detector tracks joins per guild.
type Detector struct {
	d       Deps
	windows *cache.Cache[string, *window]
}

// New builds a detector.
func New(d Deps) (*Detector, error) {
	if d.Now == nil {
		d.Now = time.Now
	}
	w, err := cache.New[string, *window](maxGuilds, time.Hour, d.Now)
	if err != nil {
		return nil, err
	}
	return &Detector{d: d, windows: w}, nil
}

// MemberAdd hands a join to the guild's lane. It never blocks the gateway reader.
func (x *Detector) MemberAdd(m *discordgo.GuildMemberAdd) {
	joined := m.JoinedAt
	userID := m.User.ID
	if !x.d.Pool.Submit(m.GuildID, func(ctx context.Context) { x.handle(ctx, m.GuildID, userID, joined) }) {
		x.d.Metrics.EventsDropped.WithLabelValues("raidQueueFull").Inc()
	}
}

// Sweep drops idle windows.
func (x *Detector) Sweep() { x.windows.Sweep(1000) }
