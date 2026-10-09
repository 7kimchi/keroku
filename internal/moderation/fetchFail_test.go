package moderation

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
)

// A panicking call drops this action, not the process, and never leaves a nil guild behind.
func TestFetchPanicBecomesError(t *testing.T) {
	for _, op := range []string{"guild", "member"} {
		e := setup(t)
		e.fake.OnCall(op, func() { panic("boom") })
		if _, err := e.svc.fetch(context.Background(), act(cases.Warn)); !errors.Is(err, errFetchPanic) {
			t.Fatalf("%s: %v", op, err)
		}
	}
}

func TestFetchBanPanicBecomesBanError(t *testing.T) {
	e := setup(t)
	e.fake.OnCall("getBan", func() { panic("boom") })
	f, err := e.svc.fetch(context.Background(), act(cases.Ban))
	if err != nil || !errors.Is(f.banErr, errFetchPanic) {
		t.Fatalf("err %v banErr %v", err, f.banErr)
	}
}

// A failed ban lookup does not hide a refusal: acting on the owner is refused either way.
func TestBanLookupFailureAfterRefusal(t *testing.T) {
	e := setup(t)
	e.fake.FailNext("getBan", errDown, 1)
	a := act(cases.Ban)
	a.TargetID = owner
	_, err := e.svc.inspect(context.Background(), a)
	var ue *commands.UserError
	if !errors.As(err, &ue) || ue.Cause != nil {
		t.Fatalf("want refusal, got %v", err)
	}
}

// Once checks pass, a failed ban lookup stops the ban and nothing is applied.
func TestBanLookupFailureStopsBan(t *testing.T) {
	e := setup(t)
	e.fake.FailNext("getBan", errDown, 1)
	if _, err := e.svc.Execute(context.Background(), act(cases.Ban)); err == nil {
		t.Fatal("ban went ahead without a ban lookup")
	}
	if e.fake.Banned(gs, us) || e.fake.Calls("ban") != 0 {
		t.Fatal("ban applied")
	}
}

func TestUnbanRefusedWhenNotBanned(t *testing.T) {
	e := setup(t)
	_, err := e.svc.inspect(context.Background(), act(cases.Unban))
	var ue *commands.UserError
	if !errors.As(err, &ue) || ue.Detail != NotBanned {
		t.Fatalf("got %v", err)
	}
}

// Hundreds of parallel fetches share the client without races or crossed results.
func TestFetchConcurrent(t *testing.T) {
	e := setup(t)
	var wg sync.WaitGroup
	errs := make(chan error, 400)
	for i := range 400 {
		wg.Go(func() {
			k := []cases.Kind{cases.Ban, cases.Warn, cases.Unban, cases.Kick}[i%4]
			f, err := e.svc.fetch(context.Background(), act(k))
			if err == nil && (f.member == nil || f.member.User.ID != us || f.bot.User.ID != bs) {
				err = errors.New("crossed result")
			}
			if err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}
