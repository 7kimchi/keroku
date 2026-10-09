package cases

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInsertRejectsInvalid(t *testing.T) {
	pool := db(t)
	for _, n := range []New{
		{GuildID: 0, Kind: Ban, TargetID: 1, ModeratorID: 1},
		{GuildID: 1, Kind: "nuke", TargetID: 1, ModeratorID: 1},
		{GuildID: 1, Kind: Ban, TargetID: -1, ModeratorID: 1},
		{GuildID: 1, Kind: Ban, TargetID: 1, ModeratorID: 0},
		{GuildID: 1, Kind: Ban, TargetID: 1, ModeratorID: 1, Duration: -time.Second},
	} {
		if _, err := insert(t, pool, n); err == nil {
			t.Fatalf("%+v accepted", n)
		}
	}
}

func TestKindValid(t *testing.T) {
	for _, k := range []Kind{Ban, Unban, Kick, Timeout, Untimeout, Warn, Note} {
		if !k.Valid() {
			t.Fatal(k)
		}
	}
	if Kind("BAN").Valid() || Kind("").Valid() {
		t.Fatal("bad kind valid")
	}
}

func TestDatabaseOffline(t *testing.T) {
	pool := db(t)
	pool.Close()
	if _, err := insert(t, pool, newCase(1, 2, Ban)); err == nil {
		t.Fatal("insert on closed pool")
	}
	if _, err := Get(context.Background(), pool, 1, 1); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("offline read: %v", err)
	}
}
