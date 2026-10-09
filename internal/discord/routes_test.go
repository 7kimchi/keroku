package discord

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestEveryMethodHitsItsRoute(t *testing.T) {
	a, r := newServer(t, func(w http.ResponseWriter, req *http.Request, _ int) {
		switch {
		case strings.HasSuffix(req.URL.Path, "/users/@me/channels"):
			reply(w, 200, `{"id":"dm1"}`)
		case req.Method == "GET" && strings.HasSuffix(req.URL.Path, "/messages"):
			reply(w, 200, `[{"id":"m1"}]`)
		case req.Method == "GET" && strings.HasSuffix(req.URL.Path, "/rules"):
			reply(w, 200, `[]`)
		case req.Method == "PUT" && strings.Contains(req.URL.Path, "/commands"):
			reply(w, 200, `[]`)
		default:
			reply(w, 200, `{"id":"1"}`)
		}
	})
	ctx := t.Context()
	i := &discordgo.Interaction{ID: "5", AppID: "6", Token: "tok"}
	rule := &discordgo.AutoModerationRule{Name: "x"}
	steps := []struct {
		call   func(context.Context) error
		method string
		path   string
	}{
		{func(c context.Context) error { _, e := r.Guild(c, "1"); return e }, "GET", "/guilds/1"},
		{func(c context.Context) error { _, e := r.Member(c, "1", "2"); return e }, "GET", "/guilds/1/members/2"},
		{func(c context.Context) error { return r.Unban(c, "1", "2", "r") }, "DELETE", "/guilds/1/bans/2"},
		{func(c context.Context) error { return r.Kick(c, "1", "2", "r") }, "DELETE", "/guilds/1/members/2"},
		{func(c context.Context) error { _, e := r.Channel(c, "3"); return e }, "GET", "/channels/3"},
		{func(c context.Context) error { return r.DM(c, "2", testEmbed()) }, "POST", "/channels/dm1/messages"},
		{func(c context.Context) error { _, e := r.Messages(c, "3", 100, "9"); return e }, "GET", "/channels/3/messages"},
		{func(c context.Context) error { return r.BulkDelete(c, "3", []string{"a", "b"}, "r") }, "POST", "/channels/3/messages/bulk-delete"},
		{func(c context.Context) error { return r.DeleteMessage(c, "3", "a", "r") }, "DELETE", "/channels/3/messages/a"},
		{func(c context.Context) error { return r.SetSlowmode(c, "3", 0, "r") }, "PATCH", "/channels/3"},
		{func(c context.Context) error { return r.SetRoleOverwrite(c, "3", "1", 0, 2048, "r") }, "PUT", "/channels/3/permissions/1"},
		{func(c context.Context) error { return r.DeleteOverwrite(c, "3", "1", "r") }, "DELETE", "/channels/3/permissions/1"},
		{func(c context.Context) error { return r.SetRolePermissions(c, "1", "1", 1024, "r") }, "PATCH", "/guilds/1/roles/1"},
		{func(c context.Context) error {
			return r.Respond(c, i, &discordgo.InteractionResponse{Type: 5, Data: &discordgo.InteractionResponseData{}})
		}, "POST", "/interactions/5/tok/callback"},
		{func(c context.Context) error { return r.EditResponse(c, i, testEmbed()) }, "PATCH", "/webhooks/6/tok/messages/@original"},
		{func(c context.Context) error { _, e := r.AutoModRules(c, "1"); return e }, "GET", "/guilds/1/auto-moderation/rules"},
		{func(c context.Context) error { return r.CreateAutoModRule(c, "1", rule, "r") }, "POST", "/guilds/1/auto-moderation/rules"},
		{func(c context.Context) error { return r.EditAutoModRule(c, "1", "7", rule, "r") }, "PATCH", "/guilds/1/auto-moderation/rules/7"},
		{func(c context.Context) error { return r.DeleteAutoModRule(c, "1", "7", "r") }, "DELETE", "/guilds/1/auto-moderation/rules/7"},
		{func(c context.Context) error { return r.OverwriteCommands(c, "6", nil) }, "PUT", "/applications/6/commands"},
	}
	for _, s := range steps {
		if err := s.call(ctx); err != nil {
			t.Fatalf("%s %s: %v", s.method, s.path, err)
		}
		got := a.last()
		if got.method != s.method || !strings.HasSuffix(got.path, s.path) {
			t.Fatalf("want %s %s, got %s %s", s.method, s.path, got.method, got.path)
		}
	}
}

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
