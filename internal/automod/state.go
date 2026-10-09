package automod

import "time"

// Per user caps. Settings never ask for more than these, so the slices stay small no
// matter how fast a user posts.
const (
	maxSpamEntries = 50
	maxDupEntries  = 20
)

// userState is what automod remembers about one member in one guild.
type userState struct {
	times    []time.Time // recent message times
	dups     []dup       // recent content hashes
	noticeAt time.Time   // last modlog notice for this member
}

type dup struct {
	hash uint64
	at   time.Time
}

// keepTimes drops entries older than cutoff and keeps at most limit of the newest.
func keepTimes(ts []time.Time, cutoff time.Time, limit int) []time.Time {
	i := 0
	for i < len(ts) && ts[i].Before(cutoff) {
		i++
	}
	ts = ts[i:]
	if len(ts) > limit {
		ts = ts[len(ts)-limit:]
	}
	return append([]time.Time(nil), ts...)
}

func keepDups(ds []dup, cutoff time.Time, limit int) []dup {
	i := 0
	for i < len(ds) && ds[i].at.Before(cutoff) {
		i++
	}
	ds = ds[i:]
	if len(ds) > limit {
		ds = ds[len(ds)-limit:]
	}
	return append([]dup(nil), ds...)
}
