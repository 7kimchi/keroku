package raid

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/clock"
	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

const gs, bs, guild = "100000000000000001", "100000000000000003", int64(100000000000000001)

type modlog struct {
	mu     sync.Mutex
	titles []string
}

func (m *modlog) Case(_ int64, e *discordgo.MessageEmbed) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.titles = append(m.titles, e.Title)
	return true
}

func (m *modlog) count(title string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, t := range m.titles {
		if t == title {
			n++
		}
	}
	return n
}

type env struct {
	x    *Detector
	fake *discord.Fake
	st   *store.Store
	pool *workers.Pool
	clk  *clock.Manual
	log  *modlog
}

func setup(t *testing.T, c store.RaidSettings) *env {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", perms.ViewChannel|perms.SendMessages,
		&discordgo.Role{ID: "botrole", Position: 9, Permissions: perms.KickMembers | perms.BanMembers | perms.ManageRoles})
	f.AddMember(gs, bs, "botrole")
	st := store.New(dbtest.New(t))
	if err := st.SetRaid(context.Background(), guild, c); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewManual(time.Now())
	f.SetNow(clk.Now)
	logger, m, ml := slog.New(slog.DiscardHandler), metrics.New(), &modlog{}
	cached, _ := store.NewCached(st.Raid, 100, time.Minute)
	pool, _ := workers.New("raid", 4, 100_000, safe.NewGuard(logger, nil))
	mod := moderation.New(moderation.Deps{Store: st, Client: f, Modlog: ml, Metrics: m, Log: logger, BotID: bs, Now: clk.Now})
	x, err := New(Deps{Settings: cached, Moderation: mod, Cleanup: cleanup.New(st, f, ml, clk.Now), Modlog: ml,
		Pool: pool, Metrics: m, Log: logger, BotID: 100000000000000003, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close(context.Background()) })
	return &env{x: x, fake: f, st: st, pool: pool, clk: clk, log: ml}
}

func (v *env) drain() { _ = v.pool.Close(context.Background()) }

// join adds a member whose account was created age ago and reports the join.
func (v *env) join(n int, age time.Duration) string {
	created := v.clk.Now().Add(-age).UnixMilli() - 1420070400000
	id := strconv.FormatInt(created<<22|int64(n), 10)
	v.fake.AddMember(gs, id)
	v.x.MemberAdd(&discordgo.GuildMemberAdd{Member: &discordgo.Member{GuildID: gs, User: &discordgo.User{ID: id},
		JoinedAt: v.clk.Now()}})
	return id
}
