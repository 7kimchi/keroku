package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestImplementations(t *testing.T) {
	var _ Client = NewFake()
	var _ Client = (*REST)(nil)
}

func seeded() *Fake {
	f := NewFake()
	f.AddGuild("g", "owner", 0, &discordgo.Role{ID: "mod", Position: 5})
	f.AddMember("g", "u", "mod")
	f.AddChannel("g", "c")
	return f
}

func TestFakeMemberLifecycle(t *testing.T) {
	f, ctx := seeded(), t.Context()
	m, err := f.Member(ctx, "g", "u")
	if err != nil || m.Roles[0] != "mod" {
		t.Fatalf("got %v %v", m, err)
	}
	m.Roles[0] = "hacked"
	if again, _ := f.Member(ctx, "g", "u"); again.Roles[0] != "mod" {
		t.Fatal("caller mutation leaked into fake")
	}
	until := time.Now().Add(time.Hour)
	if err := f.Timeout(ctx, "g", "u", &until, ""); err != nil || f.TimeoutUntil("g", "u") == nil {
		t.Fatalf("timeout: %v", err)
	}
	tooLong := time.Now().Add(29 * 24 * time.Hour)
	if !Is(f.Timeout(ctx, "g", "u", &tooLong, ""), BadRequest) {
		t.Fatal("over 28 days accepted")
	}
	if err := f.Ban(ctx, "g", "u", 0, ""); err != nil || !f.Banned("g", "u") || f.IsMember("g", "u") {
		t.Fatalf("ban: %v", err)
	}
	if err := f.Ban(ctx, "g", "u", 0, ""); err != nil {
		t.Fatal("double ban should succeed")
	}
	if !Is(f.Kick(ctx, "g", "u", ""), NotFound) || !Is(f.Unban(ctx, "g", "x", ""), NotFound) {
		t.Fatal("missing targets must be NotFound")
	}
	if err := f.Unban(ctx, "g", "u", ""); err != nil || f.Banned("g", "u") {
		t.Fatalf("unban: %v", err)
	}
	if !Is(f.Ban(ctx, "g", "u", 604801, ""), BadRequest) {
		t.Fatal("delete window over 7 days accepted")
	}
}

func TestFakeFailureInjection(t *testing.T) {
	f, ctx := seeded(), t.Context()
	boom := &Error{Op: "ban", Kind: Unavailable, Status: 503}
	f.FailNext("ban", boom, 2)
	for range 2 {
		if !errors.Is(f.Ban(ctx, "g", "u", 0, ""), boom) {
			t.Fatal("injected failure not returned")
		}
	}
	if err := f.Ban(ctx, "g", "u", 0, ""); err != nil || f.Calls("ban") != 3 {
		t.Fatalf("after failures: %v calls %d", err, f.Calls("ban"))
	}
	f.FailNext("*", boom, 1)
	if f.Send(ctx, "c", &discordgo.MessageEmbed{}) == nil {
		t.Fatal("wildcard failure ignored")
	}
}

func TestFakeDelayHonorsContext(t *testing.T) {
	f := seeded()
	f.SetDelay("member", time.Second)
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := f.Member(ctx, "g", "u"); !Is(err, Timeout) || time.Since(start) > 500*time.Millisecond {
		t.Fatalf("got %v after %v", err, time.Since(start))
	}
}

func TestFakeOnCall(t *testing.T) {
	f := seeded()
	n := 0
	f.OnCall("kick", func() { n++ })
	_ = f.Kick(t.Context(), "g", "u", "")
	_ = f.Kick(t.Context(), "g", "u", "")
	if n != 2 {
		t.Fatalf("hook ran %d times", n)
	}
}
