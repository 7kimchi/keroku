package app

import (
	"log/slog"
	"testing"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

func TestStartupFailureStillShutsDown(t *testing.T) {
	a, fake := newTestApp(t, testConfig())
	fake.FailNext("overwriteCommands", &discord.Error{Op: "overwriteCommands", Kind: discord.Unauthorized}, 1)
	cancel, done := runApp(t, a)
	defer cancel()
	if err := waitDone(t, done); !discord.Is(err, discord.Unauthorized) {
		t.Fatalf("got %v", err)
	}
	if a.ack.Submit("x", nil) {
		t.Fatal("pools still accept work after a failed start")
	}
}

func TestIdentityFailureStopsAssembly(t *testing.T) {
	fake := discord.NewFake()
	fake.FailNext("identity", &discord.Error{Op: "identity", Kind: discord.Unauthorized}, 1)
	if _, err := assemble(testConfig(), slog.New(slog.DiscardHandler), store.New(dbtest.New(t)), fake); err == nil {
		t.Fatal("assembled without an identity")
	}
}
