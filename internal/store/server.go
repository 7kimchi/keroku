package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// minServer is the oldest Postgres Keroku runs on, as server_version_num.
const minServer = 180000

// checkServer refuses to run against Postgres older than 18.
func checkServer(ctx context.Context, pool *pgxpool.Pool) error {
	var n int
	if err := pool.QueryRow(ctx, "SELECT current_setting('server_version_num')::int").Scan(&n); err != nil {
		return fmt.Errorf("store: server version: %w", err)
	}
	return requireServer(n)
}

func requireServer(n int) error {
	if n < minServer {
		return fmt.Errorf("store: PostgreSQL 18 or newer required, server is %d.%d", n/10000, n%10000)
	}
	return nil
}
