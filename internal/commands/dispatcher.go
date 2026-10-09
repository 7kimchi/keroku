package commands

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/workers"
)

// Tokens last 15 minutes. Anything older cannot be answered, so it is dropped.
const tokenLifetime = 14 * time.Minute

// Deps are the dispatcher's collaborators.
type Deps struct {
	Client         discord.Interactions
	Registry       *Registry
	Ack            *workers.Pool // fast lane: validate, rate limit, defer
	Lanes          *workers.Pool // keyed by laneKey: run handlers in order per target
	UserLimit      *ratelimit.Limiter
	GuildLimit     *ratelimit.Limiter
	Seen           *cache.Cache[string, struct{}]
	Metrics        *metrics.Metrics
	Log            *slog.Logger
	Guard          *safe.Guard
	Now            func() time.Time
	HandlerTimeout time.Duration
}

// Dispatcher routes interactions from the gateway to command handlers.
type Dispatcher struct{ d Deps }

// NewDispatcher wires a dispatcher.
func NewDispatcher(d Deps) *Dispatcher {
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Dispatcher{d: d}
}

// Interaction is called on the gateway reader. It never blocks: it dedupes and hands off.
func (x *Dispatcher) Interaction(i *discordgo.Interaction) {
	m := x.d.Metrics
	m.Events.WithLabelValues("INTERACTION_CREATE").Inc()
	if i == nil || i.Type != discordgo.InteractionApplicationCommand {
		m.EventsDropped.WithLabelValues("notCommand").Inc()
		return
	}
	if !x.d.Seen.Add(i.ID, struct{}{}) {
		m.EventsDropped.WithLabelValues("duplicate").Inc()
		return
	}
	if !x.d.Ack.Submit(i.ID, func(ctx context.Context) { x.ack(ctx, i) }) {
		m.EventsDropped.WithLabelValues("ackQueueFull").Inc()
	}
}

func (x *Dispatcher) ack(ctx context.Context, i *discordgo.Interaction) {
	ref := NewRef()
	if i.GuildID == "" {
		x.respond(ctx, i, immediate(errorEmbed("Command failed", "Server only.")))
		return
	}
	req, err := Parse(i, ref)
	if err != nil {
		x.d.Log.Warn("malformed interaction", "ref", ref, "interactionId", i.ID)
		x.respond(ctx, i, immediate(errorEmbed("Command failed", "Malformed request. Ref "+ref+".")))
		return
	}
	cmd, ok := x.d.Registry.Get(req.Name)
	if !ok {
		x.respond(ctx, i, immediate(errorEmbed("Command failed", "Unknown command.")))
		return
	}
	if missing, ok := missingPermission(cmd, req); ok {
		x.respond(ctx, i, immediate(errorEmbed("Command failed", "Missing permission: "+missing+".")))
		return
	}
	if wait, limited := x.limited(req); limited {
		retry := embeds.Relative(x.d.Now().Add(wait))
		x.respond(ctx, i, immediate(errorEmbed("Rate limited", "Try again "+retry+".")))
		return
	}
	if !x.respond(ctx, i, deferred()) {
		return
	}
	if !x.d.Lanes.Submit(laneKey(cmd, req), func(ctx context.Context) { x.run(ctx, cmd, req) }) {
		x.d.Metrics.EventsDropped.WithLabelValues("laneFull").Inc()
		x.edit(ctx, req, errorEmbed("Command failed", "Busy. Try again in a few seconds."))
	}
}
