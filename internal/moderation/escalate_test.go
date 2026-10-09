package moderation

import (
	"sync"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
)

func TestEscalationFiresAtThreshold(t *testing.T) {
	e := setup(t)
	e.svc.window = 0
	_ = e.store.SetEscalationStep(t.Context(), guild, store.EscalationStep{WarnCount: 2, Action: "timeout", Duration: time.Hour})
	_ = e.store.SetEscalationStep(t.Context(), guild, store.EscalationStep{WarnCount: 3, Action: "kick"})
	if r := mustRun(t, e, act(cases.Warn)); r.Escalation != nil {
		t.Fatal("escalated at one warning")
	}
	r := mustRun(t, e, act(cases.Warn))
	if r.Escalation == nil || r.Escalation.Case.Kind != cases.Timeout || r.Escalation.Case.ModeratorID != bot {
		t.Fatalf("escalation %+v", r.Escalation)
	}
	if e.fake.TimeoutUntil(gs, us) == nil {
		t.Fatal("timeout not applied")
	}
	r = mustRun(t, e, act(cases.Warn))
	if r.Escalation == nil || r.Escalation.Case.Kind != cases.Kick || e.fake.IsMember(gs, us) {
		t.Fatal("kick step did not fire")
	}
}

func TestConcurrentWarnsEscalateOnce(t *testing.T) {
	e := setup(t)
	e.svc.window = 0
	_ = e.store.SetEscalationStep(t.Context(), guild, store.EscalationStep{WarnCount: 3, Action: "ban"})
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Go(func() {
			a := act(cases.Warn)
			a.Reason = "w" + string(rune('0'+i))
			if _, err := e.svc.Execute(t.Context(), a); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if e.fake.Calls("ban") != 1 {
		t.Fatalf("ban called %d times", e.fake.Calls("ban"))
	}
}
