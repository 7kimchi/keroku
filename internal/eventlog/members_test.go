package eventlog

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestJoinsAndLeaves(t *testing.T) {
	l, p, pool := setup(t, 99)
	l.MemberAdd(&discordgo.GuildMemberAdd{Member: &discordgo.Member{GuildID: gs, User: &discordgo.User{ID: us}}})
	l.MemberRemove(&discordgo.GuildMemberRemove{Member: &discordgo.Member{GuildID: gs, User: &discordgo.User{ID: us}}})
	l.MemberAdd(&discordgo.GuildMemberAdd{Member: &discordgo.Member{GuildID: gs, User: &discordgo.User{ID: "bad"}}})
	_ = pool.Close(context.Background())
	if len(p.sent) != 2 || p.sent[0].Title != "Member joined" || p.sent[1].Title != "Member left" {
		t.Fatalf("posts %d", len(p.sent))
	}
	if !strings.HasPrefix(fields(p.sent[0])["Account created"], "<t:") {
		t.Fatal("account age missing")
	}
}

func TestBotsAndHugeTextAndMemory(t *testing.T) {
	l, _, pool := setup(t, 99)
	bot := msg("b", "x")
	bot.Author.Bot = true
	l.MessageCreate(bot)
	l.MessageCreate(msg("big", strings.Repeat("a", 4000)))
	for i := range 100_000 {
		l.MessageCreate(msg(string(rune('a'+i%26))+strings.Repeat("x", i%7)+itoa(i), "m"))
	}
	_ = pool.Close(context.Background())
	if _, ok := l.messages.Get("b"); ok {
		t.Fatal("bot message cached")
	}
	if l.messages.Len() > maxMessages {
		t.Fatalf("%d cached", l.messages.Len())
	}
	l.Sweep()
}
