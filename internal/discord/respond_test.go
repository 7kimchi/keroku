package discord

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestSlowmodeZeroIsSent(t *testing.T) {
	a, r := newServer(t, ok)
	if err := r.SetSlowmode(t.Context(), "3", 0, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(a.last().body), `"rate_limit_per_user":0`) {
		t.Fatalf("body %s", a.last().body)
	}
}

func TestRespondDisablesMentions(t *testing.T) {
	a, r := newServer(t, noContent)
	resp := &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "@everyone"}}
	if err := r.Respond(t.Context(), &discordgo.Interaction{ID: "1", Token: "t"}, resp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(a.last().body), `"allowed_mentions":{"parse":[]`) {
		t.Fatalf("body %s", a.last().body)
	}
}

func TestKindNames(t *testing.T) {
	for k := Unknown; k <= Timeout; k++ {
		if k != Unknown && k.String() == "unknown" {
			t.Fatalf("kind %d unnamed", k)
		}
	}
	if Is(nil, Unknown) || KindOf(context.Canceled) != Unknown {
		t.Fatal("unclassified errors")
	}
}
