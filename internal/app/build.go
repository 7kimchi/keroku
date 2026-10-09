package app

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/modlog"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
	"github.com/bwmarrin/discordgo"
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
	if a.events, err = workers.New("events", a.cfg.Workers, a.cfg.QueueSize, a.guard); err != nil {
		return err
	}
	if err = a.services(); err != nil {
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
	a.router = &router{interactions: d,
		messages: []func(*discordgo.Message){a.automod.Message, a.eventlog.MessageCreate},
		updates:  []func(*discordgo.MessageUpdate){a.eventlog.MessageUpdate},
		deletes:  []func(*discordgo.MessageDelete){a.eventlog.MessageDelete},
		joins:    []func(*discordgo.GuildMemberAdd){a.raid.MemberAdd, a.eventlog.MemberAdd},
		leaves:   []func(*discordgo.GuildMemberRemove){a.eventlog.MemberRemove}}
	return nil
}
