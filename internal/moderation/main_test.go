package moderation

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

// Ids used across the tests. Guild 1, owner 2, bot 3, moderator 4, member 5.
const (
	guild = int64(100000000000000001)
	owner = int64(100000000000000002)
	bot   = int64(100000000000000003)
	mod   = int64(100000000000000004)
	user  = int64(100000000000000005)
)

const gs, os, bs, ms, us = "100000000000000001", "100000000000000002", "100000000000000003", "100000000000000004", "100000000000000005"

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
	svc   *Service
	fake  *discord.Fake
	store *store.Store
	log   *modlog
}

func setup(t *testing.T) *env {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, os, perms.ViewChannel,
		&discordgo.Role{ID: "member", Position: 1},
		&discordgo.Role{ID: "mod", Position: 5, Permissions: perms.BanMembers | perms.KickMembers | perms.ModerateMembers},
		&discordgo.Role{ID: "admin", Position: 8, Permissions: perms.Administrator},
		&discordgo.Role{ID: "botrole", Position: 10, Permissions: perms.BanMembers | perms.KickMembers | perms.ModerateMembers})
	f.AddMember(gs, bs, "botrole")
	f.AddMember(gs, ms, "mod")
	f.AddMember(gs, us, "member")
	f.AddMember(gs, os)
	st := store.New(dbtest.New(t))
	ml := &modlog{}
	svc := New(Deps{Store: st, Client: f, Modlog: ml, Metrics: metrics.New(), Log: slog.New(slog.DiscardHandler), BotID: bs})
	return &env{svc: svc, fake: f, store: st, log: ml}
}

var nextInteraction atomic.Int64

// act builds a command action from the moderator against the member.
func act(kind cases.Kind) Action {
	return Action{
		Kind: kind, GuildID: guild, TargetID: user, ModeratorID: mod, Reason: "spam",
		InteractionID: 1300000000000000000 + nextInteraction.Add(1),
		InvokerRoles:  []string{"mod"}, InvokerPerms: perms.BanMembers | perms.KickMembers | perms.ModerateMembers,
		BotPerms: perms.BanMembers | perms.KickMembers | perms.ModerateMembers,
	}
}
