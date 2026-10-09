package cleanup

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/perms"
)

func TestLockdownRoundTrip(t *testing.T) {
	s, f, _ := setup(t)
	var ok atomic.Int64
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if s.Lockdown(t.Context(), guild, "r", time.Time{}) == nil {
				ok.Add(1)
			}
		})
	}
	wg.Wait()
	g, _ := f.Guild(t.Context(), gs)
	if ok.Load() != 1 || g.Roles[0].Permissions&perms.LockBits != 0 || g.Roles[0].Permissions&perms.ViewChannel == 0 {
		t.Fatalf("lockdowns %d perms %b", ok.Load(), g.Roles[0].Permissions)
	}
	if found, err := s.EndLockdown(t.Context(), guild, "r"); !found || err != nil {
		t.Fatal(found, err)
	}
	g, _ = f.Guild(t.Context(), gs)
	if g.Roles[0].Permissions != perms.ViewChannel|perms.SendMessages|perms.AddReactions|perms.EmbedLinks {
		t.Fatalf("restored %b", g.Roles[0].Permissions)
	}
	if found, _ := s.EndLockdown(t.Context(), guild, "r"); found {
		t.Fatal("ended twice")
	}
}
