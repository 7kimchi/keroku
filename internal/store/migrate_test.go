package store_test

import (
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/keroku/keroku/internal/dbtest"
	"github.com/keroku/keroku/internal/store"
	"github.com/keroku/keroku/migrations"
)

func sqlFS(files map[string]string) fstest.MapFS {
	m := fstest.MapFS{}
	for k, v := range files {
		m[k] = &fstest.MapFile{Data: []byte(v)}
	}
	return m
}

func TestMigrateIsIdempotent(t *testing.T) {
	pool := dbtest.New(t)
	for range 3 {
		if err := store.Migrate(t.Context(), pool, migrations.Files()); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM "schemaMigrations"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("%d migrations recorded", n)
	}
}

func TestMigrateConcurrentInstances(t *testing.T) {
	pool := dbtest.Empty(t)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Go(func() { errs <- store.Migrate(t.Context(), pool, migrations.Files()) })
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrateDetectsEditedFile(t *testing.T) {
	pool := dbtest.Empty(t)
	v1 := sqlFS(map[string]string{"0001A.sql": `CREATE TABLE "a" ("x" int)`})
	if err := store.Migrate(t.Context(), pool, v1); err != nil {
		t.Fatal(err)
	}
	v2 := sqlFS(map[string]string{"0001A.sql": `CREATE TABLE "a" ("x" bigint)`})
	if err := store.Migrate(t.Context(), pool, v2); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("got %v", err)
	}
}

func TestMigrateRejectsBadNames(t *testing.T) {
	for _, files := range []map[string]string{
		{"init.sql": "SELECT 1"},
		{"0000Zero.sql": "SELECT 1"},
		{"0001A.sql": "SELECT 1", "1B.sql": "SELECT 1"},
	} {
		if err := store.Migrate(t.Context(), dbtest.Empty(t), sqlFS(files)); err == nil {
			t.Fatalf("%v accepted", files)
		}
	}
}

func TestMigrateFailureRollsBack(t *testing.T) {
	pool := dbtest.Empty(t)
	bad := sqlFS(map[string]string{
		"0001Ok.sql":  `CREATE TABLE "ok" ("x" int)`,
		"0002Bad.sql": `CREATE TABLE "half" ("x" int); SELECT nonsense FROM nowhere;`,
	})
	if err := store.Migrate(t.Context(), pool, bad); err == nil {
		t.Fatal("bad migration applied")
	}
	var half, recorded int
	_ = pool.QueryRow(t.Context(), `SELECT count(*) FROM pg_tables WHERE tablename = 'half'`).Scan(&half)
	_ = pool.QueryRow(t.Context(), `SELECT count(*) FROM "schemaMigrations"`).Scan(&recorded)
	if half != 0 || recorded != 1 {
		t.Fatalf("half table %d, recorded %d", half, recorded)
	}
}
