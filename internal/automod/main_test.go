package automod

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

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

const gs, bs, cs = "100000000000000001", "100000000000000003", "100000000000000010"

type modlog struct {
	mu   sync.Mutex
	sent []*discordgo.MessageEmbed
}

func (m *modlog) Case(_ int64, e *discordgo.MessageEmbed) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, e)
	return true
}

func (m *modlog) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sent)
}

type env struct {
	e     *Engine
	fake  *discord.Fake
	store *store.Store
	pool  *workers.Pool
	clk   *clock.Manual
	log   *modlog
	m     *metrics.Metrics
}

func setup(t *testing.T, c store.AutomodSettings) *env {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", perms.ViewChannel|perms.SendMessages,
		&discordgo.Role{ID: "staff", Position: 3, Permissions: perms.ManageMessages},
		&discordgo.Role{ID: "botrole", Position: 9, Permissions: perms.ModerateMembers | perms.ManageMessages})
	f.AddMember(gs, bs, "botrole")
	f.AddChannel(gs, cs)
	st := store.New(dbtest.New(t))
	if err := st.SetAutomod(context.Background(), 100000000000000001, c); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewManual(time.Now())
	f.SetNow(clk.Now)
	logger, m, ml := slog.New(slog.DiscardHandler), metrics.New(), &modlog{}
	cached, _ := store.NewCached(st.Automod, 100, time.Minute)
	pool, _ := workers.New("automod", 4, 100_000, safe.NewGuard(logger, nil))
	mod := moderation.New(moderation.Deps{Store: st, Client: f, Modlog: ml, Metrics: m, Log: logger, BotID: bs, Now: clk.Now})
	e, err := New(Deps{Settings: cached, Client: f, Moderation: mod, Modlog: ml, Pool: pool, Metrics: m, Log: logger,
		BotID: 100000000000000003, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close(context.Background()) })
	return &env{e: e, fake: f, store: st, pool: pool, clk: clk, log: ml, m: m}
}

func (v *env) drain() { _ = v.pool.Close(context.Background()) }

var msgID = 1300000000000600000

// post seeds a message in the fake and sends it through the engine.
func (v *env) post(author, content string, roles ...string) string {
	msgID++
	id := itoa(int64(msgID))
	v.fake.AddMember(gs, author, roles...)
	v.fake.AddMessage(cs, id, author, v.clk.Now())
	v.e.Message(&discordgo.Message{ID: id, GuildID: gs, ChannelID: cs, Content: content,
		Author: &discordgo.User{ID: author}, Member: &discordgo.Member{Roles: roles}})
	return id
}
