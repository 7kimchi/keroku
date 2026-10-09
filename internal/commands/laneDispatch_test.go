package commands

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/embeds"
)

// tracker records how many handlers ran at once and what each target was handled with.
type tracker struct {
	running, peak atomic.Int64
	mu            sync.Mutex
	order         map[string][]string
}

func (tr *tracker) cmd(name string) stub {
	return stub{name, func(_ context.Context, r *Request) (*discordgo.MessageEmbed, error) {
		n := tr.running.Add(1)
		for p := tr.peak.Load(); n > p && !tr.peak.CompareAndSwap(p, n); p = tr.peak.Load() {
		}
		target, _, _ := r.User("user")
		seq, _, _ := r.Text("reason", 10)
		tr.mu.Lock()
		tr.order[target] = append(tr.order[target], seq)
		tr.mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		tr.running.Add(-1)
		return embeds.New("Done").Build(), nil
	}}
}

func send(h *harness, name, user string, seq int) {
	h.d.Interaction(interaction(name, userOpt(user),
		opt("reason", discordgo.ApplicationCommandOptionString, strconv.Itoa(seq))))
}

// Sixteen targets in one guild must not all queue behind each other.
func TestDispatchRunsDifferentTargetsInParallel(t *testing.T) {
	tr := &tracker{order: map[string][]string{}}
	h := newHarness(t, 1000, tr.cmd("act"))
	for i := range 16 {
		send(h, "act", "1000000000000001"+strconv.Itoa(10+i), 0)
	}
	h.drain()
	if tr.peak.Load() < 2 {
		t.Fatal("targets in one guild ran one at a time")
	}
}

// Commands on one target run one at a time. Acks run in parallel, so arrival order is not kept.
func TestDispatchOrdersOneTarget(t *testing.T) {
	tr := &tracker{order: map[string][]string{}}
	h := newHarness(t, 1000, tr.cmd("act"))
	for i := range 30 {
		send(h, "act", userA, i)
	}
	h.drain()
	if tr.peak.Load() != 1 {
		t.Fatalf("peak %d on one target", tr.peak.Load())
	}
	seen := map[string]bool{}
	for _, s := range tr.order[userA] {
		seen[s] = true
	}
	if len(tr.order[userA]) != 30 || len(seen) != 30 {
		t.Fatalf("handled %v", tr.order[userA])
	}
}

// Commands without a target still share the guild lane.
func TestDispatchOrdersGuildWideCommands(t *testing.T) {
	tr := &tracker{order: map[string][]string{}}
	h := newHarness(t, 1000, tr.cmd("lockdown"))
	for i := range 10 {
		h.d.Interaction(interaction("lockdown", opt("reason", discordgo.ApplicationCommandOptionString, strconv.Itoa(i))))
	}
	h.drain()
	if tr.peak.Load() != 1 || len(tr.order[""]) != 10 {
		t.Fatalf("peak %d handled %d", tr.peak.Load(), len(tr.order[""]))
	}
}
