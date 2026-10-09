package store_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/keroku/keroku/internal/dbtest"
	"github.com/keroku/keroku/internal/store"
)

func countCounters(t *testing.T, s *store.Store) int {
	t.Helper()
	var n int
	if err := s.Pool().QueryRow(t.Context(), `SELECT count(*) FROM "caseCounters"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestInTxCommitsAndRollsBack(t *testing.T) {
	s := store.New(dbtest.New(t))
	err := s.InTx(t.Context(), func(tx pgx.Tx) error {
		_, err := tx.Exec(t.Context(), `INSERT INTO "caseCounters" ("guildId") VALUES (1)`)
		return err
	})
	if err != nil || countCounters(t, s) != 1 {
		t.Fatalf("commit failed: %v", err)
	}
	boom := errors.New("boom")
	err = s.InTx(t.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), `INSERT INTO "caseCounters" ("guildId") VALUES (2)`); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) || countCounters(t, s) != 1 {
		t.Fatalf("rollback failed: %v", err)
	}
}

func TestInTxRollsBackOnPanic(t *testing.T) {
	s := store.New(dbtest.New(t))
	func() {
		defer func() { _ = recover() }()
		_ = s.InTx(t.Context(), func(tx pgx.Tx) error {
			_, _ = tx.Exec(t.Context(), `INSERT INTO "caseCounters" ("guildId") VALUES (3)`)
			panic("mid transaction")
		})
	}()
	if countCounters(t, s) != 0 {
		t.Fatal("panic left a committed row")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	s := store.New(dbtest.New(t))
	insert := func() error {
		_, err := s.Pool().Exec(t.Context(), `INSERT INTO "caseCounters" ("guildId") VALUES (9)`)
		return err
	}
	if err := insert(); err != nil {
		t.Fatal(err)
	}
	if err := insert(); !store.IsUniqueViolation(err) {
		t.Fatalf("got %v", err)
	}
	if store.IsUniqueViolation(errors.New("x")) || store.IsUniqueViolation(nil) {
		t.Fatal("false positive")
	}
}
