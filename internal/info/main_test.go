package info

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const (
	gs   = "100000000000000001"
	self = "100000000000000003"
	us   = "100000000000000005"
	rs   = "100000000000000007"
)

// request parses an interaction run by self in guild gs.
func request(t testing.TB, name string, res *discordgo.ApplicationCommandInteractionDataResolved,
	opts ...*discordgo.ApplicationCommandInteractionDataOption) *commands.Request {
	t.Helper()
	i := &discordgo.Interaction{ID: "1300000000000000001", Type: discordgo.InteractionApplicationCommand,
		GuildID: gs, ChannelID: "100000000000000010",
		Member: &discordgo.Member{User: &discordgo.User{ID: self, Username: "me"}, Roles: []string{rs},
			JoinedAt: time.Unix(1700000000, 0), Permissions: 1 << 10},
		Data: discordgo.ApplicationCommandInteractionData{Name: name, CommandType: discordgo.ChatApplicationCommand,
			Options: opts, Resolved: res}}
	r, err := commands.Parse(i, "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func opt(name string, t discordgo.ApplicationCommandOptionType, v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: t, Value: v}
}

func field(e *discordgo.MessageEmbed, name string) string {
	for _, f := range e.Fields {
		if f.Name == name || strings.HasPrefix(f.Name, name+" (") {
			return f.Value
		}
	}
	return ""
}

// fits fails the test if e breaks any Discord limit or has an empty field.
func fits(t testing.TB, e *discordgo.MessageEmbed) {
	t.Helper()
	total := embeds.Length(e.Title) + embeds.Length(e.Description)
	if embeds.Length(e.Title) > embeds.MaxTitle || len(e.Fields) > embeds.MaxFields || e.Color != embeds.Color {
		t.Fatalf("embed over limits: %+v", e)
	}
	for _, f := range e.Fields {
		if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.Value) == "" || embeds.Length(f.Name) > embeds.MaxFieldName || embeds.Length(f.Value) > embeds.MaxFieldValue {
			t.Fatalf("bad field %q: %q", f.Name, f.Value)
		}
		total += embeds.Length(f.Name) + embeds.Length(f.Value)
	}
	if e.Footer != nil {
		total += embeds.Length(e.Footer.Text)
	}
	if total > embeds.MaxTotal || e.Image != nil || e.Thumbnail != nil || e.Author != nil {
		t.Fatalf("embed total %d or media set", total)
	}
}

// userErr asserts err is a refusal with detail, and whether it carries a cause for the log.
func userErr(t testing.TB, err error, detail string, logged bool) {
	t.Helper()
	var ue *commands.UserError
	if !errors.As(err, &ue) || ue.Detail != detail || (ue.Cause != nil) != logged {
		t.Fatalf("got %#v", err)
	}
}
