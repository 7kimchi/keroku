package moderation

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
)

// Modlog receives case embeds. Posting is queued and never blocks.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Service runs actions.
type Service struct {
	store   *store.Store
	client  discord.Client
	modlog  Modlog
	metrics *metrics.Metrics
	log     *slog.Logger
	guard   *safe.Guard
	botID   string
	now     func() time.Time
	window  time.Duration // double submit window
}

// Deps are the service's collaborators.
type Deps struct {
	Store   *store.Store
	Client  discord.Client
	Modlog  Modlog
	Metrics *metrics.Metrics
	Log     *slog.Logger
	Guard   *safe.Guard // nil gets one that logs to Log
	BotID   string
	Now     func() time.Time
}

// New builds a service.
func New(d Deps) *Service {
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Guard == nil {
		d.Guard = safe.NewGuard(d.Log, nil)
	}
	return &Service{store: d.Store, client: d.Client, modlog: d.Modlog, metrics: d.Metrics, log: d.Log, guard: d.Guard,
		botID: d.BotID, now: d.Now, window: duplicateWindow}
}
