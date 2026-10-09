// Package store owns the Postgres pool, migrations and the settings and timer queries.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryTimeout bounds a single statement issued outside a caller supplied deadline.
const QueryTimeout = 5 * time.Second

// Querier is satisfied by both the pool and a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store wraps the pool.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps an open pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Pool exposes the pool for packages that run their own queries.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// Close waits for checked out connections and closes the pool.
func (s *Store) Close() { s.pool.Close() }

// Ping checks the database answers within QueryTimeout.
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	return s.pool.Ping(ctx)
}

// InTx runs fn in a transaction. fn's error rolls back, nil commits.
func (s *Store) InTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, s.pool, fn)
}

// Bounded returns ctx with QueryTimeout applied unless ctx already ends sooner.
func Bounded(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, QueryTimeout)
}

// IsUniqueViolation reports a unique constraint failure.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
