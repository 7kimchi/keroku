package app

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

type sink struct{ n int }

func (s *sink) Interaction(*discordgo.Interaction) { s.n++ }

func TestRouterFiltersEvents(t *testing.T) {
	var msgs, ups, dels, joins, leaves int
	s := &sink{}
	r := &router{
		interactions: s,
		messages:     []func(*discordgo.Message){func(*discordgo.Message) { msgs++ }},
		updates:      []func(*discordgo.MessageUpdate){func(*discordgo.MessageUpdate) { ups++ }},
		deletes:      []func(*discordgo.MessageDelete){func(*discordgo.MessageDelete) { dels++ }},
		joins:        []func(*discordgo.GuildMemberAdd){func(*discordgo.GuildMemberAdd) { joins++ }},
		leaves:       []func(*discordgo.GuildMemberRemove){func(*discordgo.GuildMemberRemove) { leaves++ }},
	}
	user := &discordgo.User{ID: "1"}
	r.Interaction(&discordgo.Interaction{})
	r.MessageCreate(nil)
	r.MessageCreate(&discordgo.Message{GuildID: "g"})
	r.MessageCreate(&discordgo.Message{Author: user})
	r.MessageCreate(&discordgo.Message{GuildID: "g", Author: &discordgo.User{ID: "2", Bot: true}})
	r.MessageCreate(&discordgo.Message{GuildID: "g", Author: user})
	r.MessageUpdate(&discordgo.MessageUpdate{})
	r.MessageUpdate(&discordgo.MessageUpdate{Message: &discordgo.Message{GuildID: "g"}})
	r.MessageDelete(&discordgo.MessageDelete{Message: &discordgo.Message{}})
	r.MessageDelete(&discordgo.MessageDelete{Message: &discordgo.Message{GuildID: "g"}})
	r.MemberAdd(&discordgo.GuildMemberAdd{})
	r.MemberAdd(&discordgo.GuildMemberAdd{Member: &discordgo.Member{User: user}})
	r.MemberRemove(&discordgo.GuildMemberRemove{Member: &discordgo.Member{}})
	r.MemberRemove(&discordgo.GuildMemberRemove{Member: &discordgo.Member{User: user}})
	if s.n != 1 || msgs != 1 || ups != 1 || dels != 1 || joins != 1 || leaves != 1 {
		t.Fatalf("counts %d %d %d %d %d %d", s.n, msgs, ups, dels, joins, leaves)
	}
}
