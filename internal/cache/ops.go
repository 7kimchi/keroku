package cache

// Delete removes key. Reports whether it was present.
func (c *Cache[K, V]) Delete(key K) bool {
	i := c.index(key)
	c.locks[i].Lock()
	defer c.locks[i].Unlock()
	return c.shards[i].delete(key)
}

// Len counts stored entries, including expired ones not yet swept.
func (c *Cache[K, V]) Len() int {
	total := 0
	for i := range c.shards {
		c.locks[i].Lock()
		total += c.shards[i].order.Len()
		c.locks[i].Unlock()
	}
	return total
}

// Sweep checks up to limit entries per shard and drops the expired ones.
func (c *Cache[K, V]) Sweep(limit int) int {
	now := c.now()
	removed := 0
	for i := range c.shards {
		c.locks[i].Lock()
		removed += c.shards[i].sweep(now, limit)
		c.locks[i].Unlock()
	}
	return removed
}
