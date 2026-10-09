package discord

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestBanSendsSecondsReasonAndAuth(t *testing.T) {
	a, rest := newServer(t, noContent)
	reason := "spam \U00004E16 & more\nline"
	if err := rest.Ban(t.Context(), "10", "20", 3600, reason); err != nil {
		t.Fatal(err)
	}
	req := a.last()
	if req.method != "PUT" || !strings.HasSuffix(req.path, "/guilds/10/bans/20") {
		t.Fatalf("got %s %s", req.method, req.path)
	}
	var body map[string]int
	if err := json.Unmarshal(req.body, &body); err != nil || body["delete_message_seconds"] != 3600 {
		t.Fatalf("body %s", req.body)
	}
	if decoded, _ := url.PathUnescape(req.reason); decoded != reason {
		t.Fatalf("reason header %q", req.reason)
	}
	if req.auth != "Bot "+testToken {
		t.Fatal("auth header missing")
	}
}

func TestAuditReasonIsCapped(t *testing.T) {
	a, rest := newServer(t, noContent)
	if err := rest.Kick(t.Context(), "1", "2", strings.Repeat("\U0001F600", 600)); err != nil {
		t.Fatal(err)
	}
	decoded, _ := url.PathUnescape(a.last().reason)
	if n := len([]rune(decoded)); n != MaxAuditReason {
		t.Fatalf("reason has %d runes", n)
	}
	_ = rest.Kick(t.Context(), "1", "2", "")
	if a.last().reason != "" {
		t.Fatal("empty reason should send no header")
	}
}

func TestSendDisablesMentions(t *testing.T) {
	a, rest := newServer(t, ok)
	if err := rest.Send(t.Context(), "5", testEmbed()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(a.last().body), `"allowed_mentions":{"parse":[]`) {
		t.Fatalf("body %s", a.last().body)
	}
}

func TestTimeoutClearSendsNull(t *testing.T) {
	a, rest := newServer(t, ok)
	if err := rest.Timeout(t.Context(), "1", "2", nil, "x"); err != nil {
		t.Fatal(err)
	}
	if string(a.last().body) != `{"communication_disabled_until":null}` {
		t.Fatalf("body %s", a.last().body)
	}
	until := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = rest.Timeout(t.Context(), "1", "2", &until, "x")
	if !strings.Contains(string(a.last().body), "2030-01-01T00:00:00Z") {
		t.Fatalf("body %s", a.last().body)
	}
}

func TestNilEmbedRejectedWithoutCall(t *testing.T) {
	a, rest := newServer(t, ok)
	i := &discordgo.Interaction{AppID: "1", Token: "t"}
	for _, err := range []error{rest.Send(t.Context(), "1", nil), rest.DM(t.Context(), "1", nil), rest.EditResponse(t.Context(), i, nil)} {
		if !Is(err, BadRequest) {
			t.Fatalf("got %v", err)
		}
	}
	if a.count() != 0 {
		t.Fatal("nil embed reached the API")
	}
}
