package moderation

import (
	"errors"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

func refusal(t *testing.T, e *env, a Action) string {
	t.Helper()
	_, err := e.svc.Execute(t.Context(), a)
	var u *commands.UserError
	if !errors.As(err, &u) {
		t.Fatalf("%s: want a refusal, got %v", a.Kind, err)
	}
	return u.Title + ". " + u.Detail
}

func TestRefusals(t *testing.T) {
	e := setup(t)
	edit := func(k cases.Kind, f func(*Action)) Action { a := act(k); f(&a); return a }
	e.fake.AddMember(gs, "100000000000000006", "admin")
	cases_ := map[string]Action{
		"Ban failed. Target is the server owner.":                edit(cases.Ban, func(a *Action) { a.TargetID = owner }),
		"Kick failed. Target is you.":                            edit(cases.Kick, func(a *Action) { a.TargetID = mod }),
		"Ban failed. Target is Keroku.":                          edit(cases.Ban, func(a *Action) { a.TargetID = bot }),
		"Ban failed. Missing permission: Ban Members.":           edit(cases.Ban, func(a *Action) { a.InvokerPerms = perms.KickMembers }),
		"Kick failed. Target's top role is at or above yours.":   edit(cases.Kick, func(a *Action) { a.TargetID = 100000000000000006 }),
		"Ban failed. Keroku is missing permission: Ban Members.": edit(cases.Ban, func(a *Action) { a.BotPerms = perms.KickMembers }),
		"Unban failed. Not banned.":                              act(cases.Unban),
		"Timeout removal failed. Not timed out.":                 act(cases.Untimeout),
		"Kick failed. Not a member of this server.":              edit(cases.Kick, func(a *Action) { a.TargetID = 100000000000000099 }),
	}
	for want, a := range cases_ {
		if got := refusal(t, e, a); got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
	if e.fake.Calls("ban") != 0 || e.fake.Calls("kick") != 0 || e.log.count() != 0 {
		t.Fatal("a refused action reached Discord or the modlog")
	}
}

func TestAlreadyBannedAndAdminTimeout(t *testing.T) {
	e := setup(t)
	mustRun(t, e, act(cases.Ban))
	if r := mustRun(t, e, act(cases.Ban)); !r.Duplicate || r.Case.Number != 1 {
		t.Fatalf("second ban within the window: %+v", r)
	}
	// A ban made outside Keroku has no case, so the state check catches it.
	_ = e.fake.Ban(t.Context(), gs, "100000000000000008", 0, "")
	b := act(cases.Ban)
	b.TargetID = 100000000000000008
	if got := refusal(t, e, b); got != "Ban failed. Already banned." {
		t.Fatalf("got %q", got)
	}
	e.fake.AddMember(gs, "100000000000000007", "admin")
	a := act(cases.Timeout)
	a.TargetID, a.Duration, a.ModeratorID = 100000000000000007, time.Hour, owner
	if got := refusal(t, e, a); got != "Timeout failed. Target has Administrator, which timeouts do not affect." {
		t.Fatalf("got %q", got)
	}
}
