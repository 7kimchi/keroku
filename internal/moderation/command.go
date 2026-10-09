package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/validate"
)

// fromRequest builds the common part of an action from a slash command.
func fromRequest(r *commands.Request, kind cases.Kind, reasonRequired bool) (Action, error) {
	target, ok, err := r.User("user")
	if err != nil {
		return Action{}, err
	}
	if !ok {
		return Action{}, commands.Fail(verb(kind)+" failed", "Pick a member.")
	}
	reason, ok, err := r.Text("reason", validate.MaxReason)
	if err != nil {
		return Action{}, err
	}
	if reasonRequired && (!ok || reason == "") {
		return Action{}, commands.Fail(verb(kind)+" failed", "A reason is required.")
	}
	gid, _ := validate.Snowflake(r.GuildID)
	tid, _ := validate.Snowflake(target)
	mid, _ := validate.Snowflake(r.UserID)
	return Action{
		Kind: kind, GuildID: gid, TargetID: tid, ModeratorID: mid, Reason: reason,
		InteractionID: r.InteractionID(), InvokerRoles: r.Member.Roles,
		InvokerPerms: r.Permissions, BotPerms: r.AppPermissions,
	}, nil
}

// run executes an action and renders the reply.
func (s *Service) run(ctx context.Context, a Action) (*discordgo.MessageEmbed, error) {
	res, err := s.Execute(ctx, a)
	if err != nil {
		return nil, err
	}
	return Reply(res, s.botID), nil
}

func userOption(desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: desc, Required: true}
}

func reasonOption(required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "reason",
		Description: "Reason, shown to the member and in the modlog", Required: required, MaxLength: validate.MaxReason}
}

func durationOption(desc string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "duration",
		Description: desc, Required: required, MaxLength: 32}
}

func definition(name, desc string, perm int64, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: name, Description: desc, Options: opts,
		DefaultMemberPermissions: commands.Perm(perm), Contexts: commands.GuildOnly()}
}
