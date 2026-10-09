package store

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cache"
)

// Cached keeps per guild values loaded by load in a bounded cache.
type Cached[T any] struct {
	load  func(ctx context.Context, guildID int64) (T, error)
	cache *cache.Cache[int64, T]
}

// NewCached builds a cache of at most maxGuilds entries, each kept for ttl.
func NewCached[T any](load func(context.Context, int64) (T, error), maxGuilds int, ttl time.Duration) (*Cached[T], error) {
	c, err := cache.New[int64, T](maxGuilds, ttl, nil)
	if err != nil {
		return nil, err
	}
	return &Cached[T]{load: load, cache: c}, nil
}

// Get returns the cached value or loads it.
func (c *Cached[T]) Get(ctx context.Context, guildID int64) (T, error) {
	if v, ok := c.cache.Get(guildID); ok {
		return v, nil
	}
	v, err := c.load(ctx, guildID)
	if err != nil {
		var zero T
		return zero, err
	}
	c.cache.Set(guildID, v)
	return v, nil
}

// Invalidate drops a guild's cached value after a write.
func (c *Cached[T]) Invalidate(guildID int64) { c.cache.Delete(guildID) }

// Sweep drops expired entries.
func (c *Cached[T]) Sweep() { c.cache.Sweep(1000) }
