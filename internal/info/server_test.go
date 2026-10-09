package info

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

func serverFake() *discord.Fake {
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", 0, &discordgo.Role{ID: rs, Name: "mod"})
	f.AddMember(gs, self)
	f.AddMember(gs, us)
	return f
}

func TestServerInfo(t *testing.T) {
	f := serverFake()
	e, err := ServerInfoCommand{D: Deps{Client: f}}.Handle(context.Background(), request(t, "serverinfo", nil))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	want := map[string]string{"Owner": "<@100000000000000002> (100000000000000002)", "Members": "2",
		"Online": "0", "Emojis": "0", "Boost level": "0", "Boosts": "0", "Verification": "None"}
	for k, v := range want {
		if got := field(e, k); got != v {
			t.Fatalf("%s: %q", k, got)
		}
	}
	if !strings.HasPrefix(field(e, "Created"), "<t:") || e.Footer.Text != "ID "+gs {
		t.Fatalf("%+v", e)
	}
	if f.Calls("guildCounts") != 1 {
		t.Fatal("expected one call")
	}
}

// Whatever Discord sends back, the embed fits and never pings.
func TestServerEmbedHostile(t *testing.T) {
	g := &discordgo.Guild{Name: strings.Repeat("@everyone **x** ", 500), OwnerID: "<@&1>",
		ApproximateMemberCount: -4, ApproximatePresenceCount: -1, PremiumTier: -2, PremiumSubscriptionCount: -9,
		VerificationLevel: 99}
	e := serverEmbed(g, gs)
	fits(t, e)
	if field(e, "Owner") != "Unknown" || field(e, "Members") != "0" || field(e, "Boost level") != "0" ||
		field(e, "Verification") != "Unknown" {
		t.Fatalf("%+v", e.Fields)
	}
	if strings.Contains(e.Title, "**x**") {
		t.Fatal("markdown not escaped")
	}
	if serverEmbed(&discordgo.Guild{}, gs).Title != "Server" {
		t.Fatal("empty name not replaced")
	}
}

func TestVerificationNames(t *testing.T) {
	want := []string{"None", "Low", "Medium", "High", "Highest"}
	for i, w := range want {
		if got := verification(discordgo.VerificationLevel(i)); got != w {
			t.Fatalf("%d: %q", i, got)
		}
	}
}

func TestServerInfoFailures(t *testing.T) {
	cases := []struct {
		err    error
		detail string
		logged bool
	}{
		{&discord.Error{Kind: discord.NotFound, Status: 404}, "Not found.", false},
		{&discord.Error{Kind: discord.Unavailable, Status: 503}, "Discord did not respond. Try again.", true},
		{&discord.Error{Kind: discord.RateLimited, Status: 429}, "Discord did not respond. Try again.", true},
		{&discord.Error{Kind: discord.Forbidden, Status: 403}, "Discord did not respond. Try again.", true},
	}
	for _, c := range cases {
		f := serverFake()
		f.FailNext("guildCounts", c.err, 1)
		_, err := ServerInfoCommand{D: Deps{Client: f}}.Handle(context.Background(), request(t, "serverinfo", nil))
		userErr(t, err, c.detail, c.logged)
	}
}

func TestServerInfoTimeout(t *testing.T) {
	f := serverFake()
	f.SetDelay("guildCounts", time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := ServerInfoCommand{D: Deps{Client: f}}.Handle(ctx, request(t, "serverinfo", nil))
	userErr(t, err, "Discord did not respond. Try again.", true)
}
