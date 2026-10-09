package sweeper

import (
	"context"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

func TestLongTimeoutRenewsUntilItEnds(t *testing.T) {
	e := setup(t)
	nextID++
	_, err := e.mod.Execute(t.Context(), moderation.Action{Kind: cases.Timeout, GuildID: guild, TargetID: user, ModeratorID: mod,
		Duration: 70 * 24 * time.Hour, InteractionID: nextID, InvokerRoles: []string{"mod"}, InvokerPerms: modPerms, BotPerms: modPerms})
	if err != nil {
		t.Fatal(err)
	}
	end := e.clk.Now().Add(70 * 24 * time.Hour)
	e.clk.Advance(27*24*time.Hour + 22*time.Hour)
	if once(t, e) {
		t.Fatal("renewed early")
	}
	e.clk.Advance(time.Hour)
	if !once(t, e) || e.status(t, store.TimerTimeoutRenew, user) != "pending" {
		t.Fatal("first renewal did not run")
	}
	if until := e.fake.TimeoutUntil(gs, us); until == nil || until.Sub(e.clk.Now().Add(28*24*time.Hour)).Abs() > time.Second {
		t.Fatalf("renewed until %v", until)
	}
	e.clk.Advance(28 * 24 * time.Hour)
	if !once(t, e) || e.status(t, store.TimerTimeoutRenew, user) != "done" {
		t.Fatal("second renewal did not finish the timeout")
	}
	if until := e.fake.TimeoutUntil(gs, us); until == nil || until.Sub(end).Abs() > time.Second {
		t.Fatalf("final until %v, want %v", until, end)
	}
}

func TestTimedUnlockAndLockdown(t *testing.T) {
	e := setup(t)
	if err := e.clean.Lock(t.Context(), guild, chanID, "r", e.clk.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := e.clean.Lockdown(t.Context(), guild, "r", e.clk.Now().Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	e.clk.Advance(2 * time.Hour)
	once(t, e)
	once(t, e)
	ch, _ := e.fake.Channel(t.Context(), cs)
	g, _ := e.fake.Guild(t.Context(), gs)
	if len(ch.PermissionOverwrites) != 0 || g.Roles[0].Permissions&perms.SendMessages == 0 {
		t.Fatalf("not restored: %+v %b", ch.PermissionOverwrites, g.Roles[0].Permissions)
	}
	if e.status(t, store.TimerChannelUnlock, chanID) != "done" || e.status(t, store.TimerLockdownEnd, guild) != "done" {
		t.Fatal("timers not finished")
	}
}

func TestRunCatchesUpAfterRestartAndStops(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Minute)
	e.clk.Advance(time.Hour)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() { e.sw.Run(ctx); close(done) }()
	deadline := time.Now().Add(5 * time.Second)
	for e.fake.Banned(gs, us) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
	if e.fake.Banned(gs, us) {
		t.Fatal("overdue timer not processed on start")
	}
}

func TestCrashAfterCaseBeforeFinish(t *testing.T) {
	e := setup(t)
	tempBan(t, e, time.Minute)
	e.clk.Advance(time.Minute)
	var id int64
	_ = e.store.Pool().QueryRow(t.Context(), `SELECT "id" FROM "tempActions"`).Scan(&id)
	// The previous attempt unbanned and saved its case, then died before marking the timer.
	_, err := e.mod.Execute(t.Context(), moderation.Action{Kind: cases.Unban, GuildID: guild, TargetID: user, ModeratorID: bot,
		IdempotencyKey: "timer:" + itoa(id), Automated: true, FromTimer: true})
	if err != nil {
		t.Fatal(err)
	}
	if !once(t, e) || e.status(t, store.TimerUnban, user) != "done" {
		t.Fatal("not finished")
	}
	if n, _ := cases.CountByKind(t.Context(), e.store.Pool(), guild, user, cases.Unban); n != 1 {
		t.Fatalf("%d unban cases", n)
	}
}
