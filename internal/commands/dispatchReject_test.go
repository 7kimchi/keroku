package commands

import (
	"sync/atomic"

	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/perms"

	"github.com/bwmarrin/discordgo"
)

func TestDispatchRejectsBeforeDeferring(t *testing.T) {
	h := newHarness(t, 5, okCmd("ping"))
	dm := interaction("ping")
	dm.GuildID = ""
	unknown := interaction("nope")
	bad := interaction("ping")
	bad.ChannelID = "x"
	for _, i := range []*discordgo.Interaction{dm, unknown, bad} {
		h.d.Interaction(i)
	}
	h.drain()
	want := map[string]string{dm.ID: "Server only.", unknown.ID: "Unknown command."}
	for id, text := range want {
		if lastReply(t, h, id).Description != text {
			t.Fatalf("%s: %q", id, lastReply(t, h, id).Description)
		}
	}
	if !strings.HasPrefix(lastReply(t, h, bad.ID).Description, "Malformed request.") {
		t.Fatal("malformed not reported")
	}
}

func TestDispatchIgnoresOtherTypes(t *testing.T) {
	h := newHarness(t, 5, okCmd("ping"))
	i := interaction("ping")
	i.Type = discordgo.InteractionMessageComponent
	h.d.Interaction(i)
	h.d.Interaction(nil)
	h.drain()
	if h.fake.Calls("respond") != 0 {
		t.Fatal("responded to a non command")
	}
}

func TestDispatchChecksInvokerPermission(t *testing.T) {
	var runs atomic.Int64
	h := newHarness(t, 5, counting("ban", &runs))
	i := interaction("ban")
	i.Member.Permissions = 0
	h.d.Interaction(i)
	admin := interaction("ban")
	admin.Member.Permissions = perms.Administrator
	h.d.Interaction(admin)
	h.drain()
	if lastReply(t, h, i.ID).Description != "Missing permission: Ban Members." || runs.Load() != 1 {
		t.Fatalf("got %q, runs %d", lastReply(t, h, i.ID).Description, runs.Load())
	}
}
