package store_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestAutomodSettings(t *testing.T) {
	s := store.New(dbtest.New(t))
	ctx := t.Context()
	const g = int64(1300000000000000001)
	if a, err := s.Automod(ctx, g); err != nil || !reflect.DeepEqual(a, store.DefaultAutomod()) {
		t.Fatalf("defaults %+v %v", a, err)
	}
	want := store.AutomodSettings{SpamEnabled: true, SpamMessages: 5, SpamWindow: 3 * time.Second, DuplicateEnabled: true,
		DuplicateCount: 4, DuplicateWindow: time.Minute, LinksEnabled: true, AllowedDomains: []string{"example.com"},
		Timeout: 10 * time.Minute, MentionLimit: 8, InvitesBlocked: true}
	if err := s.SetAutomod(ctx, g, want); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Automod(ctx, g); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if other, _ := s.Automod(ctx, g+1); !reflect.DeepEqual(other, store.DefaultAutomod()) {
		t.Fatal("settings leaked across guilds")
	}
	bad := want
	bad.SpamMessages = 1
	if err := s.SetAutomod(ctx, g, bad); err == nil {
		t.Fatal("out of range accepted")
	}
	bad = want
	bad.AllowedDomains = make([]string, 51)
	if err := s.SetAutomod(ctx, g, bad); err == nil {
		t.Fatal("too many domains accepted")
	}
}

func TestRaidSettings(t *testing.T) {
	s := store.New(dbtest.New(t))
	ctx := t.Context()
	const g = int64(1300000000000000001)
	if r, _ := s.Raid(ctx, g); r != store.DefaultRaid() {
		t.Fatalf("defaults %+v", r)
	}
	want := store.RaidSettings{Enabled: true, JoinLimit: 20, Window: 30 * time.Second, MinAccountAge: 24 * time.Hour,
		Action: "lockdown", LockdownFor: time.Hour}
	if err := s.SetRaid(ctx, g, want); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Raid(ctx, g); got != want {
		t.Fatalf("got %+v", got)
	}
	want.Action = "nuke"
	if err := s.SetRaid(ctx, g, want); err == nil {
		t.Fatal("bad action accepted")
	}
}
