// Package commands parses slash command interactions, routes them to handlers and replies.
package commands

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/validate"
)

type options = map[string]*discordgo.ApplicationCommandInteractionDataOption

// Request is a validated slash command invocation.
type Request struct {
	Interaction    *discordgo.Interaction
	GuildID        string
	ChannelID      string
	UserID         string
	Member         *discordgo.Member
	Permissions    int64 // invoker permissions in the channel, computed by Discord
	AppPermissions int64 // bot permissions in the channel, computed by Discord
	Name           string
	Sub            string // subcommand, empty if none
	Ref            string // correlation id shown to users on failure
	At             time.Time
	opts           options
	resolved       *discordgo.ApplicationCommandInteractionDataResolved
}

var errMalformed = errors.New("commands: malformed interaction")

// Parse validates the interaction's shape and every id in it.
func Parse(i *discordgo.Interaction, ref string) (*Request, error) {
	if i == nil || i.Type != discordgo.InteractionApplicationCommand || i.Member == nil || i.Member.User == nil {
		return nil, errMalformed
	}
	data, ok := i.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok || data.CommandType != discordgo.ChatApplicationCommand {
		return nil, errMalformed
	}
	interactionID, err := validate.Snowflake(i.ID)
	if err != nil {
		return nil, errMalformed
	}
	for _, id := range []string{i.GuildID, i.ChannelID, i.Member.User.ID} {
		if _, err := validate.Snowflake(id); err != nil {
			return nil, errMalformed
		}
	}
	r := &Request{
		Interaction: i, GuildID: i.GuildID, ChannelID: i.ChannelID, UserID: i.Member.User.ID,
		Member: i.Member, Permissions: i.Member.Permissions, AppPermissions: i.AppPermissions,
		Name: data.Name, Ref: ref, At: validate.SnowflakeTime(interactionID), resolved: data.Resolved,
	}
	r.Sub, r.opts, err = flatten(data.Options)
	return r, err
}

// flatten unwraps at most one subcommand group and one subcommand.
func flatten(in []*discordgo.ApplicationCommandInteractionDataOption) (string, options, error) {
	sub := ""
	for depth := 0; len(in) == 1 && in[0] != nil && isSub(in[0].Type); depth++ {
		if depth == 2 {
			return "", nil, errMalformed
		}
		if sub != "" {
			sub += " "
		}
		sub += in[0].Name
		in = in[0].Options
	}
	if len(in) > 25 {
		return "", nil, errMalformed
	}
	out := make(options, len(in))
	for _, o := range in {
		if o == nil || isSub(o.Type) || out[o.Name] != nil {
			return "", nil, errMalformed
		}
		out[o.Name] = o
	}
	return sub, out, nil
}

func isSub(t discordgo.ApplicationCommandOptionType) bool {
	return t == discordgo.ApplicationCommandOptionSubCommand || t == discordgo.ApplicationCommandOptionSubCommandGroup
}

// InteractionID returns the interaction id as a number. Parse already validated it.
func (r *Request) InteractionID() int64 {
	id, _ := validate.Snowflake(r.Interaction.ID)
	return id
}
