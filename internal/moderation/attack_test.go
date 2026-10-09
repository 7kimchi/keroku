package moderation

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
)

// Roles change between commands. The check must use what Discord says now.
func TestHierarchyUsesFreshRoles(t *testing.T) {
	e := setup(t)
	e.fake.SetMemberRoles(gs, us, "admin")
	if got := refusal(t, e, act(cases.Kick)); got != "Kick failed. Target's top role is at or above yours." {
		t.Fatalf("got %q", got)
	}
	e.fake.SetMemberRoles(gs, us, "member")
	mustRun(t, e, act(cases.Kick))
}

func TestHostileReasonIsEscaped(t *testing.T) {
	e := setup(t)
	a := act(cases.Warn)
	a.Reason = "[click](https://evil.example) @everyone <@&1> **x**"
	r := mustRun(t, e, a)
	reason := Reply(r, bs).Fields[2].Value
	if strings.Contains(reason, "[click]") || strings.Contains(reason, "<@&1>") || strings.Contains(reason, "**x**") {
		t.Fatalf("unescaped: %s", reason)
	}
}

func TestInvalidActionsRejected(t *testing.T) {
	e := setup(t)
	bad := []func(*Action){
		func(a *Action) { a.Kind = "nuke" },
		func(a *Action) { a.GuildID = 0 },
		func(a *Action) { a.InteractionID = 0 },
		func(a *Action) { a.Reason = strings.Repeat("x", 513) },
		func(a *Action) { a.DeleteSeconds = MaxDeleteSeconds + 1 },
		func(a *Action) { a.Kind = cases.Kick; a.DeleteSeconds = 10 },
		func(a *Action) { a.Kind = cases.Timeout; a.Duration = 0 },
		func(a *Action) { a.Kind = cases.Timeout; a.Duration = MaxTimeout + time.Second },
		func(a *Action) { a.Duration = 30 * time.Second },
		func(a *Action) { a.Kind = cases.Warn; a.Duration = time.Hour },
	}
	for i, f := range bad {
		a := act(cases.Ban)
		f(&a)
		if _, err := e.svc.Execute(t.Context(), a); !errors.Is(err, errBadAction) {
			t.Errorf("case %d: got %v", i, err)
		}
	}
}
