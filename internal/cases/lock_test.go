package cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestLockTargetSerializes(t *testing.T) {
	pool := db(t)
	held := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = pgx.BeginFunc(context.Background(), pool, func(tx pgx.Tx) error {
			_ = LockTarget(context.Background(), tx, 1, 2)
			close(held)
			<-release
			return nil
		})
	}()
	<-held
	start := time.Now()
	go func() { time.Sleep(200 * time.Millisecond); close(release) }()
	_ = pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error { return LockTarget(t.Context(), tx, 1, 2) })
	if time.Since(start) < 150*time.Millisecond {
		t.Fatal("second lock did not wait")
	}
	start = time.Now()
	_ = pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error { return LockTarget(t.Context(), tx, 1, 3) })
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("other target waited")
	}
}

func TestClaim(t *testing.T) {
	pool := db(t)
	if ok, err := Claim(t.Context(), pool, 10, 1); !ok || err != nil {
		t.Fatal("first claim")
	}
	if ok, _ := Claim(t.Context(), pool, 10, 1); ok {
		t.Fatal("second claim won")
	}
	_ = pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		_, _ = Claim(t.Context(), tx, 11, 1)
		return errors.New("rollback")
	})
	if ok, _ := Claim(t.Context(), pool, 11, 1); !ok {
		t.Fatal("rolled back claim still held")
	}
}
