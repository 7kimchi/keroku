// Package ratelimit is an in memory GCRA limiter with a bounded number of keys.
package ratelimit

import (
	"errors"
	"time"

	"github.com/7kimchi/keroku/internal/cache"
)

// Limiter allows burst requests at once, then one every interval.
type Limiter struct {
	interval time.Duration
	burst    int
	now      func() time.Time
	tat      *cache.Cache[string, time.Time] // theoretical arrival time per key
}

// New builds a limiter. maxKeys bounds memory: the least recently used keys are dropped first.
func New(interval time.Duration, burst, maxKeys int, now func() time.Time) (*Limiter, error) {
	if interval <= 0 || burst < 1 || maxKeys < 1 {
		return nil, errors.New("ratelimit: interval, burst and maxKeys must be positive")
	}
	if now == nil {
		now = time.Now
	}
	// A key idle for a full burst window is back at full allowance, so it can go.
	ttl := interval * time.Duration(burst)
	c, err := cache.New[string, time.Time](maxKeys, ttl, now)
	if err != nil {
		return nil, err
	}
	return &Limiter{interval: interval, burst: burst, now: now, tat: c}, nil
}

// Allow spends one token for key. When refused it returns how long until the next token.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	now := l.now()
	tolerance := l.interval * time.Duration(l.burst-1)
	var allowed bool
	var wait time.Duration
	l.tat.Update(key, func(tat time.Time, found bool) time.Time {
		if !found || tat.Before(now) {
			tat = now
		}
		next := tat.Add(l.interval)
		if allowAt := next.Add(-l.interval - tolerance); now.Before(allowAt) {
			allowed, wait = false, allowAt.Sub(now)
			return tat
		}
		allowed = true
		return next
	})
	return allowed, wait
}

// Len reports how many keys are tracked.
func (l *Limiter) Len() int { return l.tat.Len() }

// Sweep drops idle keys, checking at most limit per shard.
func (l *Limiter) Sweep(limit int) int { return l.tat.Sweep(limit) }
