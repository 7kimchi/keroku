package app

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/modlog"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/records"
	"github.com/7kimchi/keroku/internal/settings"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/sweeper"
	"github.com/7kimchi/keroku/internal/validate"
	"github.com/7kimchi/keroku/internal/workers"
)

// Limits on per process state. Each bounds memory no matter how many users show up.
const (
	maxLimiterKeys  = 200_000
	maxSeen         = 200_000
	maxCachedGuilds = 50_000
)

// build resolves the bot's identity and creates the queues, limiters, services and dispatcher.
func (a *App) build(ctx context.Context) error {
	var err error
	if a.appID, a.botID, err = a.client.Identity(ctx); err != nil {
		return err
	}
	if a.settings, err = store.NewSettingsCache(a.store, maxCachedGuilds, time.Minute); err != nil {
		return err
	}
	if a.modlog, err = modlog.New(a.client, a.settings, 8, 1024, a.metrics, a.log, a.guard); err != nil {
		return err
	}
	if a.ack, err = workers.New("ack", max(2, a.cfg.Workers/4), a.cfg.QueueSize, a.guard); err != nil {
		return err
	}
	if a.lanes, err = workers.New("lanes", a.cfg.Workers, a.cfg.QueueSize, a.guard); err != nil {
		return err
	}
	if a.userLimit, err = ratelimit.New(2*time.Second, 5, maxLimiterKeys, nil); err != nil {
		return err
	}
	if a.guildLimit, err = ratelimit.New(200*time.Millisecond, 30, maxLimiterKeys, nil); err != nil {
		return err
	}
	if a.seen, err = cache.New[string, struct{}](maxSeen, 15*time.Minute, nil); err != nil {
		return err
	}
	if a.registry, err = commands.NewRegistry(a.commandList()...); err != nil {
		return err
	}
	d := commands.NewDispatcher(commands.Deps{
		Client: a.client, Registry: a.registry, Ack: a.ack, Lanes: a.lanes,
		UserLimit: a.userLimit, GuildLimit: a.guildLimit, Seen: a.seen,
		Metrics: a.metrics, Log: a.log, Guard: a.guard, HandlerTimeout: 60 * time.Second,
	})
	a.router = &router{interactions: d}
	return nil
}

// commandList returns every slash command.
func (a *App) commandList() []commands.Command {
	mod := moderation.New(moderation.Deps{Store: a.store, Client: a.client, Modlog: a.modlog,
		Metrics: a.metrics, Log: a.log, BotID: a.botID})
	rec := records.Deps{Store: a.store, Modlog: a.modlog, BotID: a.botID}
	clean := cleanup.New(a.store, a.client, a.modlog, nil)
	botID, _ := validate.Snowflake(a.botID)
	a.sweeper = sweeper.New(sweeper.Deps{Store: a.store, Client: a.client, Moderation: mod, Cleanup: clean,
		Modlog: a.modlog, Metrics: a.metrics, Log: a.log, BotID: botID})
	cfg := settings.Deps{Store: a.store, Settings: a.settings, Client: a.client, BotID: a.botID}
	return []commands.Command{
		moderation.BanCommand{S: mod}, moderation.UnbanCommand{S: mod}, moderation.KickCommand{S: mod},
		moderation.TimeoutCommand{S: mod}, moderation.UntimeoutCommand{S: mod}, moderation.WarnCommand{S: mod},
		moderation.NoteCommand{S: mod},
		records.CaseCommand{D: rec}, records.HistoryCommand{D: rec}, records.WarningsCommand{D: rec},
		cleanup.PurgeCommand{S: clean}, cleanup.SlowmodeCommand{S: clean}, cleanup.LockCommand{S: clean},
		cleanup.UnlockCommand{S: clean}, cleanup.LockdownCommand{S: clean},
		settings.ConfigCommand{D: cfg},
	}
}
