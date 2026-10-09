package moderation

import (
	"context"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/discord"
)

var errDown = &discord.Error{Op: "x", Kind: discord.Unavailable, Status: 503}

func TestFetchReadsEverything(t *testing.T) {
	e := setup(t)
	f, err := e.svc.fetch(context.Background(), act(cases.Ban))
	if err != nil {
		t.Fatal(err)
	}
	if f.guild == nil || f.guild.ID != gs || f.member == nil || f.member.User.ID != us || f.bot == nil || f.bot.User.ID != bs {
		t.Fatalf("%+v", f)
	}
	if f.banned || f.banErr != nil {
		t.Fatalf("banned %v err %v", f.banned, f.banErr)
	}
}

// The calls overlap. Run in order the ban path took four round trips.
func TestFetchIsParallel(t *testing.T) {
	e := setup(t)
	for _, op := range []string{"guild", "member", "getBan"} {
		e.fake.SetDelay(op, 100*time.Millisecond)
	}
	start := time.Now()
	if _, err := e.svc.fetch(context.Background(), act(cases.Ban)); err != nil {
		t.Fatal(err)
	}
	// Two members share a bucket and run in order, so 200ms is the floor.
	if d := time.Since(start); d < 200*time.Millisecond || d > 350*time.Millisecond {
		t.Fatalf("took %v", d)
	}
}

func TestFetchSkipsBanCheckForOtherActions(t *testing.T) {
	e := setup(t)
	for _, k := range []cases.Kind{cases.Kick, cases.Timeout, cases.Untimeout, cases.Warn, cases.Note} {
		if _, err := e.svc.fetch(context.Background(), act(k)); err != nil {
			t.Fatal(k, err)
		}
	}
	if n := e.fake.Calls("getBan"); n != 0 {
		t.Fatalf("getBan called %d times", n)
	}
}

func TestFetchTargetNotMember(t *testing.T) {
	e := setup(t)
	a := act(cases.Ban)
	a.TargetID = 100000000000000077
	f, err := e.svc.fetch(context.Background(), a)
	if err != nil || f.member != nil || f.bot == nil {
		t.Fatalf("err %v member %v bot %v", err, f.member, f.bot)
	}
}

func TestFetchFailures(t *testing.T) {
	for _, op := range []string{"guild", "member"} {
		e := setup(t)
		e.fake.FailNext(op, errDown, 1)
		if _, err := e.svc.fetch(context.Background(), act(cases.Warn)); !discord.Is(err, discord.Unavailable) {
			t.Fatalf("%s: %v", op, err)
		}
	}
}

// The bot is fetched after the target. Losing it fails the fetch.
func TestFetchBotFailure(t *testing.T) {
	e := setup(t)
	n := 0
	e.fake.OnCall("member", func() {
		if n++; n == 1 {
			e.fake.FailNext("member", errDown, 1)
		}
	})
	if _, err := e.svc.fetch(context.Background(), act(cases.Warn)); !discord.Is(err, discord.Unavailable) {
		t.Fatal(err)
	}
}
