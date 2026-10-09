package commands

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func withResolved(t *testing.T, res *discordgo.ApplicationCommandInteractionDataResolved,
	opts ...*discordgo.ApplicationCommandInteractionDataOption) *Request {
	t.Helper()
	i := interaction("x", opts...)
	d := i.Data.(discordgo.ApplicationCommandInteractionData)
	d.Resolved = res
	i.Data = d
	r, err := Parse(i, "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestResolvedNil(t *testing.T) {
	r := withResolved(t, nil)
	if _, ok := r.ResolvedUser(userA); ok {
		t.Fatal("user from nothing")
	}
	if r.ResolvedMember(userA) != nil {
		t.Fatal("member from nothing")
	}
	if _, ok := r.ResolvedRole(userA); ok {
		t.Fatal("role from nothing")
	}
}

func TestResolvedFound(t *testing.T) {
	r := withResolved(t, &discordgo.ApplicationCommandInteractionDataResolved{
		Users:   map[string]*discordgo.User{userA: {ID: userA, Username: "a"}},
		Members: map[string]*discordgo.Member{userA: {Nick: "n"}},
		Roles:   map[string]*discordgo.Role{userB: {ID: userB, Name: "r"}},
	})
	if u, ok := r.ResolvedUser(userA); !ok || u.Username != "a" {
		t.Fatal("user missing")
	}
	if m := r.ResolvedMember(userA); m == nil || m.Nick != "n" {
		t.Fatal("member missing")
	}
	if role, ok := r.ResolvedRole(userB); !ok || role.Name != "r" {
		t.Fatal("role missing")
	}
}

// A payload whose map key and object id disagree is not trusted.
func TestResolvedMismatchedIDs(t *testing.T) {
	r := withResolved(t, &discordgo.ApplicationCommandInteractionDataResolved{
		Users:   map[string]*discordgo.User{userA: {ID: userB}, userB: nil},
		Members: map[string]*discordgo.Member{userB: nil},
		Roles:   map[string]*discordgo.Role{userA: {ID: userB}, userB: nil},
	})
	for _, id := range []string{userA, userB, ""} {
		if _, ok := r.ResolvedUser(id); ok {
			t.Fatalf("user %q", id)
		}
		if _, ok := r.ResolvedRole(id); ok {
			t.Fatalf("role %q", id)
		}
		if r.ResolvedMember(id) != nil {
			t.Fatalf("member %q", id)
		}
	}
}

func TestRoleOption(t *testing.T) {
	r := withResolved(t, nil, opt("role", discordgo.ApplicationCommandOptionRole, userB))
	if id, ok, err := r.Role("role"); err != nil || !ok || id != userB {
		t.Fatalf("%q %v %v", id, ok, err)
	}
	for _, v := range []any{"abc", "", 5, "0"} {
		r := withResolved(t, nil, opt("role", discordgo.ApplicationCommandOptionRole, v))
		if _, _, err := r.Role("role"); err == nil {
			t.Fatalf("%v accepted", v)
		}
	}
	r = withResolved(t, nil, opt("role", discordgo.ApplicationCommandOptionUser, userB))
	if _, _, err := r.Role("role"); err == nil {
		t.Fatal("user option read as role")
	}
}
