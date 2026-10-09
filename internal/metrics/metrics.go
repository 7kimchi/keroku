// Package metrics owns the Prometheus registry and every collector the bot exports.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Metrics groups the collectors. Names follow Prometheus conventions, not Go ones.
type Metrics struct {
	Registry       *prometheus.Registry
	Events         *prometheus.CounterVec
	EventsDropped  *prometheus.CounterVec
	Commands       *prometheus.CounterVec
	CommandSeconds *prometheus.HistogramVec
	Panics         *prometheus.CounterVec
	QueueDepth     *prometheus.GaugeVec
	DiscordErrors  *prometheus.CounterVec
	RateLimited    *prometheus.CounterVec
	Sweeper        *prometheus.CounterVec
	Automod        *prometheus.CounterVec
	Raids          prometheus.Counter
	Modlog         *prometheus.CounterVec
}

// New registers every collector on a fresh registry.
func New() *Metrics {
	r := prometheus.NewRegistry()
	r.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	f := factory{r}
	return &Metrics{
		Registry:      r,
		Events:        f.counter("keroku_events_total", "Gateway events received.", "type"),
		EventsDropped: f.counter("keroku_events_dropped_total", "Events dropped before handling.", "reason"),
		Commands:      f.counter("keroku_commands_total", "Commands handled.", "command", "result"),
		CommandSeconds: f.histogram("keroku_command_seconds", "Command handling time.",
			[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}, "command"),
		Panics:        f.counter("keroku_panics_total", "Recovered panics.", "where"),
		QueueDepth:    f.gauge("keroku_queue_depth", "Items waiting in a queue.", "queue"),
		DiscordErrors: f.counter("keroku_discord_errors_total", "Failed Discord API calls.", "kind"),
		RateLimited:   f.counter("keroku_rate_limited_total", "Commands refused by the rate limiter.", "scope"),
		Sweeper:       f.counter("keroku_sweeper_jobs_total", "Expired actions processed.", "kind", "result"),
		Automod:       f.counter("keroku_automod_hits_total", "Automod rule hits.", "rule"),
		Raids:         f.counter("keroku_raids_total", "Raid detections.").WithLabelValues(),
		Modlog:        f.counter("keroku_modlog_posts_total", "Modlog and log channel posts.", "result"),
	}
}

type factory struct{ r *prometheus.Registry }

func (f factory) counter(name, help string, labels ...string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, labels)
	f.r.MustRegister(c)
	return c
}

func (f factory) gauge(name, help string, labels ...string) *prometheus.GaugeVec {
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labels)
	f.r.MustRegister(g)
	return g
}

func (f factory) histogram(name, help string, buckets []float64, labels ...string) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: help, Buckets: buckets}, labels)
	f.r.MustRegister(h)
	return h
}
