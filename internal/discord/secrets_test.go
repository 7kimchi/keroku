package discord

import (
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestErrorsNeverCarrySecrets(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 404, `{"code":10015,"message":"Unknown Webhook"}`)
	})
	i := &discordgo.Interaction{AppID: "1", ID: "2", Token: "interactionsecrettoken"}
	err := rest.EditResponse(t.Context(), i, testEmbed())
	if !Is(err, Expired) {
		t.Fatalf("got %v", err)
	}
	if s := err.Error(); strings.Contains(s, "interactionsecrettoken") || strings.Contains(s, testToken) || strings.Contains(s, "http") {
		t.Fatalf("leak: %s", s)
	}
}
