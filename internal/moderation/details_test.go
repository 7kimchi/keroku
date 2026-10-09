package moderation

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
)

func TestCommandCasesRecordSource(t *testing.T) {
	e := setup(t)
	a := act(cases.Ban)
	a.DeleteSeconds = 86400
	r := mustRun(t, e, a)
	if r.Case.Details != (cases.Details{Source: cases.FromCommand, DeleteSeconds: 86400}) {
		t.Fatalf("%+v", r.Case.Details)
	}
	got, _ := cases.Get(t.Context(), e.store.Pool(), guild, r.Case.Number)
	if got.Details != r.Case.Details {
		t.Fatalf("stored %+v", got.Details)
	}
}

func TestWarnHasNoDeleteWindow(t *testing.T) {
	e := setup(t)
	if r := mustRun(t, e, act(cases.Warn)); r.Case.Details != (cases.Details{Source: cases.FromCommand}) {
		t.Fatalf("%+v", r.Case.Details)
	}
}

func TestEscalationRecordsWarnCase(t *testing.T) {
	e := setup(t)
	e.svc.window = 0
	_ = e.store.SetEscalationStep(t.Context(), guild, store.EscalationStep{WarnCount: 2, Action: "timeout", Duration: time.Hour})
	mustRun(t, e, act(cases.Warn))
	r := mustRun(t, e, act(cases.Warn))
	if r.Escalation == nil {
		t.Fatal("no escalation")
	}
	want := cases.Details{Source: cases.FromEscalation, WarnCase: r.Case.Number}
	if r.Escalation.Case.Details != want {
		t.Fatalf("%+v want %+v", r.Escalation.Case.Details, want)
	}
}

// A forged source or oversized rule is rejected before Discord or the database is touched.
func TestInvalidDetailsRejected(t *testing.T) {
	e := setup(t)
	for _, d := range []cases.Details{{Source: "owner"}, {Rule: string(make([]byte, 200))}, {WarnCase: -1}} {
		a := act(cases.Kick)
		a.Details = d
		if _, err := e.svc.Execute(t.Context(), a); err == nil {
			t.Fatalf("%+v accepted", d)
		}
	}
	if e.fake.Calls("kick") != 0 || e.fake.Calls("guild") != 0 {
		t.Fatal("reached Discord")
	}
}
