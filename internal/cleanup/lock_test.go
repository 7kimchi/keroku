package cleanup

import (
	"strings"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

func overwrite(t *testing.T, f *discord.Fake) (int64, int64, bool) {
	t.Helper()
	ch, _ := f.Channel(t.Context(), cs)
	for _, o := range ch.PermissionOverwrites {
		if o.ID == gs {
			return o.Allow, o.Deny, true
		}
	}
	return 0, 0, false
}

func TestLockAndUnlockWithoutPriorOverwrite(t *testing.T) {
	s, f, st := setup(t)
	until := time.Now().Add(time.Hour)
	if err := s.Lock(t.Context(), guild, chan_, "r", until); err != nil {
		t.Fatal(err)
	}
	if _, deny, ok := overwrite(t, f); !ok || deny&perms.LockBits != perms.LockBits {
		t.Fatal("lock overwrite missing")
	}
	if _, ok, _ := store.PendingTimer(t.Context(), st.Pool(), guild, store.TimerChannelUnlock, chan_); !ok {
		t.Fatal("unlock timer missing")
	}
	if err := s.Lock(t.Context(), guild, chan_, "r", time.Time{}); err == nil || !strings.Contains(err.Error(), "already locked") {
		t.Fatalf("second lock: %v", err)
	}
	if found, err := s.Unlock(t.Context(), guild, chan_, "r"); !found || err != nil {
		t.Fatal(found, err)
	}
	if _, _, ok := overwrite(t, f); ok {
		t.Fatal("overwrite left behind where there was none")
	}
	if _, ok, _ := store.PendingTimer(t.Context(), st.Pool(), guild, store.TimerChannelUnlock, chan_); ok {
		t.Fatal("timer left after unlock")
	}
	if found, _ := s.Unlock(t.Context(), guild, chan_, "r"); found {
		t.Fatal("unlocked twice")
	}
}

func TestUnlockRestoresExactlyAndKeepsLaterEdits(t *testing.T) {
	s, f, _ := setup(t)
	_ = f.SetRoleOverwrite(t.Context(), cs, gs, perms.SendMessages|perms.EmbedLinks, perms.AddReactions|perms.ManageMessages, "")
	if err := s.Lock(t.Context(), guild, chan_, "r", time.Time{}); err != nil {
		t.Fatal(err)
	}
	allow, deny, _ := overwrite(t, f)
	if allow&perms.SendMessages != 0 || allow&perms.EmbedLinks == 0 || deny&perms.SendMessages == 0 {
		t.Fatalf("lock wrong: allow %b deny %b", allow, deny)
	}
	// An admin edits an unrelated bit while locked.
	_ = f.SetRoleOverwrite(t.Context(), cs, gs, allow|perms.ReadHistory, deny, "")
	_, _ = s.Unlock(t.Context(), guild, chan_, "r")
	allow, deny, _ = overwrite(t, f)
	wantAllow := perms.SendMessages | perms.EmbedLinks | perms.ReadHistory
	wantDeny := perms.AddReactions | perms.ManageMessages
	if allow != wantAllow || deny != wantDeny {
		t.Fatalf("restored allow %b deny %b, want %b %b", allow, deny, wantAllow, wantDeny)
	}
}

func TestLockFailureRollsBack(t *testing.T) {
	s, f, _ := setup(t)
	f.FailNext("setOverwrite", &discord.Error{Op: "setOverwrite", Kind: discord.Forbidden}, 1)
	if err := s.Lock(t.Context(), guild, chan_, "r", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("lock succeeded")
	}
	if err := s.Lock(t.Context(), guild, chan_, "r", time.Time{}); err != nil {
		t.Fatalf("failed lock left a record: %v", err)
	}
	f.AddChannel("100000000000000099", "100000000000000011")
	if err := s.Lock(t.Context(), guild, 100000000000000011, "r", time.Time{}); err == nil {
		t.Fatal("locked another guild's channel")
	}
}
