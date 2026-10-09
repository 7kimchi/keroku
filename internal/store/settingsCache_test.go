package store_test

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestSettingsCache(t *testing.T) {
	s := store.New(dbtest.New(t))
	c, err := store.NewSettingsCache(s, 10, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if g, _ := c.Get(ctx, 1); g.ModlogChannelID != 0 {
		t.Fatal("expected empty")
	}
	if err := c.SetModlog(ctx, 1, 5); err != nil {
		t.Fatal(err)
	}
	if g, _ := c.Get(ctx, 1); g.ModlogChannelID != 5 {
		t.Fatal("write through did not invalidate")
	}
	_ = c.SetLog(ctx, 1, 6)
	if g, _ := c.Get(ctx, 1); g.LogChannelID != 6 {
		t.Fatal("log channel not visible")
	}
	// A write from another instance shows up only after the TTL; this one is cached.
	_ = s.SetModlogChannel(ctx, 1, 9)
	if g, _ := c.Get(ctx, 1); g.ModlogChannelID != 5 {
		t.Fatal("cache bypassed")
	}
	c.Sweep()
	if _, err := store.NewSettingsCache(s, 0, time.Hour); err == nil {
		t.Fatal("zero size accepted")
	}
}

func TestSettingsCacheDatabaseDown(t *testing.T) {
	s := store.New(dbtest.New(t))
	c, _ := store.NewSettingsCache(s, 10, time.Hour)
	s.Close()
	if _, err := c.Get(t.Context(), 1); err == nil {
		t.Fatal("read from closed pool succeeded")
	}
}
