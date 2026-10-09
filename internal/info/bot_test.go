package info

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/discord"
)

func TestBotInfo(t *testing.T) {
	f := discord.NewFake()
	f.AddGuild(gs, self, 0)
	f.AddGuild("100000000000000009", self, 0)
	f.SetDelay("serverCount", 30*time.Millisecond)
	start := time.Now().Add(-26 * time.Hour)
	e, err := BotInfoCommand{D: Deps{Client: f, Started: start, Version: "v1.0.0"}}.
		Handle(context.Background(), request(t, "botinfo", nil))
	if err != nil {
		t.Fatal(err)
	}
	fits(t, e)
	if e.Title != "Keroku" || field(e, "Version") != "v1.0.0" {
		t.Fatalf("%+v", e)
	}
	if field(e, "Servers") != "2" || field(e, "Uptime") != "1d 2h" {
		t.Fatalf("%+v", e.Fields)
	}
	var ms int
	if _, err := fmt.Sscanf(field(e, "Latency"), "%d ms", &ms); err != nil || ms < 30 || ms > 1000 {
		t.Fatalf("latency %q", field(e, "Latency"))
	}
}

func TestBotInfoDefaults(t *testing.T) {
	f := discord.NewFake()
	now := time.Unix(1800000000, 0)
	e, err := BotInfoCommand{D: Deps{Client: f, Started: now, Now: func() time.Time { return now }}}.
		Handle(context.Background(), request(t, "botinfo", nil))
	if err != nil {
		t.Fatal(err)
	}
	if field(e, "Version") != "dev" || field(e, "Uptime") != "0s" || field(e, "Servers") != "0" || field(e, "Latency") != "0 ms" {
		t.Fatalf("%+v", e.Fields)
	}
}

// A clock that steps back never shows a negative latency.
func TestBotInfoClockSkew(t *testing.T) {
	calls := 0
	base := time.Unix(1800000000, 0)
	clock := func() time.Time { calls++; return base.Add(-time.Duration(calls) * time.Second) }
	e, err := BotInfoCommand{D: Deps{Client: discord.NewFake(), Started: base, Now: clock}}.
		Handle(context.Background(), request(t, "botinfo", nil))
	if err != nil || field(e, "Latency") != "0 ms" || field(e, "Uptime") != "0s" {
		t.Fatalf("err %v %+v", err, e)
	}
}

func TestBotInfoFailure(t *testing.T) {
	f := discord.NewFake()
	f.FailNext("serverCount", &discord.Error{Kind: discord.Unavailable, Status: 502}, 1)
	_, err := BotInfoCommand{D: Deps{Client: f}}.Handle(context.Background(), request(t, "botinfo", nil))
	userErr(t, err, "Discord did not respond. Try again.", true)
}
