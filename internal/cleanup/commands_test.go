package cleanup

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/discord"
)

func TestSlowmodeCommand(t *testing.T) {
	s, f, _ := setup(t)
	e, err := SlowmodeCommand{s}.Handle(t.Context(), request(t, "slowmode", "", opt("interval", tStr, "5m")))
	if err != nil || e.Fields[1].Value != "5m" {
		t.Fatalf("%+v %v", e, err)
	}
	if ch, _ := f.Channel(t.Context(), cs); ch.RateLimitPerUser != 300 {
		t.Fatalf("slowmode %d", ch.RateLimitPerUser)
	}
	if e, _ = (SlowmodeCommand{s}).Handle(t.Context(), request(t, "slowmode", "", opt("interval", tStr, "OFF"))); e.Fields[1].Value != "Off" {
		t.Fatal("off")
	}
	for _, bad := range []string{"7h", "-1s", "soon"} {
		if _, err := (SlowmodeCommand{s}).Handle(t.Context(), request(t, "slowmode", "", opt("interval", tStr, bad))); err == nil {
			t.Fatalf("%s accepted", bad)
		}
	}
	f.FailNext("slowmode", &discord.Error{Op: "slowmode", Kind: discord.Forbidden}, 1)
	_, err = SlowmodeCommand{s}.Handle(t.Context(), request(t, "slowmode", "", opt("interval", tStr, "1s")))
	if err == nil || !strings.Contains(err.Error(), "Manage Channels") {
		t.Fatalf("got %v", err)
	}
}

func TestLockCommands(t *testing.T) {
	s, _, _ := setup(t)
	e, err := LockCommand{s}.Handle(t.Context(), request(t, "lock", "", opt("duration", tStr, "30m")))
	if err != nil || e.Title != "Channel locked" || len(e.Fields) != 2 {
		t.Fatalf("%+v %v", e, err)
	}
	if _, err := (LockCommand{s}).Handle(t.Context(), request(t, "lock", "")); err == nil {
		t.Fatal("double lock")
	}
	if e, err = (UnlockCommand{s}).Handle(t.Context(), request(t, "unlock", "")); err != nil || e.Title != "Channel unlocked" {
		t.Fatal(err)
	}
	if _, err := (UnlockCommand{s}).Handle(t.Context(), request(t, "unlock", "")); err == nil || !strings.Contains(err.Error(), "not locked") {
		t.Fatalf("got %v", err)
	}
	if _, err := (LockCommand{s}).Handle(t.Context(), request(t, "lock", "", opt("duration", tStr, "31d"))); err == nil {
		t.Fatal("lock over 30 days accepted")
	}
}

func TestLockdownCommand(t *testing.T) {
	s, _, _ := setup(t)
	if e, err := (LockdownCommand{s}).Handle(t.Context(), request(t, "lockdown", "start", opt("duration", tStr, "1h"))); err != nil || e.Title != "Lockdown started" {
		t.Fatal(err)
	}
	if e, err := (LockdownCommand{s}).Handle(t.Context(), request(t, "lockdown", "end")); err != nil || e.Title != "Lockdown ended" {
		t.Fatal(err)
	}
	if _, err := (LockdownCommand{s}).Handle(t.Context(), request(t, "lockdown", "end")); err == nil {
		t.Fatal("ended twice")
	}
}
