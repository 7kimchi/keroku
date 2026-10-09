package automod

import (
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

const spammer = "100000000000000007"

func TestSpamFloodIsDeleted(t *testing.T) {
	c := cfg()
	v := setup(t, c)
	for i := range 20 {
		v.post(spammer, "hello "+itoa(int64(i)))
	}
	v.drain()
	// The first SpamMessages-1 messages are allowed, the rest go.
	if left := v.fake.MessageCount(cs); left != c.SpamMessages-1 {
		t.Fatalf("%d messages left", left)
	}
	if testutil.ToFloat64(v.m.Automod.WithLabelValues("spam")) == 0 {
		t.Fatal("hits not counted")
	}
	if v.log.count() != 1 {
		t.Fatalf("%d notices during one flood, want 1", v.log.count())
	}
}

func TestStaffAndBotsAreExempt(t *testing.T) {
	v := setup(t, cfg())
	for range 20 {
		v.post("100000000000000008", "https://evil.com", "staff")
	}
	v.post("100000000000000002", "https://evil.com")
	v.drain()
	if v.fake.MessageCount(cs) != 21 || v.fake.Calls("deleteMessage") != 0 {
		t.Fatal("exempt member's messages removed")
	}
}

func TestDisabledGuildDoesNoWork(t *testing.T) {
	v := setup(t, defaultsOff())
	for range 50 {
		v.post(spammer, "https://evil.com")
	}
	v.drain()
	if v.fake.Calls("guild") != 0 || v.fake.Calls("deleteMessage") != 0 {
		t.Fatal("disabled automod made calls")
	}
}

func TestTimeoutCreatesOneCasePerHit(t *testing.T) {
	c := cfg()
	c.Timeout = 10 * time.Minute
	v := setup(t, c)
	v.post(spammer, "https://evil.com")
	v.drain()
	if until := v.fake.TimeoutUntil(gs, spammer); until == nil {
		t.Fatal("not timed out")
	}
	n, _ := cases.CountByKind(t.Context(), v.store.Pool(), 100000000000000001, 100000000000000007, cases.Timeout)
	if n != 1 || v.fake.Calls("deleteMessage") != 1 {
		t.Fatalf("%d timeout cases", n)
	}
}

func defaultsOff() store.AutomodSettings { return store.DefaultAutomod() }
