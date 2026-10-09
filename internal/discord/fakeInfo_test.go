package discord

import (
	"context"
	"testing"
	"time"
)

func TestFakeGuildCounts(t *testing.T) {
	f := NewFake()
	f.AddGuild("1", "2", 0)
	f.AddMember("1", "2")
	f.AddMember("1", "3")
	g, err := f.GuildCounts(context.Background(), "1")
	if err != nil || g.ApproximateMemberCount != 2 {
		t.Fatalf("g %+v err %v", g, err)
	}
	g.Name = "changed"
	if again, _ := f.GuildCounts(context.Background(), "1"); again.Name == "changed" {
		t.Fatal("fake returned its own guild")
	}
	if _, err := f.GuildCounts(context.Background(), "9"); !Is(err, NotFound) {
		t.Fatal(err)
	}
}

func TestFakeServerCount(t *testing.T) {
	f := NewFake()
	f.AddGuild("1", "2", 0)
	f.AddGuild("3", "2", 0)
	if n, err := f.ServerCount(context.Background()); err != nil || n != 2 {
		t.Fatalf("n %d err %v", n, err)
	}
	f.FailNext("serverCount", &Error{Kind: Unavailable}, 1)
	if _, err := f.ServerCount(context.Background()); !Is(err, Unavailable) {
		t.Fatal(err)
	}
	f.SetDelay("guildCounts", time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := f.GuildCounts(ctx, "1"); !Is(err, Timeout) {
		t.Fatal(err)
	}
}
