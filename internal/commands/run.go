package commands

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

const replyTimeout = 10 * time.Second

// run executes a handler on the guild's lane and edits the deferred reply with the result.
func (x *Dispatcher) run(ctx context.Context, cmd Command, req *Request) {
	if x.d.Now().Sub(req.At) > tokenLifetime {
		x.d.Metrics.EventsDropped.WithLabelValues("expired").Inc()
		return
	}
	start := x.d.Now()
	hctx, cancel := context.WithTimeout(ctx, x.d.HandlerTimeout)
	var reply *discordgo.MessageEmbed
	var err error
	panicked := x.d.Guard.Run("command "+req.Name, func() { reply, err = cmd.Handle(hctx, req) })
	cancel()
	result := "ok"
	var userErr *UserError
	switch {
	case panicked:
		result, reply = "panic", internalError(req.Ref)
	case errors.As(err, &userErr):
		result, reply = "refused", errorEmbed(userErr.Title, userErr.Detail)
	case err != nil:
		result, reply = "error", internalError(req.Ref)
		x.d.Log.Error("command failed", "ref", req.Ref, "command", req.Name, "guildId", req.GuildID, "err", err)
	case reply == nil:
		result, reply = "error", internalError(req.Ref)
		x.d.Log.Error("command returned no reply", "ref", req.Ref, "command", req.Name)
	}
	x.d.Metrics.Commands.WithLabelValues(req.Name, result).Inc()
	x.d.Metrics.CommandSeconds.WithLabelValues(req.Name).Observe(x.d.Now().Sub(start).Seconds())
	x.edit(ctx, req, reply)
}

// respond sends the initial response and reports whether it landed.
func (x *Dispatcher) respond(ctx context.Context, i *discordgo.Interaction, resp *discordgo.InteractionResponse) bool {
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), replyTimeout)
	defer cancel()
	if err := x.d.Client.Respond(rctx, i, resp); err != nil {
		x.d.Metrics.DiscordErrors.WithLabelValues(discord.KindOf(err).String()).Inc()
		x.d.Log.Warn("interaction response failed", "interactionId", i.ID, "err", err)
		return false
	}
	return true
}

// edit replaces the deferred reply. It uses its own deadline so a timed out handler still answers.
func (x *Dispatcher) edit(ctx context.Context, req *Request, e *discordgo.MessageEmbed) {
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), replyTimeout)
	defer cancel()
	if err := x.d.Client.EditResponse(rctx, req.Interaction, e); err != nil {
		x.d.Metrics.DiscordErrors.WithLabelValues(discord.KindOf(err).String()).Inc()
		x.d.Log.Warn("interaction edit failed", "ref", req.Ref, "err", err)
	}
}
