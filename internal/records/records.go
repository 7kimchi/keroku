// Package records serves the commands that read and amend the case record.
package records

import (
	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

// Modlog receives embeds for reason changes.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Deps are shared by the record commands.
type Deps struct {
	Store  *store.Store
	Modlog Modlog
	BotID  string
}

func definition(name, desc string, opts ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: name, Description: desc, Options: opts,
		DefaultMemberPermissions: commands.Perm(perms.ModerateMembers), Contexts: commands.GuildOnly()}
}

func userOption(desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: desc, Required: true}
}

func numberOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: "case",
		Description: "Case number", Required: true, MinValue: ptr(1.0)}
}

func ptr(f float64) *float64 { return &f }
