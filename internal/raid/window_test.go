package raid

import (
	"strconv"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

func raidCfg() store.RaidSettings {
	return store.RaidSettings{Enabled: true, JoinLimit: 3, Window: 10 * time.Second, Action: "kick", LockdownFor: time.Hour}
}

func TestWindowDetectsAndContinues(t *testing.T) {
	w, c, now := &window{}, raidCfg(), time.Unix(1000, 0)
	if v := w.observe("a", now, c); v.detected || len(v.targets) != 0 {
		t.Fatal("one join is not a raid")
	}
	w.observe("b", now.Add(time.Second), c)
	v := w.observe("c", now.Add(2*time.Second), c)
	if !v.detected || len(v.targets) != 3 {
		t.Fatalf("raid not detected: %+v", v)
	}
	v = w.observe("d", now.Add(5*time.Second), c)
	if v.detected || len(v.targets) != 1 || v.targets[0] != "d" {
		t.Fatalf("join during raid: %+v", v)
	}
	if v := w.observe("e", now.Add(time.Minute), c); v.detected || len(v.targets) != 0 {
		t.Fatalf("raid did not end after a quiet window: %+v", v)
	}
}

func TestSlowJoinsAreNotARaid(t *testing.T) {
	w, c, now := &window{}, raidCfg(), time.Unix(1000, 0)
	for i := range 100 {
		if v := w.observe(strconv.Itoa(i), now.Add(time.Duration(i)*6*time.Second), c); v.detected {
			t.Fatalf("join %d flagged", i)
		}
	}
}

func TestWindowIsCapped(t *testing.T) {
	w, c, now := &window{}, raidCfg(), time.Unix(1000, 0)
	c.JoinLimit = 500
	for i := range 5000 {
		w.observe(strconv.Itoa(i), now, c)
	}
	if len(w.joins) > maxJoins {
		t.Fatalf("%d joins kept", len(w.joins))
	}
}
