package automod

import (
	"hash/maphash"
	"strings"
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

// duplicate fires when the same text, ignoring case and spacing, is sent DuplicateCount
// times within DuplicateWindow. Messages with no text never count.
func duplicate(st *userState, cfg store.AutomodSettings, content string, seed maphash.Seed, now time.Time) bool {
	if !cfg.DuplicateEnabled {
		return false
	}
	text := normalize(content)
	if text == "" {
		return false
	}
	h := maphash.String(seed, text)
	st.dups = keepDups(append(st.dups, dup{hash: h, at: now}), now.Add(-cfg.DuplicateWindow), maxDupEntries)
	same := 0
	for _, d := range st.dups {
		if d.hash == h {
			same++
		}
	}
	return same >= cfg.DuplicateCount
}

func normalize(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
