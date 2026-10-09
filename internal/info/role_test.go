package info

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/perms"
)

func roleOpt(v any) *discordgo.ApplicationCommandInteractionDataOption {
	return opt("role", discordgo.ApplicationCommandOptionRole, v)
}

func runRole(t *testing.T, res *resolved, opts ...*discordgo.ApplicationCommandInteractionDataOption) (*discordgo.MessageEmbed, error) {
	t.Helper()
	return RoleInfoCommand{}.Handle(context.Background(), request(t, "roleinfo", res, opts...))
}

func TestRoleInfo(t *testing.T) {
	role := &discordgo.Role{ID: rs, Name: "Mods", Color: 0x5865F2, Position: 4, Hoist: true,
		Permissions: perms.BanMembers | perms.KickMembers | perms.SendMessages}
	e, err := runRole(t, &resolved{Roles: map[string]*discordgo.Role{rs: role}}, roleOpt(rs))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	want := map[string]string{"Role": "<@&" + rs + ">", "Color": "#5865F2", "Position": "4",
		"Shown separately": "Yes", "Mentionable": "No", "Managed": "No", "Key permissions": "Kick Members, Ban Members"}
	for k, v := range want {
		if got := field(e, k); got != v {
			t.Fatalf("%s: %q", k, got)
		}
	}
	if e.Title != "Mods" || !strings.HasPrefix(field(e, "Created"), "<t:") {
		t.Fatalf("%+v", e)
	}
}

// @everyone shares the guild id. Its mention would ping everyone if allowed, so it is plain text.
func TestRoleInfoEveryone(t *testing.T) {
	e, err := runRole(t, &resolved{Roles: map[string]*discordgo.Role{gs: {ID: gs, Name: "@everyone"}}}, roleOpt(gs))
	if err != nil || field(e, "Role") != "@everyone" || field(e, "Color") != "None" || field(e, "Key permissions") != "None" {
		t.Fatalf("err %v %+v", err, e)
	}
}

func TestRoleInfoAdministrator(t *testing.T) {
	e, _ := runRole(t, &resolved{Roles: map[string]*discordgo.Role{rs: {ID: rs, Permissions: perms.Administrator}}}, roleOpt(rs))
	if field(e, "Key permissions") != "Administrator" || e.Title != "Role" {
		t.Fatalf("%+v", e)
	}
}

func TestRoleInfoRejects(t *testing.T) {
	_, err := runRole(t, nil)
	userErr(t, err, "Pick a role.", false)
	for _, res := range []*resolved{nil, {Roles: map[string]*discordgo.Role{rs: {ID: us}}}} {
		_, err := runRole(t, res, roleOpt(rs))
		userErr(t, err, "Role not found.", false)
	}
	for _, v := range []any{"x", "", 3, "0"} {
		if _, err := runRole(t, nil, roleOpt(v)); err == nil {
			t.Fatalf("%v accepted", v)
		}
	}
}

func TestRoleInfoHostile(t *testing.T) {
	role := &discordgo.Role{ID: rs, Name: strings.Repeat("**@here** ", 200), Color: -1, Position: -3}
	e, err := runRole(t, &resolved{Roles: map[string]*discordgo.Role{rs: role}}, roleOpt(rs))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	if field(e, "Color") != "None" || field(e, "Position") != "0" || strings.Contains(e.Title, "**@here**") {
		t.Fatalf("%+v", e)
	}
}
