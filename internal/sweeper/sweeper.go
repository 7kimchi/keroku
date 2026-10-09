// Package sweeper ends temporary actions when they expire. It survives restarts because
// timers live in Postgres, and several instances can run it at once: each timer is
// claimed with FOR UPDATE SKIP LOCKED in its own transaction.
package sweeper

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
)

// Modlog receives notices about timers that could not be completed.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Deps are the sweeper's collaborators.
type Deps struct {
	Store      *store.Store
	Client     discord.Client
	Moderation *moderation.Service
	Cleanup    *cleanup.Service
	Modlog     Modlog
	Metrics    *metrics.Metrics
	Log        *slog.Logger
	BotID      int64
	Now        func() time.Time
	Interval   time.Duration // pause between sweeps when idle
	JobTimeout time.Duration // bound on one timer's work
}

// Sweeper processes due timers.
type Sweeper struct{ d Deps }

// New builds a sweeper with defaults for unset durations.
func New(d Deps) *Sweeper {
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Interval <= 0 {
		d.Interval = 5 * time.Second
	}
	if d.JobTimeout <= 0 {
		d.JobTimeout = 30 * time.Second
	}
	return &Sweeper{d: d}
}

// Run sweeps at once, which covers timers that expired while the bot was down, then
// keeps sweeping until ctx ends.
func (s *Sweeper) Run(ctx context.Context) {
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		for ctx.Err() == nil {
			worked, err := s.Once(ctx)
			if err != nil {
				s.d.Log.Warn("sweep failed", "err", err)
			}
			if !worked || err != nil {
				break
			}
		}
		t.Reset(s.d.Interval)
	}
}
