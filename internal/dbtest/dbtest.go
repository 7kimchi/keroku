// Package dbtest gives each test package its own migrated Postgres database.
package dbtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keroku/keroku/internal/store"
	"github.com/keroku/keroku/migrations"
)

// EnvURL names the variable holding an admin connection string for the test server.
const EnvURL = "KEROKU_TEST_DATABASE_URL"

// New creates a fresh database, migrates it and drops it when the test ends.
func New(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool := Empty(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := store.Migrate(ctx, pool, migrations.Files()); err != nil {
		t.Fatal(err)
	}
	return pool
}

// Empty creates a fresh database without running migrations.
func Empty(t testing.TB) *pgxpool.Pool {
	t.Helper()
	admin := os.Getenv(EnvURL)
	if admin == "" {
		t.Fatalf("%s is not set. Run scripts/testDb.sh start and export its output.", EnvURL)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	name := "kerokuTest" + hex.EncodeToString(buf)
	exec(ctx, t, admin, `CREATE DATABASE "`+name+`"`)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		exec(ctx, t, admin, `DROP DATABASE IF EXISTS "`+name+`" WITH (FORCE)`)
	})
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatalf("bad %s", EnvURL)
	}
	u.Path = "/" + name
	pool, err := store.Open(ctx, u.String(), 50)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// URL returns a connection string for a database created by New, for tests that need a second pool.
func URL(t testing.TB, pool *pgxpool.Pool) string {
	t.Helper()
	return pool.Config().ConnString()
}

func exec(ctx context.Context, t testing.TB, url, sql string) {
	t.Helper()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect test server: %v", err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}
