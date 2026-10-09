package commands

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

type lanedStub struct {
	stub
	key string
}

func (l lanedStub) Lane(*Request) string { return l.key }

func parsed(t *testing.T, opts ...*discordgo.ApplicationCommandInteractionDataOption) *Request {
	t.Helper()
	r, err := Parse(interaction("x", opts...), "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

const userA, userB = "100000000000000010", "100000000000000011"

func userOpt(v any) *discordgo.ApplicationCommandInteractionDataOption {
	return opt("user", discordgo.ApplicationCommandOptionUser, v)
}

func TestLaneKeyDefaultsToGuild(t *testing.T) {
	r := parsed(t, opt("reason", discordgo.ApplicationCommandOptionString, "spam"))
	if got := laneKey(okCmd("x"), r); got != r.GuildID {
		t.Fatalf("got %q", got)
	}
}

func TestLaneKeyPerTarget(t *testing.T) {
	a := laneKey(okCmd("x"), parsed(t, userOpt(userA)))
	b := laneKey(okCmd("x"), parsed(t, userOpt(userB)))
	again := laneKey(okCmd("y"), parsed(t, userOpt(userA)))
	if a == b || a != again {
		t.Fatalf("a %q b %q again %q", a, b, again)
	}
	if a != "100000000000000001/u"+userA {
		t.Fatalf("got %q", a)
	}
}

// A hostile option value must not pick an arbitrary lane key.
func TestLaneKeyRejectsBadUserOption(t *testing.T) {
	for _, v := range []any{"", "abc", "-1", "0", "1/../2", "100000000000000001/u1", 12, nil, "99999999999999999999999"} {
		r := parsed(t, userOpt(v))
		if got := laneKey(okCmd("x"), r); got != r.GuildID {
			t.Fatalf("%v gave %q", v, got)
		}
	}
}

func TestLaneKeyWrongOptionTypeIgnored(t *testing.T) {
	r := parsed(t, opt("user", discordgo.ApplicationCommandOptionString, userA))
	if got := laneKey(okCmd("x"), r); got != r.GuildID {
		t.Fatalf("got %q", got)
	}
}

func TestLaneKeyLanedOverrides(t *testing.T) {
	r := parsed(t, userOpt(userA))
	if got := laneKey(lanedStub{okCmd("x"), "c5"}, r); got != r.GuildID+"/c5" {
		t.Fatalf("got %q", got)
	}
}

// Keys from different guilds never collide, whatever the target.
func TestLaneKeyScopedToGuild(t *testing.T) {
	i := interaction("x", userOpt(userA))
	r1, _ := Parse(i, "r")
	i2 := interaction("x", userOpt(userA))
	i2.GuildID = "100000000000000099"
	r2, err := Parse(i2, "r")
	if err != nil {
		t.Fatal(err)
	}
	if laneKey(okCmd("x"), r1) == laneKey(okCmd("x"), r2) {
		t.Fatal("guilds share a lane key")
	}
	if laneKey(lanedStub{okCmd("x"), "c1"}, r1) == laneKey(lanedStub{okCmd("x"), "c1"}, r2) {
		t.Fatal("guilds share a laned key")
	}
}
