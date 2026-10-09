// Package app wires every component together and owns startup and shutdown order.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/automod"
	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/config"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/eventlog"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/modlog"
	"github.com/7kimchi/keroku/internal/raid"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/sweeper"
	"github.com/7kimchi/keroku/internal/workers"
	"github.com/7kimchi/keroku/migrations"
)

// App holds the running components.
type App struct {
	cfg        config.Config
	log        *slog.Logger
	metrics    *metrics.Metrics
	guard      *safe.Guard
	store      *store.Store
	client     discord.Client
	ack        *workers.Pool
	lanes      *workers.Pool
	userLimit  *ratelimit.Limiter
	guildLimit *ratelimit.Limiter
	seen       *cache.Cache[string, struct{}]
	registry   *commands.Registry
	router     *router
	settings   *store.SettingsCache
	modlog     *modlog.Poster
	sweeper    *sweeper.Sweeper
	events     *workers.Pool // automod, raid and logs, keyed by guild
	mod        *moderation.Service
	clean      *cleanup.Service
	automod    *automod.Engine
	raid       *raid.Detector
	eventlog   *eventlog.Logger

	automodSettings *store.Cached[store.AutomodSettings]
	raidSettings    *store.Cached[store.RaidSettings]
	appID           string
	botID           string
	started         time.Time
	gatewayHTTP     *http.Client // tests point the gateway at a fake
}

// New connects to Postgres, migrates and builds every component. Nothing touches the gateway yet.
func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	a := &App{cfg: cfg, log: log, metrics: metrics.New(), started: time.Now()}
	a.guard = safe.NewGuard(log, func(where string) { a.metrics.Panics.WithLabelValues(where).Inc() })
	pool, err := store.Open(ctx, cfg.DatabaseURL.Reveal(), cfg.DBMaxConns)
	if err != nil {
		return nil, err
	}
	a.store = store.New(pool)
	if err := store.Migrate(ctx, pool, migrations.Files()); err != nil {
		a.store.Close()
		return nil, err
	}
	session, err := discordgo.New("Bot " + cfg.Token.Reveal())
	if err != nil {
		a.store.Close()
		return nil, err
	}
	session.StateEnabled = false
	a.client = discord.NewREST(session, 10*time.Second)
	if err := a.build(ctx); err != nil {
		a.store.Close()
		return nil, err
	}
	return a, nil
}

// assemble builds an app around an existing store and client, for tests.
func assemble(cfg config.Config, log *slog.Logger, st *store.Store, client discord.Client) (*App, error) {
	a := &App{cfg: cfg, log: log, metrics: metrics.New(), store: st, client: client, started: time.Now()}
	a.guard = safe.NewGuard(log, func(where string) { a.metrics.Panics.WithLabelValues(where).Inc() })
	return a, a.build(context.Background())
}
