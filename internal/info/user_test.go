package info

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

type resolved = discordgo.ApplicationCommandInteractionDataResolved

func userOpt(v any) *discordgo.ApplicationCommandInteractionDataOption {
	return opt("user", discordgo.ApplicationCommandOptionUser, v)
}

func runUser(t *testing.T, res *resolved, opts ...*discordgo.ApplicationCommandInteractionDataOption) (*discordgo.MessageEmbed, error) {
	t.Helper()
	now := time.Unix(1800000000, 0)
	return UserInfoCommand{D: Deps{Now: func() time.Time { return now }}}.
		Handle(context.Background(), request(t, "userinfo", res, opts...))
}

func TestUserInfoSelf(t *testing.T) {
	e, err := runUser(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	if e.Title != "me" || !strings.Contains(field(e, "User"), self) || field(e, "Roles") != "<@&"+rs+">" ||
		field(e, "Joined") != "<t:1700000000:R>" || field(e, "Member") != "" {
		t.Fatalf("%+v %+v", e, e.Fields)
	}
}

func TestUserInfoMember(t *testing.T) {
	until := time.Unix(1900000000, 0)
	e, err := runUser(t, &resolved{
		Users:   map[string]*discordgo.User{us: {ID: us, Username: "target", GlobalName: "Target", Bot: true}},
		Members: map[string]*discordgo.Member{us: {Nick: "nick", CommunicationDisabledUntil: &until}},
	}, userOpt(us))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	want := map[string]string{"Display name": "Target", "Bot": "Yes", "Nickname": "nick",
		"Timed out until": "<t:1900000000:R>", "Roles": "None"}
	for k, v := range want {
		if got := field(e, k); got != v {
			t.Fatalf("%s: %q", k, got)
		}
	}
	if field(e, "Joined") != "" {
		t.Fatal("zero join time shown")
	}
}

func TestUserInfoExpiredTimeoutHidden(t *testing.T) {
	past := time.Unix(1000000000, 0)
	e, _ := runUser(t, &resolved{
		Users:   map[string]*discordgo.User{us: {ID: us, Username: "t"}},
		Members: map[string]*discordgo.Member{us: {CommunicationDisabledUntil: &past}},
	}, userOpt(us))
	if field(e, "Timed out until") != "" {
		t.Fatal("expired timeout shown")
	}
}

func TestUserInfoNotMember(t *testing.T) {
	e, err := runUser(t, &resolved{Users: map[string]*discordgo.User{us: {ID: us, Username: "gone"}}}, userOpt(us))
	if err != nil || field(e, "Member") != "Not in this server" || field(e, "Roles") != "" {
		t.Fatalf("err %v %+v", err, e)
	}
}

// Forged payloads: the option names a user the resolved data does not back up.
func TestUserInfoForged(t *testing.T) {
	for _, res := range []*resolved{nil, {}, {Users: map[string]*discordgo.User{us: {ID: self}}}} {
		_, err := runUser(t, res, userOpt(us))
		userErr(t, err, "User not found.", false)
	}
	for _, v := range []any{"abc", "", 7, "0", "-100000000000000005"} {
		if _, err := runUser(t, nil, userOpt(v)); err == nil {
			t.Fatalf("%v accepted", v)
		}
	}
}

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
