package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open connects with bounded pool size and server side timeouts so a stuck query or an
// abandoned transaction cannot hold connections forever.
func Open(ctx context.Context, url string, maxConns int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		// pgx errors can echo parts of the URL, so they stay out of the message.
		return nil, errors.New("store: DATABASE_URL is not a valid connection string")
	}
	if maxConns < 2 || maxConns > 1000 {
		return nil, fmt.Errorf("store: maxConns %d out of range", maxConns)
	}
	cfg.MaxConns = int32(maxConns) // #nosec G115 -- bounded above
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	params := cfg.ConnConfig.RuntimeParams
	params["application_name"] = "keroku"
	params["statement_timeout"] = "10000"
	params["lock_timeout"] = "10000"
	params["idle_in_transaction_session_timeout"] = "60000"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("store: open pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return pool, nil
}
