package gateway

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/gatewaytest"
	"github.com/7kimchi/keroku/internal/logging"
)

func TestInstallLoggerRoutesAndRedacts(t *testing.T) {
	var buf bytes.Buffer
	InstallLogger(logging.New(&buf, slog.LevelDebug, logging.NewRedactor("fake.test.token")))
	t.Cleanup(func() { discordgo.Logger = nil })
	discordgo.Logger(discordgo.LogError, 0, "boom with %s", "fake.test.token")
	discordgo.Logger(discordgo.LogWarning, 0, "warned")
	discordgo.Logger(discordgo.LogDebug, 0, "noise")
	out := buf.String()
	if strings.Contains(out, "fake.test.token") || !strings.Contains(out, `"level":"ERROR"`) ||
		!strings.Contains(out, `"level":"WARN"`) || !strings.Contains(out, "noise") {
		t.Fatalf("got %s", out)
	}
}

func TestErrText(t *testing.T) {
	if got := errText(&discord.Error{Op: "x", Kind: discord.Unauthorized}); got != "unauthorized" {
		t.Fatalf("got %s", got)
	}
	rest := &discordgo.RESTError{Response: &http.Response{Status: "401 Unauthorized"},
		Request: &http.Request{Header: http.Header{"Authorization": {"Bot fake.test.token"}}}}
	if got := errText(rest); got != "http status 401 Unauthorized" {
		t.Fatalf("got %s", got)
	}
	if got := errText(errors.New("dial failed")); got != "dial failed" {
		t.Fatalf("got %s", got)
	}
}

func TestOpenFailsCleanlyWithBadToken(t *testing.T) {
	f := gatewaytest.New(t, 1)
	f.Reject = true
	gw := New(Config{Token: "fake.test.token", HTTPClient: f.Client()}, &recorder{}, nil, slog.New(slog.DiscardHandler))
	err := gw.Open(t.Context())
	if err == nil || strings.Contains(err.Error(), "fake.test.token") {
		t.Fatalf("got %v", err)
	}
}
