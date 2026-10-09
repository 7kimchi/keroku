package commands

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestSlowHandlerStillAnswers(t *testing.T) {
	h := newHarness(t, 5, stub{"slow", func(ctx context.Context, _ *Request) (*discordgo.MessageEmbed, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}})
	i := interaction("slow")
	h.d.Interaction(i)
	h.drain()
	if lastReply(t, h, i.ID).Title != "Command failed" {
		t.Fatal("timed out handler left the user without a reply")
	}
}

func TestThousandsOfConcurrentCommands(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 1000, counting("ban", &runs))
	var wg sync.WaitGroup
	for u := range 2000 {
		wg.Go(func() {
			i := interaction("ban")
			i.Member.User.ID = snowflakeAt(time.Now().Add(time.Duration(u) * time.Millisecond))
			h.d.Interaction(i)
		})
	}
	wg.Wait()
	h.drain()
	if runs.Load() != 2000 || h.fake.Calls("editResponse") != 2000 {
		t.Fatalf("ran %d, edits %d", runs.Load(), h.fake.Calls("editResponse"))
	}
}
