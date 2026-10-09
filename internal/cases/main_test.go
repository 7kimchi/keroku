package cases

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/dbtest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

func newCase(guild, target int64, kind Kind) New {
	return New{GuildID: guild, Kind: kind, TargetID: target, ModeratorID: 7, Reason: "r"}
}

// insert stores one case in its own transaction.
func insert(t testing.TB, pool *pgxpool.Pool, n New) (Case, error) {
	t.Helper()
	var c Case
	err := pgx.BeginFunc(context.Background(), pool, func(tx pgx.Tx) error {
		var err error
		c, err = Insert(context.Background(), tx, n)
		return err
	})
	return c, err
}

func db(t *testing.T) *pgxpool.Pool { return dbtest.New(t) }
