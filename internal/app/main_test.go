package app

import (
	"log/slog"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/config"
	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

func testConfig() config.Config {
	return config.Config{
		Token: config.NewSecret("fake.test.token"), Workers: 4, QueueSize: 64, DBMaxConns: 10,
		ShutdownTimeout: 5 * time.Second,
	}
}

func newTestApp(t *testing.T, cfg config.Config) (*App, *discord.Fake) {
	t.Helper()
	fake := discord.NewFake()
	fake.SetIdentity("100000000000000500", "100000000000000999")
	a, err := assemble(cfg, slog.New(slog.DiscardHandler), store.New(dbtest.New(t)), fake)
	if err != nil {
		t.Fatal(err)
	}
	return a, fake
}
