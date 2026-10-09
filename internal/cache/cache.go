// Package cache is a sharded, size bounded LRU with a TTL on every entry.
package cache

import (
	"errors"
	"hash/maphash"
	"sync"
	"time"
)

const maxShards = 16

// Cache holds at most maxSize entries. Old or expired entries are evicted first.
type Cache[K comparable, V any] struct {
	seed   maphash.Seed
	ttl    time.Duration
	now    func() time.Time
	locks  []sync.Mutex
	shards []*shard[K, V]
}

// New builds a cache. now may be nil to use the wall clock.
func New[K comparable, V any](maxSize int, ttl time.Duration, now func() time.Time) (*Cache[K, V], error) {
	if maxSize < 1 {
		return nil, errors.New("cache: maxSize must be at least 1")
	}
	if ttl <= 0 {
		return nil, errors.New("cache: ttl must be positive")
	}
	if now == nil {
		now = time.Now
	}
	n := min(maxShards, maxSize)
	c := &Cache[K, V]{seed: maphash.MakeSeed(), ttl: ttl, now: now,
		locks: make([]sync.Mutex, n), shards: make([]*shard[K, V], n)}
	// Spread maxSize exactly so the total never goes over it.
	for i := range c.shards {
		size := maxSize / n
		if i < maxSize%n {
			size++
		}
		c.shards[i] = newShard[K, V](size)
	}
	return c, nil
}

func (c *Cache[K, V]) index(key K) uint64 {
	return maphash.Comparable(c.seed, key) % uint64(len(c.shards))
}

// Get returns the live value for key.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	i := c.index(key)
	c.locks[i].Lock()
	defer c.locks[i].Unlock()
	if e, ok := c.shards[i].get(key, c.now()); ok {
		return e.value, true
	}
	var zero V
	return zero, false
}

// Set stores value under key with the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	i := c.index(key)
	c.locks[i].Lock()
	defer c.locks[i].Unlock()
	c.shards[i].set(key, value, c.now().Add(c.ttl))
}

// Add stores value only if key has no live entry. Reports whether it stored.
func (c *Cache[K, V]) Add(key K, value V) bool {
	i := c.index(key)
	c.locks[i].Lock()
	defer c.locks[i].Unlock()
	now := c.now()
	if _, ok := c.shards[i].get(key, now); ok {
		return false
	}
	c.shards[i].set(key, value, now.Add(c.ttl))
	return true
}

// Update replaces the value for key with fn(old, found) atomically and refreshes its TTL.
func (c *Cache[K, V]) Update(key K, fn func(old V, found bool) V) V {
	i := c.index(key)
	c.locks[i].Lock()
	defer c.locks[i].Unlock()
	now := c.now()
	var old V
	e, ok := c.shards[i].get(key, now)
	if ok {
		old = e.value
	}
	next := fn(old, ok)
	c.shards[i].set(key, next, now.Add(c.ttl))
	return next
}
