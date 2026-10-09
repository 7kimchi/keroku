package sweeper

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
)

var nextID int64 = 1300000000000500000

func tempBan(t *testing.T, e *env, d time.Duration) {
	t.Helper()
	nextID++
	_, err := e.mod.Execute(t.Context(), moderation.Action{Kind: cases.Ban, GuildID: guild, TargetID: user, ModeratorID: mod,
		Duration: d, InteractionID: nextID, InvokerRoles: []string{"mod"}, InvokerPerms: modPerms, BotPerms: modPerms})
	if err != nil {
		t.Fatal(err)
	}
}

func once(t *testing.T, e *env) bool {
	t.Helper()
	worked, err := e.sw.Once(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return worked
}

func TestTempBanEnds(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Hour)
	if once(t, e) {
		t.Fatal("processed before it was due")
	}
	e.clk.Advance(time.Hour)
	if !once(t, e) || e.fake.Banned(gs, us) || e.status(t, store.TimerUnban, user) != "done" {
		t.Fatal("ban not lifted")
	}
	c, err := cases.Get(t.Context(), e.store.Pool(), guild, 2)
	if err != nil || c.Kind != cases.Unban || c.ModeratorID != bot || c.IdempotencyKey == "" {
		t.Fatalf("unban case %+v %v", c, err)
	}
	if once(t, e) {
		t.Fatal("processed twice")
	}
}

func TestManualUnbanCancelsTimer(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Hour)
	nextID++
	_, err := e.mod.Execute(t.Context(), moderation.Action{Kind: cases.Unban, GuildID: guild, TargetID: user, ModeratorID: mod,
		InteractionID: nextID, InvokerRoles: []string{"mod"}, InvokerPerms: modPerms, BotPerms: modPerms})
	if err != nil {
		t.Fatal(err)
	}
	e.clk.Advance(2 * time.Hour)
	if once(t, e) || e.status(t, store.TimerUnban, user) != "cancelled" {
		t.Fatal("cancelled timer ran")
	}
}

func TestUnbannedOutsideKeroku(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Hour)
	_ = e.fake.Unban(t.Context(), gs, us, "")
	e.clk.Advance(time.Hour)
	if !once(t, e) || e.status(t, store.TimerUnban, user) != "done" {
		t.Fatal("timer not finished")
	}
	if _, err := cases.Get(t.Context(), e.store.Pool(), guild, 2); err == nil {
		t.Fatal("case made for an unban that did not happen")
	}
}
