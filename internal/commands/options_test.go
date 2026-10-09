package commands

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	tString  = discordgo.ApplicationCommandOptionString
	tInt     = discordgo.ApplicationCommandOptionInteger
	tBool    = discordgo.ApplicationCommandOptionBoolean
	tUser    = discordgo.ApplicationCommandOptionUser
	tChannel = discordgo.ApplicationCommandOptionChannel
)

func req(t *testing.T, opts ...*discordgo.ApplicationCommandInteractionDataOption) *Request {
	t.Helper()
	r, err := Parse(interaction("x", opts...), "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func isUserErr(err error) bool {
	var u *UserError
	return errors.As(err, &u)
}

func TestTextOption(t *testing.T) {
	r := req(t, opt("reason", tString, "  spam\u0000 links  "), opt("long", tString, strings.Repeat("a", 513)), opt("num", tInt, 3.0))
	if s, ok, err := r.Text("reason", 512); !ok || err != nil || s != "spam links" {
		t.Fatalf("got %q %v %v", s, ok, err)
	}
	if _, _, err := r.Text("long", 512); !isUserErr(err) {
		t.Fatalf("long: %v", err)
	}
	if _, _, err := r.Text("num", 10); !isUserErr(err) {
		t.Fatal("wrong type accepted")
	}
	if _, ok, err := r.Text("missing", 10); ok || err != nil {
		t.Fatal("missing option")
	}
}

func TestIDOptions(t *testing.T) {
	r := req(t, opt("user", tUser, "100000000000000009"), opt("bad", tUser, "1e9"), opt("chan", tChannel, "100000000000000010"),
		opt("wrongValue", tUser, 5.0), opt("asString", tString, "100000000000000009"))
	if id, ok, err := r.User("user"); !ok || err != nil || id != "100000000000000009" {
		t.Fatal("user")
	}
	if id, ok, err := r.Channel("chan"); !ok || err != nil || id != "100000000000000010" {
		t.Fatal("channel")
	}
	for _, name := range []string{"bad", "wrongValue", "asString"} {
		if _, _, err := r.User(name); !isUserErr(err) {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestIntOption(t *testing.T) {
	r := req(t, opt("a", tInt, 100.0), opt("b", tInt, 101.0), opt("c", tInt, 1.5), opt("d", tInt, -1.0), opt("e", tInt, 1e300), opt("f", tInt, "5"))
	if n, ok, err := r.Int("a", 1, 100); !ok || err != nil || n != 100 {
		t.Fatal("boundary rejected")
	}
	for _, name := range []string{"b", "c", "d", "e", "f"} {
		if _, _, err := r.Int(name, 0, 100); !isUserErr(err) {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestBoolAndDurationOptions(t *testing.T) {
	r := req(t, opt("b", tBool, true), opt("nb", tBool, "true"), opt("d", tString, "1d4h"), opt("big", tString, "29d"), opt("junk", tString, "soon"))
	if b, ok, err := r.Bool("b"); !b || !ok || err != nil {
		t.Fatal("bool")
	}
	if _, _, err := r.Bool("nb"); !isUserErr(err) {
		t.Fatal("string bool accepted")
	}
	if d, ok, err := r.Duration("d", time.Second, 28*24*time.Hour); !ok || err != nil || d != 28*time.Hour {
		t.Fatalf("duration %v %v", d, err)
	}
	for _, name := range []string{"big", "junk"} {
		if _, _, err := r.Duration(name, time.Second, 28*24*time.Hour); !isUserErr(err) {
			t.Fatalf("%s accepted", name)
		}
	}
	if r.ResolvedMember("x") != nil {
		t.Fatal("no resolved data expected")
	}
}
