package store_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestOpenRejectsBadURL(t *testing.T) {
	_, err := store.Open(t.Context(), "postgres://user:hunter2@[::1", 10)
	if err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("got %v", err)
	}
}

func TestOpenRejectsPoolSize(t *testing.T) {
	url := dbtest.URL(t, dbtest.Empty(t))
	for _, n := range []int{-1, 0, 1, 1001} {
		if _, err := store.Open(t.Context(), url, n); err == nil {
			t.Fatalf("maxConns %d accepted", n)
		}
	}
}

func TestOpenUnreachableFailsFast(t *testing.T) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	_, err := store.Open(ctx, "postgres://keroku@127.0.0.1:1/x?sslmode=disable&connect_timeout=2", 5)
	if err == nil {
		t.Fatal("connected to nothing")
	}
	if time.Since(start) > 12*time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
}

func TestSessionTimeoutsApplied(t *testing.T) {
	pool := dbtest.New(t)
	for setting, want := range map[string]string{
		"statement_timeout":                   "10s",
		"lock_timeout":                        "10s",
		"idle_in_transaction_session_timeout": "1min",
		"application_name":                    "keroku",
	} {
		var got string
		if err := pool.QueryRow(t.Context(), "SHOW "+setting).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s = %s, want %s", setting, got, want)
		}
	}
}

func TestStatementTimeoutKillsSlowQuery(t *testing.T) {
	pool := dbtest.New(t)
	start := time.Now()
	_, err := pool.Exec(t.Context(), "SELECT pg_sleep(30)")
	if err == nil || time.Since(start) > 20*time.Second {
		t.Fatalf("slow query not cancelled: %v after %v", err, time.Since(start))
	}
}

func TestPingAndClosedPool(t *testing.T) {
	s := store.New(dbtest.New(t))
	if err := s.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if err := s.Ping(t.Context()); err == nil {
		t.Fatal("ping on closed pool succeeded")
	}
}
