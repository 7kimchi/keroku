package gateway

import (
	"log/slog"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/gatewaytest"
	"github.com/7kimchi/keroku/internal/safe"
)

func TestOpenSubsetOfShards(t *testing.T) {
	f := gatewaytest.New(t, 1)
	gw := New(Config{Token: "fake.test.token", ShardCount: 6, ShardIDs: []int{2, 5}, IdentifyWait: time.Millisecond,
		HTTPClient: f.Client()}, &recorder{}, safe.NewGuard(slog.New(slog.DiscardHandler), nil), slog.New(slog.DiscardHandler))
	if err := gw.Open(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer gw.Close()
	waitFor(t, "two identifies", func() bool { return len(f.Identified()) == 2 })
	for _, id := range f.Identified() {
		shard := id["shard"].([]any)
		if n := shard[1].(float64); n != 6 {
			t.Fatalf("shard count %v", n)
		}
		if s := shard[0].(float64); s != 2 && s != 5 {
			t.Fatalf("unexpected shard %v", s)
		}
	}
}

func TestReconnectAfterDrop(t *testing.T) {
	f := gatewaytest.New(t, 1)
	gw := New(Config{Token: "fake.test.token", HTTPClient: f.Client()}, &recorder{},
		safe.NewGuard(slog.New(slog.DiscardHandler), nil), slog.New(slog.DiscardHandler))
	if err := gw.Open(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer gw.Close()
	waitFor(t, "first identify", func() bool { return len(f.Identified()) == 1 })
	f.DropAll()
	waitFor(t, "reconnect", func() bool { return len(f.Identified()) >= 2 })
}
