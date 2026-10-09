package commands

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

func counting(name string, n *atomic.Int64) stub {
	return stub{name, func(context.Context, *Request) (*discordgo.MessageEmbed, error) {
		n.Add(1)
		return okCmd(name).fn(context.Background(), nil)
	}}
}

func TestReplayedInteractionRunsOnce(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 50, counting("ban", &runs))
	i := interaction("ban")
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() { h.d.Interaction(i) })
	}
	wg.Wait()
	h.drain()
	if runs.Load() != 1 || h.fake.Calls("respond") != 1 {
		t.Fatalf("ran %d times, %d responses", runs.Load(), h.fake.Calls("respond"))
	}
}

func TestUserRateLimit(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 3, counting("ban", &runs))
	var ids []string
	for range 10 {
		i := interaction("ban")
		ids = append(ids, i.ID)
		h.d.Interaction(i)
	}
	h.drain()
	if runs.Load() != 3 {
		t.Fatalf("ran %d, want 3", runs.Load())
	}
	limited := 0
	for _, id := range ids {
		if lastReply(t, h, id).Title == "Rate limited" {
			limited++
		}
	}
	if limited != 7 {
		t.Fatalf("%d limited replies", limited)
	}
}

func TestExpiredTokenIsDropped(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 5, counting("ban", &runs))
	i := interaction("ban")
	i.ID = snowflakeAt(time.Now().Add(-20 * time.Minute))
	h.d.Interaction(i)
	h.drain()
	if runs.Load() != 0 {
		t.Fatal("ran a command whose token expired")
	}
}

func TestDeferFailureStopsWork(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 5, counting("ban", &runs))
	h.fake.FailNext("respond", &discord.Error{Op: "respond", Kind: discord.Expired}, 1)
	h.d.Interaction(interaction("ban"))
	h.drain()
	if runs.Load() != 0 {
		t.Fatal("ran after the interaction could not be acknowledged")
	}
}
