package moderation

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
)

func mustRun(t *testing.T, e *env, a Action) Result {
	t.Helper()
	r, err := e.svc.Execute(t.Context(), a)
	if err != nil {
		t.Fatalf("%s: %v", a.Kind, err)
	}
	return r
}

func TestBanFlow(t *testing.T) {
	e := setup(t)
	a := act(cases.Ban)
	a.DeleteSeconds, a.Duration = 3600, 48*time.Hour
	r := mustRun(t, e, a)
	if !e.fake.Banned(gs, us) || r.Case.Number != 1 || r.Case.Kind != cases.Ban || r.DM != dmSent {
		t.Fatalf("result %+v", r)
	}
	if len(e.fake.SentTo("dm:"+us)) != 1 || e.log.count() != 1 {
		t.Fatal("dm or modlog missing")
	}
	end, ok, _ := store.PendingTimer(t.Context(), e.store.Pool(), guild, store.TimerUnban, user)
	if !ok || time.Until(end) < 47*time.Hour {
		t.Fatalf("unban timer %v %v", end, ok)
	}
	perm := act(cases.Ban)
	perm.TargetID = owner
	if _, err := e.svc.Execute(t.Context(), perm); err == nil {
		t.Fatal("owner banned")
	}
}

func TestEveryKindApplies(t *testing.T) {
	e := setup(t)
	mustRun(t, e, act(cases.Warn))
	mustRun(t, e, act(cases.Note))
	to := act(cases.Timeout)
	to.Duration = 2 * time.Hour
	mustRun(t, e, to)
	if until := e.fake.TimeoutUntil(gs, us); until == nil || time.Until(*until) < 119*time.Minute {
		t.Fatalf("timeout until %v", until)
	}
	mustRun(t, e, act(cases.Untimeout))
	if e.fake.TimeoutUntil(gs, us) != nil {
		t.Fatal("timeout not cleared")
	}
	mustRun(t, e, act(cases.Kick))
	if e.fake.IsMember(gs, us) {
		t.Fatal("not kicked")
	}
	mustRun(t, e, act(cases.Ban))
	r := mustRun(t, e, act(cases.Unban))
	if e.fake.Banned(gs, us) || r.Case.Number != 7 {
		t.Fatalf("unban case %d", r.Case.Number)
	}
	if n := len(e.fake.SentTo("dm:" + us)); n != 6 {
		t.Fatalf("%d DMs, notes must not DM", n)
	}
}

func TestLongTimeoutIsCappedAndRenewed(t *testing.T) {
	e := setup(t)
	a := act(cases.Timeout)
	a.Duration = 60 * 24 * time.Hour
	r := mustRun(t, e, a)
	until := e.fake.TimeoutUntil(gs, us)
	if until == nil || time.Until(*until) > DiscordTimeout || r.Case.Duration != a.Duration {
		t.Fatalf("until %v case %v", until, r.Case.Duration)
	}
	end, ok, _ := store.PendingTimer(t.Context(), e.store.Pool(), guild, store.TimerTimeoutRenew, user)
	if !ok || time.Until(end) < 59*24*time.Hour {
		t.Fatalf("renew timer %v %v", end, ok)
	}
	mustRun(t, e, act(cases.Untimeout))
	if _, ok, _ := store.PendingTimer(t.Context(), e.store.Pool(), guild, store.TimerTimeoutRenew, user); ok {
		t.Fatal("untimeout left the renew timer")
	}
}

func TestBlockedDMDoesNotBlockAction(t *testing.T) {
	e := setup(t)
	e.fake.BlockDMs(us)
	r := mustRun(t, e, act(cases.Kick))
	if r.DM != dmFailed || e.fake.IsMember(gs, us) || r.Case.Number != 1 {
		t.Fatalf("result %+v", r)
	}
}
