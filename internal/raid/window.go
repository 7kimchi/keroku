package raid

import (
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

// maxJoins caps a window: settings allow at most 500 joins per window, so twice that is
// enough to act on every joiner during a raid.
const maxJoins = 1000

type join struct {
	userID string
	at     time.Time
}

// window is one guild's recent joins and raid state.
type window struct {
	joins     []join
	raidUntil time.Time
}

// verdict is what one join means.
type verdict struct {
	detected bool     // this join started a raid
	targets  []string // members to act on now
}

// observe records a join and decides what to do. A raid lasts while joins keep coming
// within the window of each other.
func (w *window) observe(userID string, now time.Time, cfg store.RaidSettings) verdict {
	cutoff := now.Add(-cfg.Window)
	kept := w.joins[:0]
	for _, j := range w.joins {
		if !j.at.Before(cutoff) {
			kept = append(kept, j)
		}
	}
	kept = append(kept, join{userID: userID, at: now})
	w.joins = kept
	if len(w.joins) > maxJoins {
		w.joins = append([]join(nil), w.joins[len(w.joins)-maxJoins:]...)
	}
	if now.Before(w.raidUntil) {
		w.raidUntil = now.Add(cfg.Window)
		return verdict{targets: []string{userID}}
	}
	if len(w.joins) < cfg.JoinLimit {
		return verdict{}
	}
	w.raidUntil = now.Add(cfg.Window)
	targets := make([]string, 0, len(w.joins))
	for _, j := range w.joins {
		targets = append(targets, j.userID)
	}
	return verdict{detected: true, targets: targets}
}
