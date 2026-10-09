package store_test

import (
	"testing"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestGuildSettings(t *testing.T) {
	s := store.New(dbtest.New(t))
	ctx := t.Context()
	if g, err := s.GuildSettings(ctx, 1); err != nil || g != (store.GuildSettings{}) {
		t.Fatalf("missing guild: %+v %v", g, err)
	}
	if err := s.SetModlogChannel(ctx, 1, 10); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLogChannel(ctx, 1, 20); err != nil {
		t.Fatal(err)
	}
	if g, _ := s.GuildSettings(ctx, 1); g.ModlogChannelID != 10 || g.LogChannelID != 20 {
		t.Fatalf("got %+v", g)
	}
	_ = s.SetModlogChannel(ctx, 1, 0)
	if g, _ := s.GuildSettings(ctx, 1); g.ModlogChannelID != 0 || g.LogChannelID != 20 {
		t.Fatalf("clear changed the wrong field: %+v", g)
	}
	if g, _ := s.GuildSettings(ctx, 2); g != (store.GuildSettings{}) {
		t.Fatal("settings leaked across guilds")
	}
	if err := s.SetModlogChannel(ctx, 1, -5); err == nil {
		t.Fatal("negative channel accepted")
	}
}

func TestGuildSettingsRealisticIDs(t *testing.T) {
	s := store.New(dbtest.New(t))
	const g, ch = int64(1300000000000000001), int64(1300000000000000002)
	if err := s.SetModlogChannel(t.Context(), g, ch); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLogChannel(t.Context(), g, ch+1); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GuildSettings(t.Context(), g); got.ModlogChannelID != ch || got.LogChannelID != ch+1 {
		t.Fatalf("got %+v", got)
	}
}
