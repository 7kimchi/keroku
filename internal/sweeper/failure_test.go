package sweeper

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

func TestTransientFailureBacksOffThenFails(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Minute)
	e.clk.Advance(time.Minute)
	e.fake.FailNext("unban", &discord.Error{Op: "unban", Kind: discord.Unavailable, Status: 503}, 100)
	once(t, e)
	if e.status(t, store.TimerUnban, user) != "pending" || once(t, e) {
		t.Fatal("retried without backing off")
	}
	for range 12 {
		e.clk.Advance(2 * time.Hour)
		once(t, e)
	}
	if e.status(t, store.TimerUnban, user) != "failed" {
		t.Fatalf("status %s", e.status(t, store.TimerUnban, user))
	}
	titles := e.log.titles()
	if titles[len(titles)-1] != "Timer failed" {
		t.Fatalf("no failure notice: %v", titles)
	}
}

func TestRefusalFailsAtOnce(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Minute)
	e.clk.Advance(time.Minute)
	e.fake.FailNext("unban", &discord.Error{Op: "unban", Kind: discord.Forbidden, Status: 403}, 1)
	once(t, e)
	if e.status(t, store.TimerUnban, user) != "failed" {
		t.Fatal("permission failure retried")
	}
}
