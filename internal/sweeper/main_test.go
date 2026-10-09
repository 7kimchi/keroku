package sweeper

import (
	"log/slog"
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
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

const (
	gs, bs, us = "100000000000000001", "100000000000000003", "100000000000000005"
	guild, bot = int64(100000000000000001), int64(100000000000000003)
	user, mod  = int64(100000000000000005), int64(100000000000000004)
	cs, chanID = "100000000000000010", int64(100000000000000010)
	modPerms   = perms.BanMembers | perms.KickMembers | perms.ModerateMembers
)

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

func (m *modlog) titles() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for _, e := range m.sent {
		out = append(out, e.Title)
	}
	return out
}

type env struct {
	sw    *Sweeper
	mod   *moderation.Service
	clean *cleanup.Service
	fake  *discord.Fake
	store *store.Store
	clk   *clock.Manual
	log   *modlog
}

func setup(t *testing.T) *env {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", perms.ViewChannel|perms.SendMessages,
		&discordgo.Role{ID: "mod", Position: 5, Permissions: modPerms},
		&discordgo.Role{ID: "botrole", Position: 10, Permissions: modPerms | perms.ManageRoles})
	f.AddMember(gs, bs, "botrole")
	f.AddMember(gs, "100000000000000004", "mod")
	f.AddMember(gs, us)
	f.AddChannel(gs, cs)
	clk := clock.NewManual(time.Now())
	f.SetNow(clk.Now)
	st := store.New(dbtest.New(t))
	ml := &modlog{}
	m := metrics.New()
	logger := slog.New(slog.DiscardHandler)
	mod := moderation.New(moderation.Deps{Store: st, Client: f, Modlog: ml, Metrics: m, Log: logger, BotID: bs, Now: clk.Now})
	clean := cleanup.New(st, f, ml, clk.Now)
	sw := New(Deps{Store: st, Client: f, Moderation: mod, Cleanup: clean, Modlog: ml, Metrics: m, Log: logger, BotID: bot, Now: clk.Now})
	return &env{sw: sw, mod: mod, clean: clean, fake: f, store: st, clk: clk, log: ml}
}

func (e *env) status(t *testing.T, kind string, target int64) string {
	t.Helper()
	var s string
	_ = e.store.Pool().QueryRow(t.Context(), `SELECT "status" FROM "tempActions" WHERE "kind" = $1 AND "targetId" = $2
		ORDER BY "id" DESC LIMIT 1`, kind, target).Scan(&s)
	return s
}
