package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewRegistersEverything(t *testing.T) {
	m := New()
	m.Events.WithLabelValues("MESSAGE_CREATE").Inc()
	m.EventsDropped.WithLabelValues("queueFull").Inc()
	m.Commands.WithLabelValues("ban", "ok").Inc()
	m.CommandSeconds.WithLabelValues("ban").Observe(0.02)
	m.Panics.WithLabelValues("handler").Inc()
	m.QueueDepth.WithLabelValues("lanes").Set(3)
	m.DiscordErrors.WithLabelValues("429").Inc()
	m.RateLimited.WithLabelValues("user").Inc()
	m.Sweeper.WithLabelValues("unban", "ok").Inc()
	m.Automod.WithLabelValues("spam").Inc()
	m.Raids.Inc()
	m.Modlog.WithLabelValues("ok").Inc()
	n, err := testutil.GatherAndCount(m.Registry)
	if err != nil {
		t.Fatal(err)
	}
	if n < 12 {
		t.Fatalf("only %d series", n)
	}
	if got := testutil.ToFloat64(m.Commands.WithLabelValues("ban", "ok")); got != 1 {
		t.Fatalf("got %v", got)
	}
}

func TestIndependentRegistries(t *testing.T) {
	a, b := New(), New()
	a.Raids.Inc()
	if testutil.ToFloat64(b.Raids) != 0 {
		t.Fatal("registries share state")
	}
}

func TestNamesLint(t *testing.T) {
	problems, err := testutil.GatherAndLint(New().Registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		if strings.HasPrefix(p.Metric, "keroku_") {
			t.Errorf("%s: %s", p.Metric, p.Text)
		}
	}
}
