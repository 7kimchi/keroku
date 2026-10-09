package settings

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func escalationGroup() *discordgo.ApplicationCommandOption {
	count := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: "warnings",
		Description: "Warning count that triggers the action", Required: true, MinValue: ptr(1), MaxValue: 50}
	action := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "action",
		Description: "What happens", Required: true, Choices: []*discordgo.ApplicationCommandOptionChoice{
			{Name: "Timeout", Value: "timeout"}, {Name: "Kick", Value: "kick"}, {Name: "Ban", Value: "ban"}}}
	duration := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "duration",
		Description: "Timeout or ban length like 1d. Required for timeouts", MaxLength: 32}
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommandGroup, Name: "escalation",
		Description: "Automatic actions when warnings add up", Options: []*discordgo.ApplicationCommandOption{
			sub("add", "Add or replace a step", count, action, duration),
			sub("remove", "Remove a step", count),
		}}
}

func ptr(f float64) *float64 { return &f }

func (c ConfigCommand) escalationAdd(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	count, _, err := r.Int("warnings", 1, 50)
	if err != nil {
		return nil, err
	}
	action, _, err := r.Text("action", 16)
	if err != nil {
		return nil, err
	}
	d, hasDuration, err := r.Duration("duration", time.Minute, moderation.MaxTimeout)
	if err != nil {
		return nil, err
	}
	switch {
	case action != "timeout" && action != "kick" && action != "ban":
		return nil, commands.Fail("Config failed", "Action must be timeout, kick or ban.")
	case action == "timeout" && !hasDuration:
		return nil, commands.Fail("Config failed", "Timeouts need a duration.")
	case action == "kick" && hasDuration:
		return nil, commands.Fail("Config failed", "Kicks take no duration.")
	}
	gid, _ := validate.Snowflake(r.GuildID)
	err = c.D.Store.SetEscalationStep(ctx, gid, store.EscalationStep{WarnCount: int(count), Action: action, Duration: d})
	if errors.Is(err, store.ErrTooManySteps) {
		return nil, commands.Fail("Config failed", "At most "+strconv.Itoa(store.MaxEscalationSteps)+" steps.")
	}
	if err != nil {
		return nil, err
	}
	return embeds.New("Escalation step set").Description(stepText(store.EscalationStep{WarnCount: int(count), Action: action, Duration: d})).Build(), nil
}

func (c ConfigCommand) escalationRemove(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	count, _, err := r.Int("warnings", 1, 50)
	if err != nil {
		return nil, err
	}
	gid, _ := validate.Snowflake(r.GuildID)
	ok, err := c.D.Store.RemoveEscalationStep(ctx, gid, int(count))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, commands.Fail("Config failed", "No step at "+strconv.FormatInt(count, 10)+" warnings.")
	}
	return embeds.New("Escalation step removed").Build(), nil
}

// stepText renders a step like "3 warnings: timeout for 1d".
func stepText(s store.EscalationStep) string {
	noun := " warnings: "
	if s.WarnCount == 1 {
		noun = " warning: "
	}
	out := strconv.Itoa(s.WarnCount) + noun + s.Action
	if s.Duration > 0 {
		out += " for " + embeds.Duration(s.Duration)
	}
	return out
}
