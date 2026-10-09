package info

import (
	"strconv"
	"strings"
	"testing"
)

func TestMentionListEmpty(t *testing.T) {
	for _, ids := range [][]string{nil, {}, {"", "abc", "-1", "0"}} {
		if got := mentionList("@&", ids, 1024); got != "None" {
			t.Fatalf("%v gave %q", ids, got)
		}
	}
}

func TestMentionListRenders(t *testing.T) {
	got := mentionList("@&", []string{rs, "bad", us}, 1024)
	if got != "<@&"+rs+">, <@&"+us+">" {
		t.Fatalf("got %q", got)
	}
}

func TestMentionListOverflow(t *testing.T) {
	ids := make([]string, 300)
	for i := range ids {
		ids[i] = strconv.FormatInt(100000000000000000+int64(i), 10)
	}
	got := mentionList("@&", ids, 1024)
	if len(got) > 1024 || !strings.HasSuffix(got, " more") {
		t.Fatalf("len %d: %q", len(got), got[len(got)-30:])
	}
	shown := strings.Count(got, "<@&")
	if !strings.Contains(got, " and "+strconv.Itoa(300-shown)+" more") || shown == 0 {
		t.Fatalf("shown %d: %q", shown, got[len(got)-30:])
	}
	// Order is kept: the shown ones are the first ones.
	if !strings.HasPrefix(got, "<@&"+ids[0]+">, <@&"+ids[1]+">") {
		t.Fatal("order lost")
	}
}

// A limit too small for even one mention still says how many there are.
func TestMentionListTinyLimit(t *testing.T) {
	if got := mentionList("@&", []string{rs, us}, 10); got != "2 more" {
		t.Fatalf("got %q", got)
	}
}

// A hostile payload with a million roles stays inside the limit.
func TestMentionListHuge(t *testing.T) {
	ids := make([]string, 1_000_000)
	for i := range ids {
		ids[i] = rs
	}
	if got := mentionList("@&", ids, 1024); len(got) > 1024 {
		t.Fatalf("len %d", len(got))
	}
}

func FuzzMentionList(f *testing.F) {
	f.Add(rs, "x", 1024)
	f.Add("<@everyone>", "@here", 30)
	f.Fuzz(func(t *testing.T, a, b string, limit int) {
		limit = 30 + abs(limit)%2000
		got := mentionList("@&", []string{a, b, rs}, limit)
		if len(got) > limit || strings.Contains(got, "everyone") || strings.Contains(got, "here") {
			t.Fatalf("limit %d got %q", limit, got)
		}
	})
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
