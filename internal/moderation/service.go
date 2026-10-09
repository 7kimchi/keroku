package moderation

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
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
	botID   string
	now     func() time.Time
}

// Deps are the service's collaborators.
type Deps struct {
	Store   *store.Store
	Client  discord.Client
	Modlog  Modlog
	Metrics *metrics.Metrics
	Log     *slog.Logger
	BotID   string
	Now     func() time.Time
}

// New builds a service.
func New(d Deps) *Service {
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Service{store: d.Store, client: d.Client, modlog: d.Modlog, metrics: d.Metrics, log: d.Log, botID: d.BotID, now: d.Now}
}
