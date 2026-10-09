package app

import (
	"context"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/gatewaytest"
)

const unknownCommand = `"t":"INTERACTION_CREATE","d":{"id":"1300000000000000001","type":2,"guild_id":"100000000000000001",` +
	`"channel_id":"100000000000000002","member":{"user":{"id":"100000000000000003"},"permissions":"0"},"data":{"name":"nope","type":1}}`

func runApp(t *testing.T, a *App) (context.CancelFunc, chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	return cancel, done
}

func waitDone(t *testing.T, done chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(15 * time.Second):
		t.Fatal("Run did not return")
		return nil
	}
}

func TestRunServesInteractionsAndShutsDown(t *testing.T) {
	gw := gatewaytest.New(t, 2)
	a, fake := newTestApp(t, testConfig())
	a.gatewayHTTP = gw.Client()
	cancel, done := runApp(t, a)
	deadline := time.Now().Add(5 * time.Second)
	for len(gw.Identified()) < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	gw.Broadcast(unknownCommand)
	for !fake.Responded("1300000000000000001") && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	replies := fake.Replies("1300000000000000001")
	if len(replies) != 1 || replies[0].Description != "Unknown command." {
		t.Fatalf("replies %+v", replies)
	}
	if fake.Calls("overwriteCommands") != 1 {
		t.Fatal("commands not registered by the shard 0 instance")
	}
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
}

func TestOnlyShardZeroRegisters(t *testing.T) {
	gw := gatewaytest.New(t, 1)
	cfg := testConfig()
	cfg.ShardCount, cfg.ShardIDs = 4, []int{2, 3}
	a, fake := newTestApp(t, cfg)
	a.gatewayHTTP = gw.Client()
	cancel, done := runApp(t, a)
	deadline := time.Now().Add(5 * time.Second)
	for len(gw.Identified()) < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
	if fake.Calls("overwriteCommands") != 0 {
		t.Fatal("non zero shard instance registered commands")
	}
}

func TestStartupFailureStillShutsDown(t *testing.T) {
	a, fake := newTestApp(t, testConfig())
	fake.FailNext("identity", &discord.Error{Op: "identity", Kind: discord.Unauthorized}, 1)
	cancel, done := runApp(t, a)
	defer cancel()
	err := waitDone(t, done)
	if !discord.Is(err, discord.Unauthorized) {
		t.Fatalf("got %v", err)
	}
	if a.ack.Submit("x", nil) {
		t.Fatal("pools still accept work after a failed start")
	}
}
