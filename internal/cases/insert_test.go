package cases

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestNumbersArePerGuildAndSequential(t *testing.T) {
	pool := db(t)
	for i := int64(1); i <= 3; i++ {
		c, err := insert(t, pool, newCase(1, 2, Ban))
		if err != nil || c.Number != i || c.GuildID != 1 || c.ID == 0 || c.CreatedAt.IsZero() {
			t.Fatalf("guild 1 case %d: %+v %v", i, c, err)
		}
	}
	if c, _ := insert(t, pool, newCase(2, 2, Ban)); c.Number != 1 {
		t.Fatalf("guild 2 starts at %d", c.Number)
	}
}

func TestRollbackDoesNotConsumeNumber(t *testing.T) {
	pool := db(t)
	_, _ = insert(t, pool, newCase(1, 2, Warn))
	boom := errors.New("discord failed")
	err := pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		if _, err := Insert(t.Context(), tx, newCase(1, 2, Warn)); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if c, _ := insert(t, pool, newCase(1, 2, Warn)); c.Number != 2 {
		t.Fatalf("gap after rollback: got %d", c.Number)
	}
}

func TestFieldsRoundTrip(t *testing.T) {
	pool := db(t)
	n := New{GuildID: 5, Kind: Timeout, TargetID: 6, ModeratorID: 7, Reason: "'; DROP TABLE \"cases\"; --",
		Duration: 90 * time.Minute, InteractionID: 99}
	c, err := insert(t, pool, n)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Get(t.Context(), pool, 5, c.Number)
	if err != nil || got.Reason != n.Reason || got.Duration != n.Duration || got.InteractionID != 99 || got.Kind != Timeout {
		t.Fatalf("got %+v %v", got, err)
	}
	auto, _ := insert(t, pool, New{GuildID: 5, Kind: Kick, TargetID: 6, ModeratorID: 7, IdempotencyKey: "raid:1"})
	if auto.InteractionID != 0 || auto.IdempotencyKey != "raid:1" || auto.Duration != 0 || auto.Reason != "" {
		t.Fatalf("automatic case %+v", auto)
	}
}

func TestInteractionAndKeyAreUnique(t *testing.T) {
	pool := db(t)
	n := newCase(1, 2, Ban)
	n.InteractionID = 1234
	if _, err := insert(t, pool, n); err != nil {
		t.Fatal(err)
	}
	if _, err := insert(t, pool, n); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("replayed interaction: %v", err)
	}
	k := newCase(1, 2, Kick)
	k.IdempotencyKey = "escalation:9"
	_, _ = insert(t, pool, k)
	if _, err := insert(t, pool, k); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("repeated key: %v", err)
	}
	k.GuildID = 3
	if _, err := insert(t, pool, k); err != nil {
		t.Fatalf("same key in another guild: %v", err)
	}
	if c, _ := insert(t, pool, newCase(1, 2, Warn)); c.Number != 3 {
		t.Fatalf("duplicate attempts left a gap: %d", c.Number)
	}
}

// Real snowflakes do not fit in 32 bits. Every id column must take them.
func TestRealisticSnowflakes(t *testing.T) {
	pool := db(t)
	const big = int64(1300000000000000001)
	n := New{GuildID: big, Kind: Ban, TargetID: big + 1, ModeratorID: big + 2, InteractionID: big + 3, Duration: time.Hour}
	c, err := insert(t, pool, n)
	if err != nil || c.InteractionID != big+3 || c.GuildID != big {
		t.Fatalf("got %+v %v", c, err)
	}
	if _, err := editOnce(t, pool, big, c.Number, "x", big+4); err != nil {
		t.Fatal(err)
	}
	if ok, err := Claim(t.Context(), pool, big+5, big); !ok || err != nil {
		t.Fatal(err)
	}
}
