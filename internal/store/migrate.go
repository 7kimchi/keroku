package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationLockKey serializes migrations across instances starting at once.
const migrationLockKey = 0x6b65726f6b75

// Migrate applies every unapplied *.sql file in version order, each in its own transaction.
// Changing a file after it was applied is an error.
func Migrate(ctx context.Context, pool *pgxpool.Pool, files fs.FS) error {
	pooled, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquire: %w", err)
	}
	// Take the connection out of the pool: its session timeouts are lifted for long migrations
	// and closing it releases the advisory lock even if unlock fails.
	conn := pooled.Hijack()
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	if _, err := conn.Exec(ctx, `SET statement_timeout = 0`); err != nil {
		return fmt.Errorf("migrate: session: %w", err)
	}
	if _, err := conn.Exec(ctx, `SET lock_timeout = 0`); err != nil {
		return fmt.Errorf("migrate: session: %w", err)
	}
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("migrate: lock: %w", err)
	}
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS "schemaMigrations" (
		"version" int PRIMARY KEY, "checksum" text NOT NULL, "appliedAt" timestamptz NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("migrate: bootstrap: %w", err)
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("migrate: list: %w", err)
	}
	sort.Strings(names)
	seen := map[int]string{}
	for _, name := range names {
		version, err := versionOf(name)
		if err != nil {
			return err
		}
		if prev, dup := seen[version]; dup {
			return fmt.Errorf("migrate: %s and %s share version %d", prev, name, version)
		}
		seen[version] = name
		body, err := fs.ReadFile(files, name)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}
		if err := apply(ctx, conn, version, name, body); err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, conn *pgx.Conn, version int, name string, body []byte) error {
	sum := sha256.Sum256(body)
	checksum := hex.EncodeToString(sum[:])
	var existing string
	err := conn.QueryRow(ctx, `SELECT "checksum" FROM "schemaMigrations" WHERE "version" = $1`, version).Scan(&existing)
	switch {
	case err == nil && existing == checksum:
		return nil
	case err == nil:
		return fmt.Errorf("migrate: %s changed after it was applied", name)
	case !errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("migrate: check %s: %w", name, err)
	}
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migrate: apply %s: %w", name, err)
		}
		_, err := tx.Exec(ctx, `INSERT INTO "schemaMigrations" ("version", "checksum") VALUES ($1, $2)`, version, checksum)
		return err
	})
}
