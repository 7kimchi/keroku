package app

import (
	"time"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/workers"
)

// Limits on per process state. Each bounds memory no matter how many users show up.
const (
	maxLimiterKeys = 200_000
	maxSeen        = 200_000
)

// build creates the queues, limiters and dispatcher.
func (a *App) build() error {
	var err error
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
	return nil
}
