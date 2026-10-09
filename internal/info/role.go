package info

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/perms"
)

// RoleInfoCommand is /roleinfo. Like /userinfo it needs no API calls.
type RoleInfoCommand struct{ D Deps }

// Definition describes /roleinfo.
func (RoleInfoCommand) Definition() *discordgo.ApplicationCommand {
	return definition("roleinfo", "Show a role's details",
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionRole, Name: "role",
			Description: "Role to look up", Required: true})
}

// Handle runs /roleinfo.
func (RoleInfoCommand) Handle(_ context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	id, ok, err := r.Role("role")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, commands.Fail("Role info failed", "Pick a role.")
	}
	role, found := r.ResolvedRole(id)
	if !found {
		return nil, commands.Fail("Role info failed", "Role not found.")
	}
	return roleEmbed(role, r.GuildID), nil
}

// keyPerms are the permissions worth calling out, in Discord's settings order.
var keyPerms = []int64{perms.Administrator, perms.ManageGuild, perms.ManageRoles, perms.ManageChannels,
	perms.KickMembers, perms.BanMembers, perms.ModerateMembers, perms.ManageMessages}

func roleEmbed(role *discordgo.Role, guildID string) *discordgo.MessageEmbed {
	mention := "<@&" + role.ID + ">"
	if role.ID == guildID {
		mention = "@everyone"
	}
	color := "None"
	if role.Color > 0 {
		color = fmt.Sprintf("#%06X", role.Color&0xFFFFFF)
	}
	var key []string
	for _, p := range keyPerms {
		if role.Permissions&p == p {
			key = append(key, perms.Name(p))
		}
	}
	keyText := "None"
	if len(key) > 0 {
		keyText = strings.Join(key, ", ")
	}
	return embeds.New(name(role.Name, "Role")).
		Field("Role", mention, true).
		Field("Color", color, true).
		Field("Position", itoa(max(role.Position, 0)), true).
		Field("Shown separately", yesNo(role.Hoist), true).
		Field("Mentionable", yesNo(role.Mentionable), true).
		Field("Managed", yesNo(role.Managed), true).
		Field("Created", created(role.ID), true).
		Field("Key permissions", keyText, false).
		Build()
}
