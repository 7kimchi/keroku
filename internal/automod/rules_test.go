package automod

import (
	"hash/maphash"
	"strings"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/store"
)

func cfg() store.AutomodSettings {
	c := store.DefaultAutomod()
	c.SpamEnabled, c.DuplicateEnabled, c.LinksEnabled = true, true, true
	c.AllowedDomains = []string{"example.com", "youtube.com"}
	return c
}

func TestSpamThresholdAndWindow(t *testing.T) {
	st, c, now := &userState{}, cfg(), time.Unix(1000, 0)
	for i := 1; i < c.SpamMessages; i++ {
		if spam(st, c, now.Add(time.Duration(i)*100*time.Millisecond)) {
			t.Fatalf("fired at message %d", i)
		}
	}
	if !spam(st, c, now.Add(time.Second)) {
		t.Fatal("did not fire at the limit")
	}
	if spam(st, c, now.Add(time.Minute)) {
		t.Fatal("window did not expire")
	}
	c.SpamEnabled = false
	for range 100 {
		if spam(&userState{}, c, now) {
			t.Fatal("disabled rule fired")
		}
	}
}

func TestSpamStateStaysBounded(t *testing.T) {
	st, c, now := &userState{}, cfg(), time.Unix(1000, 0)
	c.SpamMessages = 50
	for range 10_000 {
		spam(st, c, now)
	}
	if len(st.times) > maxSpamEntries {
		t.Fatalf("%d entries", len(st.times))
	}
}

func TestDuplicate(t *testing.T) {
	st, c, now, seed := &userState{}, cfg(), time.Unix(1000, 0), maphash.MakeSeed()
	msgs := []string{"Buy NOW", "buy   now", "something else", " BUY now "}
	got := []bool{}
	for i, m := range msgs {
		got = append(got, duplicate(st, c, m, seed, now.Add(time.Duration(i)*time.Second)))
	}
	if got[0] || got[1] || got[2] || !got[3] {
		t.Fatalf("got %v", got)
	}
	if duplicate(st, c, "", seed, now) || duplicate(st, c, "   ", seed, now) {
		t.Fatal("empty text counted")
	}
	if duplicate(st, c, "buy now", seed, now.Add(time.Hour)) {
		t.Fatal("old duplicates still counted")
	}
	for range 1000 {
		duplicate(st, c, strings.Repeat("x", 10), seed, now)
	}
	if len(st.dups) > maxDupEntries {
		t.Fatalf("%d dup entries", len(st.dups))
	}
}
