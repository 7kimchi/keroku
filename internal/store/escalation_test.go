package store_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestEscalationSteps(t *testing.T) {
	s := store.New(dbtest.New(t))
	ctx := t.Context()
	steps := []store.EscalationStep{{3, "timeout", time.Hour}, {1, "timeout", 10 * time.Minute}, {5, "ban", 0}}
	for _, st := range steps {
		if err := s.SetEscalationStep(ctx, 1, st); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.EscalationSteps(ctx, 1)
	if err != nil || len(got) != 3 || got[0].WarnCount != 1 || got[2].Action != "ban" || got[1].Duration != time.Hour {
		t.Fatalf("got %+v %v", got, err)
	}
	_ = s.SetEscalationStep(ctx, 1, store.EscalationStep{WarnCount: 3, Action: "kick"})
	if got, _ = s.EscalationSteps(ctx, 1); len(got) != 3 || got[1].Action != "kick" || got[1].Duration != 0 {
		t.Fatalf("replace failed: %+v", got)
	}
	if ok, err := s.RemoveEscalationStep(ctx, 1, 3); !ok || err != nil {
		t.Fatal("remove")
	}
	if ok, _ := s.RemoveEscalationStep(ctx, 1, 3); ok {
		t.Fatal("removed twice")
	}
	if other, _ := s.EscalationSteps(ctx, 2); len(other) != 0 {
		t.Fatal("steps leaked across guilds")
	}
	for _, bad := range []store.EscalationStep{{0, "ban", 0}, {51, "ban", 0}, {2, "explode", 0}} {
		if err := s.SetEscalationStep(ctx, 1, bad); err == nil {
			t.Fatalf("%+v accepted", bad)
		}
	}
}

func TestEscalationCapUnderConcurrency(t *testing.T) {
	s := store.New(dbtest.New(t))
	var wg sync.WaitGroup
	var mu sync.Mutex
	tooMany := 0
	for i := 1; i <= 30; i++ {
		wg.Go(func() {
			err := s.SetEscalationStep(t.Context(), 1, store.EscalationStep{WarnCount: i, Action: "kick"})
			switch {
			case errors.Is(err, store.ErrTooManySteps):
				mu.Lock()
				tooMany++
				mu.Unlock()
			case err != nil:
				t.Error(err)
			}
		})
	}
	wg.Wait()
	got, _ := s.EscalationSteps(t.Context(), 1)
	if len(got) != store.MaxEscalationSteps || tooMany != 30-store.MaxEscalationSteps {
		t.Fatalf("%d stored, %d refused", len(got), tooMany)
	}
	if err := s.SetEscalationStep(t.Context(), 1, store.EscalationStep{WarnCount: got[0].WarnCount, Action: "ban"}); err != nil {
		t.Fatalf("replacing at the cap refused: %v", err)
	}
}
