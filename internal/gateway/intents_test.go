package gateway

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestIntentsAreMinimal(t *testing.T) {
	base := Intents(false)
	want := discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildBans | discordgo.IntentsGuildMessages
	if base != want {
		t.Fatalf("got %b want %b", base, want)
	}
	if base&discordgo.IntentMessageContent != 0 {
		t.Fatal("message content requested without opt in")
	}
	if Intents(true) != want|discordgo.IntentMessageContent {
		t.Fatal("message content not added")
	}
	for _, unused := range []discordgo.Intent{discordgo.IntentsGuildPresences, discordgo.IntentsDirectMessages,
		discordgo.IntentsGuildVoiceStates, discordgo.IntentsGuildMessageReactions, discordgo.IntentsGuildMessageTyping} {
		if Intents(true)&unused != 0 {
			t.Fatalf("unused intent %b requested", unused)
		}
	}
}
