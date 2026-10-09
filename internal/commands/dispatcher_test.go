package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/7kimchi/keroku/internal/embeds"
)

func okCmd(name string) stub {
	return stub{name, func(context.Context, *Request) (*discordgo.MessageEmbed, error) {
		return embeds.New("Done").Build(), nil
	}}
}

func lastReply(t *testing.T, h *harness, id string) *discordgo.MessageEmbed {
	t.Helper()
	r := h.fake.Replies(id)
	if len(r) == 0 {
		t.Fatalf("no reply for %s", id)
	}
	return r[len(r)-1]
}

func TestDispatchDefersThenEdits(t *testing.T) {
	h := newHarness(t, 5, okCmd("ping"))
	i := interaction("ping")
	h.d.Interaction(i)
	h.drain()
	if !h.fake.Responded(i.ID) || lastReply(t, h, i.ID).Title != "Done" {
		t.Fatal("not deferred then edited")
	}
	if testutil.ToFloat64(h.m.Commands.WithLabelValues("ping", "ok")) != 1 {
		t.Fatal("metric not counted")
	}
}

func TestDispatchErrorKinds(t *testing.T) {
	h := newHarness(t, 50,
		stub{"refuse", func(context.Context, *Request) (*discordgo.MessageEmbed, error) {
			return nil, Fail("Ban failed", "Missing permission: Ban Members.")
		}},
		stub{"broken", func(context.Context, *Request) (*discordgo.MessageEmbed, error) {
			return nil, errors.New("pq: connection refused at 10.0.0.5")
		}},
		stub{"panics", func(context.Context, *Request) (*discordgo.MessageEmbed, error) { panic("nil map") }},
		stub{"empty", func(context.Context, *Request) (*discordgo.MessageEmbed, error) { return nil, nil }},
		stub{"partial", func(context.Context, *Request) (*discordgo.MessageEmbed, error) {
			return nil, FailWith("Kick failed", "Member kicked. Case not saved.", errors.New("tx aborted"))
		}},
	)
	ids := map[string]string{}
	for _, n := range []string{"refuse", "broken", "panics", "empty", "partial"} {
		i := interaction(n)
		ids[n] = i.ID
		h.d.Interaction(i)
	}
	h.drain()
	if e := lastReply(t, h, ids["refuse"]); e.Title != "Ban failed" || e.Description != "Missing permission: Ban Members." {
		t.Fatalf("refuse: %+v", e)
	}
	if e := lastReply(t, h, ids["partial"]); !strings.HasPrefix(e.Description, "Member kicked. Case not saved. Ref ") {
		t.Fatalf("partial: %q", e.Description)
	}
	for _, n := range []string{"broken", "panics", "empty"} {
		e := lastReply(t, h, ids[n])
		if e.Title != "Command failed" || !strings.Contains(e.Description, "Ref ") {
			t.Fatalf("%s: %+v", n, e)
		}
		if strings.Contains(e.Description, "10.0.0.5") || strings.Contains(e.Description, "nil map") {
			t.Fatalf("%s leaked internals: %s", n, e.Description)
		}
	}
}
