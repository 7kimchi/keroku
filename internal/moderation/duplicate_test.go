package moderation

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/7kimchi/keroku/internal/cases"
)

func TestReplayedInteraction(t *testing.T) {
	e := setup(t)
	a := act(cases.Warn)
	mustRun(t, e, a)
	r := mustRun(t, e, a)
	if !r.Duplicate || e.log.count() != 1 {
		t.Fatalf("replay produced %+v", r)
	}
}

func TestTwoModeratorsBanAtOnce(t *testing.T) {
	e := setup(t)
	var wg sync.WaitGroup
	var cases_, dups atomic.Int64
	for range 2 {
		wg.Go(func() {
			r, err := e.svc.Execute(t.Context(), act(cases.Ban))
			switch {
			case err != nil:
				t.Error(err)
			case r.Duplicate:
				dups.Add(1)
			default:
				cases_.Add(1)
			}
		})
	}
	wg.Wait()
	if cases_.Load() != 1 || dups.Load() != 1 || e.fake.Calls("ban") != 1 {
		t.Fatalf("cases %d dups %d ban calls %d", cases_.Load(), dups.Load(), e.fake.Calls("ban"))
	}
}

func TestDoubleSubmitStormMakesOneCase(t *testing.T) {
	e := setup(t)
	var wg sync.WaitGroup
	var made atomic.Int64
	for range 50 {
		wg.Go(func() {
			if r, err := e.svc.Execute(t.Context(), act(cases.Warn)); err == nil && !r.Duplicate {
				made.Add(1)
			}
		})
	}
	wg.Wait()
	n, _ := cases.CountByKind(t.Context(), e.store.Pool(), guild, user, cases.Warn)
	if made.Load() != 1 || n != 1 {
		t.Fatalf("made %d, stored %d", made.Load(), n)
	}
}

func TestAutomaticKeyRunsOnce(t *testing.T) {
	e := setup(t)
	a := act(cases.Kick)
	a.InteractionID, a.IdempotencyKey, a.Automated, a.ModeratorID = 0, "raid:1", true, bot
	first := mustRun(t, e, a)
	again := mustRun(t, e, a)
	if !again.Duplicate || again.Case.ID != first.Case.ID || e.fake.Calls("kick") != 1 {
		t.Fatalf("first %+v again %+v", first, again)
	}
}

func TestManyTargetsInParallel(t *testing.T) {
	e := setup(t)
	const n = 60
	for i := range n {
		e.fake.AddMember(gs, "1000000000000010"+twoDigits(i), "member")
	}
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			a := act(cases.Kick)
			a.TargetID = 100000000000001000 + int64(i)
			if _, err := e.svc.Execute(t.Context(), a); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	h, _ := cases.History(t.Context(), e.store.Pool(), guild, 100000000000001000, 10, 0)
	last, _ := cases.Get(t.Context(), e.store.Pool(), guild, n)
	if len(h) != 1 || last.Number != n || e.fake.Calls("kick") != n {
		t.Fatalf("history %d last %d kicks %d", len(h), last.Number, e.fake.Calls("kick"))
	}
}

func twoDigits(i int) string { return string(rune('0'+i/10)) + string(rune('0'+i%10)) }
