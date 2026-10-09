package info

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// Hostile names and a huge role list still produce a valid embed with no pings.
func TestUserInfoHostile(t *testing.T) {
	roles := make([]string, 5000)
	for i := range roles {
		roles[i] = rs
	}
	evil := strings.Repeat("@everyone [x](https://e.vil) `‮` ", 300)
	e, err := runUser(t, &resolved{
		Users:   map[string]*discordgo.User{us: {ID: us, Username: evil, GlobalName: evil}},
		Members: map[string]*discordgo.Member{us: {Nick: evil, Roles: roles}},
	}, userOpt(us))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	if strings.Contains(field(e, "Nickname"), "[x](") || !strings.HasPrefix(field(e, "Roles (5000)"), "<@&") {
		t.Fatalf("%+v", e.Fields)
	}
}
