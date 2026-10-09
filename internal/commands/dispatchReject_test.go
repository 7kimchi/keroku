package commands

import (
	"strings"
	"testing"

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
