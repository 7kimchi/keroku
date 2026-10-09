package store_test

import (
	"testing"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

// The test server is what the bot runs on, so Open must pass the version gate against it.
func TestOpenAcceptsTestServer(t *testing.T) {
	pool, err := store.Open(t.Context(), dbtest.URL(t, dbtest.Empty(t)), 4)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var n int
	if err := pool.QueryRow(t.Context(), "SELECT current_setting('server_version_num')::int").Scan(&n); err != nil || n < 180000 {
		t.Fatalf("server %d err %v", n, err)
	}
}
