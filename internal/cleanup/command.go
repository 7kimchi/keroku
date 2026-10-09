package cleanup

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

// MaxLockFor caps timed locks and lockdowns.
const MaxLockFor = 30 * 24 * time.Hour

func definition(name, desc string, perm int64, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: name, Description: desc, Options: opts,
		DefaultMemberPermissions: commands.Perm(perm), Contexts: commands.GuildOnly()}
}

func channelOption(desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel",
		Description: desc, ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews}}
}

func reasonOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "reason",
		Description: "Reason, shown in the audit log and modlog", MaxLength: validate.MaxReason}
}

func durationOption(desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "duration",
		Description: desc, MaxLength: 32}
}

// claim makes a replayed interaction a no-op. It reports false for a replay.
func (s *Service) claim(ctx context.Context, r *commands.Request) (bool, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	return cases.Claim(ctx, s.store.Pool(), r.InteractionID(), gid)
}

// target reads the channel option, defaulting to the channel the command ran in.
func target(r *commands.Request) (string, error) {
	ch, ok, err := r.Channel("channel")
	if err != nil {
		return "", err
	}
	if !ok {
		return r.ChannelID, nil
	}
	return ch, nil
}

func reason(r *commands.Request) (string, error) {
	text, _, err := r.Text("reason", validate.MaxReason)
	return text, err
}

// audit prefixes the moderator so the audit log shows who asked.
func audit(r *commands.Request, reason string) string {
	if reason == "" {
		reason = "No reason given."
	}
	return "Moderator " + r.UserID + ": " + reason
}

// event renders a modlog entry for a cleanup action.
func event(title string, r *commands.Request, fields ...[2]string) *discordgo.MessageEmbed {
	b := embeds.New(title).Field("Moderator", embeds.User(r.UserID), true)
	for _, f := range fields {
		b.Field(f[0], f[1], true)
	}
	return b.Timestamp(time.Now()).Build()
}

func replay() *discordgo.MessageEmbed {
	return embeds.New("No change").Description("This request was already handled.").Build()
}

func needBot(r *commands.Request, title string, bits int64, name string) error {
	if !perms.Has(r.AppPermissions, bits) {
		return commands.Fail(title, "Keroku is missing permission: "+name+".")
	}
	return nil
}
