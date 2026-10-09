package eventlog

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestEditAndDelete(t *testing.T) {
	l, p, pool := setup(t, 99)
	l.MessageCreate(msg("1", "hello **there**"))
	edited := time.Now()
	up := msg("1", "bye")
	up.EditedTimestamp = &edited
	l.MessageUpdate(&discordgo.MessageUpdate{Message: up})
	preview := msg("1", "bye")
	l.MessageUpdate(&discordgo.MessageUpdate{Message: preview})
	l.MessageDelete(&discordgo.MessageDelete{Message: &discordgo.Message{ID: "1", GuildID: gs, ChannelID: cs}})
	l.MessageDelete(&discordgo.MessageDelete{Message: &discordgo.Message{ID: "2", GuildID: gs, ChannelID: cs}})
	_ = pool.Close(context.Background())
	if len(p.sent) != 3 {
		t.Fatalf("%d posts", len(p.sent))
	}
	edit := fields(p.sent[0])
	if p.sent[0].Title != "Message edited" || edit["Before"] != `hello \*\*there\*\*` || edit["After"] != "bye" ||
		!strings.HasSuffix(edit["Message"], "/"+gs+"/"+cs+"/1") {
		t.Fatalf("edit %+v", edit)
	}
	del := fields(p.sent[1])
	if del["Content"] != "bye" || del["Author"] != "<@"+us+"> ("+us+")" {
		t.Fatalf("delete %+v", del)
	}
	if f := fields(p.sent[2]); f["Content"] != "Not cached." || f["Author"] != "Unknown" {
		t.Fatalf("uncached delete %+v", f)
	}
}

func TestNothingWithoutLogChannel(t *testing.T) {
	l, p, pool := setup(t, 0)
	l.MessageCreate(msg("1", "x"))
	l.MessageDelete(&discordgo.MessageDelete{Message: &discordgo.Message{ID: "1", GuildID: gs}})
	l.MemberAdd(&discordgo.GuildMemberAdd{Member: &discordgo.Member{GuildID: gs, User: &discordgo.User{ID: us}}})
	_ = pool.Close(context.Background())
	if len(p.sent) != 0 || l.messages.Len() != 0 {
		t.Fatal("logged or cached without a log channel")
	}
}
