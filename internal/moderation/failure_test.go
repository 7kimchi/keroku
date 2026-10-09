package moderation

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/discord"
)

// killTransactions ends every other backend that is waiting inside a transaction, which is
// where the moderation transaction sits while Discord is called.
func killTransactions(t *testing.T, e *env) func() {
	return func() {
		_, err := e.store.Pool().Exec(t.Context(), `SELECT pg_terminate_backend(pid) FROM pg_stat_activity
			WHERE datname = current_database() AND pid <> pg_backend_pid() AND state = 'idle in transaction'`)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestDiscordFailureLeavesNothing(t *testing.T) {
	e := setup(t)
	e.fake.FailNext("ban", &discord.Error{Op: "ban", Kind: discord.Unavailable, Status: 503}, 1)
	_, err := e.svc.Execute(t.Context(), act(cases.Ban))
	if err == nil || !strings.Contains(err.Error(), "Nothing was changed") {
		t.Fatalf("got %v", err)
	}
	e.fake.FailNext("kick", &discord.Error{Op: "kick", Kind: discord.Forbidden, Status: 403}, 1)
	if _, err := e.svc.Execute(t.Context(), act(cases.Kick)); err == nil || !strings.Contains(err.Error(), "Discord refused") {
		t.Fatalf("got %v", err)
	}
	if c, _ := insertProbe(t, e); c != 1 {
		t.Fatalf("failed actions consumed case numbers: next is %d", c)
	}
}

func TestCaseSaveFailureRevertsBan(t *testing.T) {
	e := setup(t)
	e.fake.OnCall("ban", killTransactions(t, e))
	_, err := e.svc.Execute(t.Context(), act(cases.Ban))
	if err == nil || !strings.Contains(err.Error(), "reverted") {
		t.Fatalf("got %v", err)
	}
	if e.fake.Banned(gs, us) || e.log.count() != 0 {
		t.Fatal("ban left in place without a case")
	}
}

func TestCaseSaveFailureAfterKickIsReported(t *testing.T) {
	e := setup(t)
	e.fake.OnCall("kick", killTransactions(t, e))
	_, err := e.svc.Execute(t.Context(), act(cases.Kick))
	if err == nil || !strings.Contains(err.Error(), "Kick incomplete") {
		t.Fatalf("got %v", err)
	}
}

func TestDatabaseDown(t *testing.T) {
	e := setup(t)
	e.store.Close()
	if _, err := e.svc.Execute(t.Context(), act(cases.Ban)); err == nil {
		t.Fatal("ran without a database")
	}
	if e.fake.Calls("ban") != 0 {
		t.Fatal("banned without being able to record it")
	}
}

// insertProbe stores a note and returns its number.
func insertProbe(t *testing.T, e *env) (int64, error) {
	r, err := e.svc.Execute(t.Context(), act(cases.Note))
	return r.Case.Number, err
}
