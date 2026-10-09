package moderation

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/bwmarrin/discordgo"
)

func TestCommandInputErrors(t *testing.T) {
	e := setup(t)
	cases_ := []struct {
		cmd  commands.Command
		req  *commands.Request
		want string
	}{
		{BanCommand{e.svc}, request(t, "ban"), "Pick a member."},
		{BanCommand{e.svc}, request(t, "ban", o("user", tUser, us), o("duration", tStr, "forever")), "duration must be"},
		{BanCommand{e.svc}, request(t, "ban", o("user", tUser, us), o("delete", tInt, 999999.0)), "delete must be"},
		{BanCommand{e.svc}, request(t, "ban", o("user", tUser, "not-an-id")), "Value for user"},
		{WarnCommand{e.svc}, request(t, "warn", o("user", tUser, us)), "A reason is required."},
		{WarnCommand{e.svc}, request(t, "warn", o("user", tUser, us), o("reason", tStr, " \U0000200B ")), "A reason is required."},
		{TimeoutCommand{e.svc}, request(t, "timeout", o("user", tUser, us)), "A duration is required."},
		{TimeoutCommand{e.svc}, request(t, "timeout", o("user", tUser, us), o("duration", tStr, "366d")), "at most"},
		{NoteCommand{e.svc}, request(t, "note", o("user", tUser, us), o("reason", tStr, strings.Repeat("x", 513))), "too long"},
	}
	for i, c := range cases_ {
		_, err := c.cmd.Handle(t.Context(), c.req)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("case %d: got %v, want %q", i, err, c.want)
		}
	}
	if e.fake.Calls("ban") != 0 {
		t.Fatal("bad input reached Discord")
	}
}

func TestSimpleCommands(t *testing.T) {
	e := setup(t)
	steps := []struct {
		cmd   commands.Command
		name  string
		title string
		extra []*discordgo.ApplicationCommandInteractionDataOption
	}{
		{WarnCommand{e.svc}, "warn", "Member warned", nil},
		{NoteCommand{e.svc}, "note", "Note added", nil},
		{TimeoutCommand{e.svc}, "timeout", "Member timed out", []*discordgo.ApplicationCommandInteractionDataOption{o("duration", tStr, "1h")}},
		{UntimeoutCommand{e.svc}, "untimeout", "Timeout removed", nil},
		{KickCommand{e.svc}, "kick", "Member kicked", nil},
	}
	for _, s := range steps {
		opts := append([]*discordgo.ApplicationCommandInteractionDataOption{o("user", tUser, us), o("reason", tStr, "x")}, s.extra...)
		reply, err := s.cmd.Handle(t.Context(), request(t, s.name, opts...))
		if err != nil || reply.Title != s.title {
			t.Fatalf("%s: %v %v", s.name, reply, err)
		}
	}
}
