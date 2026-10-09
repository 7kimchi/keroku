package automod

import (
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

// spam fires when a member sends SpamMessages messages within SpamWindow.
func spam(st *userState, cfg store.AutomodSettings, now time.Time) bool {
	if !cfg.SpamEnabled {
		return false
	}
	st.times = keepTimes(append(st.times, now), now.Add(-cfg.SpamWindow), maxSpamEntries)
	return len(st.times) >= cfg.SpamMessages
}
